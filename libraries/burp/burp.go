package burp

import (
	"fmt"
	"html"
	"reflect"
	"strconv"
	"strings"
)

type RawHTML string

func Raw(text string) RawHTML {
	return RawHTML(text)
}

func Escape(text string) string {
	return html.EscapeString(text)
}

func Html(nodes ...any) string {
	var out strings.Builder
	renderNodes(&out, nodes)
	return out.String()
}

func Html5(nodes ...any) string {
	return "<!DOCTYPE html>" + Html(nodes...)
}

var voidTags = map[string]struct{}{
	"area": {}, "base": {}, "br": {}, "col": {}, "embed": {}, "hr": {},
	"img": {}, "input": {}, "link": {}, "meta": {}, "param": {}, "source": {},
	"track": {}, "wbr": {},
}

func renderNodes(out *strings.Builder, nodes []any) {
	for _, node := range nodes {
		renderNode(out, node)
	}
}

func renderNode(out *strings.Builder, node any) {
	switch value := node.(type) {
	case nil:
		return
	case RawHTML:
		out.WriteString(string(value))
	case string:
		out.WriteString(html.EscapeString(value))
	case []any:
		renderNodeSlice(out, value)
	default:
		renderReflectValue(out, reflect.ValueOf(node))
	}
}

func renderReflectValue(out *strings.Builder, value reflect.Value) {
	if !value.IsValid() {
		return
	}

	switch value.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]any, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			items = append(items, value.Index(i).Interface())
		}
		renderNodeSlice(out, items)
	case reflect.Map:
		out.WriteString(html.EscapeString(fmt.Sprint(value.Interface())))
	case reflect.String:
		out.WriteString(html.EscapeString(value.String()))
	case reflect.Bool:
		out.WriteString(strconv.FormatBool(value.Bool()))
	default:
		out.WriteString(html.EscapeString(fmt.Sprint(value.Interface())))
	}
}

func renderNodeSlice(out *strings.Builder, items []any) {
	if len(items) == 0 {
		return
	}
	if looksLikeElement(items) {
		renderElement(out, items)
		return
	}
	for _, item := range items {
		renderNode(out, item)
	}
}

func looksLikeElement(items []any) bool {
	if len(items) == 0 {
		return false
	}
	tag, ok := tagToken(items[0])
	if !ok {
		return false
	}
	return strings.HasPrefix(tag, ":") || strings.ContainsAny(tag, "#.") || len(items) > 1
}

func renderElement(out *strings.Builder, items []any) {
	tagSpec, ok := parseTag(items[0])
	if !ok {
		renderNodeSlice(out, items)
		return
	}

	attrIndex := 1
	if len(items) > 1 && isAttrMap(items[1]) {
		attrIndex = 2
	}

	out.WriteByte('<')
	out.WriteString(tagSpec.name)
	if tagSpec.id != "" {
		writeAttr(out, "id", tagSpec.id)
	}
	if len(tagSpec.classes) > 0 {
		writeAttr(out, "class", strings.Join(tagSpec.classes, " "))
	}
	if attrIndex > 1 {
		renderAttrs(out, items[1])
	}

	if _, ok := voidTags[tagSpec.name]; ok {
		out.WriteByte('>')
		return
	}

	out.WriteByte('>')
	renderNodes(out, items[attrIndex:])
	out.WriteString("</")
	out.WriteString(tagSpec.name)
	out.WriteByte('>')
}

func renderAttrs(out *strings.Builder, attrs any) {
	switch value := attrs.(type) {
	case map[any]any:
		for key, val := range value {
			renderAttr(out, key, val)
		}
	case map[string]any:
		for key, val := range value {
			renderAttr(out, key, val)
		}
	default:
		rv := reflect.ValueOf(attrs)
		if rv.IsValid() && rv.Kind() == reflect.Map {
			iter := rv.MapRange()
			for iter.Next() {
				renderAttr(out, iter.Key().Interface(), iter.Value().Interface())
			}
		}
	}
}

func renderAttr(out *strings.Builder, key any, value any) {
	name, ok := attrName(key)
	if !ok || name == "" {
		return
	}

	rendered, ok := attrValue(value)
	if !ok {
		return
	}
	writeAttr(out, name, rendered)
}

func writeAttr(out *strings.Builder, name string, value string) {
	out.WriteByte(' ')
	out.WriteString(name)
	out.WriteString(`="`)
	out.WriteString(html.EscapeString(value))
	out.WriteByte('"')
}

func attrName(key any) (string, bool) {
	switch value := key.(type) {
	case string:
		return strings.TrimPrefix(value, ":"), true
	case RawHTML:
		return strings.TrimPrefix(string(value), ":"), true
	default:
		rv := reflect.ValueOf(key)
		if !rv.IsValid() {
			return "", false
		}
		if rv.Kind() == reflect.String {
			return strings.TrimPrefix(rv.String(), ":"), true
		}
		return strings.TrimPrefix(fmt.Sprint(key), ":"), true
	}
}

func attrValue(value any) (string, bool) {
	switch v := value.(type) {
	case nil:
		return "", false
	case bool:
		if !v {
			return "", false
		}
		return "true", true
	case string:
		return v, true
	case RawHTML:
		return string(v), true
	case []any:
		return joinValues(v), true
	default:
		rv := reflect.ValueOf(value)
		if !rv.IsValid() {
			return "", false
		}
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			items := make([]any, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				items = append(items, rv.Index(i).Interface())
			}
			return joinValues(items), true
		case reflect.Bool:
			if !rv.Bool() {
				return "", false
			}
			return "true", true
		case reflect.String:
			return rv.String(), true
		default:
			return fmt.Sprint(value), true
		}
	}
}

func joinValues(values []any) string {
	parts := make([]string, 0, len(values))
	for _, value := range values {
		if rendered := textValue(value); rendered != "" {
			parts = append(parts, rendered)
		}
	}
	return strings.Join(parts, " ")
}

func textValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case RawHTML:
		return string(v)
	default:
		rv := reflect.ValueOf(value)
		if !rv.IsValid() {
			return ""
		}
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			items := make([]any, 0, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				items = append(items, rv.Index(i).Interface())
			}
			return joinValues(items)
		case reflect.String:
			return rv.String()
		default:
			return fmt.Sprint(value)
		}
	}
}

func isAttrMap(value any) bool {
	switch value.(type) {
	case map[any]any, map[string]any:
		return true
	}
	rv := reflect.ValueOf(value)
	return rv.IsValid() && rv.Kind() == reflect.Map
}

type tagInfo struct {
	name    string
	id      string
	classes []string
}

func parseTag(tag any) (tagInfo, bool) {
	name, ok := tagName(tag)
	if !ok || name == "" {
		return tagInfo{}, false
	}

	parsed := tagInfo{name: name}
	for i := 0; i < len(name); i++ {
		if name[i] != '#' && name[i] != '.' {
			continue
		}

		parsed.name = name[:i]
		rest := name[i:]
		for len(rest) > 0 {
			switch rest[0] {
			case '#':
				rest = rest[1:]
				next := strings.IndexAny(rest, "#.")
				if next < 0 {
					parsed.id = rest
					rest = ""
				} else {
					parsed.id = rest[:next]
					rest = rest[next:]
				}
			case '.':
				rest = rest[1:]
				next := strings.IndexAny(rest, "#.")
				if next < 0 {
					if rest != "" {
						parsed.classes = append(parsed.classes, rest)
					}
					rest = ""
				} else {
					if next > 0 {
						parsed.classes = append(parsed.classes, rest[:next])
					}
					rest = rest[next:]
				}
			default:
				rest = ""
			}
		}
		if parsed.name == "" {
			parsed.name = name[:i]
		}
		return parsed, true
	}

	return parsed, true
}

func tagName(tag any) (string, bool) {
	switch value := tag.(type) {
	case string:
		return strings.TrimPrefix(value, ":"), true
	case RawHTML:
		return strings.TrimPrefix(string(value), ":"), true
	default:
		rv := reflect.ValueOf(tag)
		if !rv.IsValid() {
			return "", false
		}
		if rv.Kind() == reflect.String {
			return strings.TrimPrefix(rv.String(), ":"), true
		}
		return strings.TrimPrefix(fmt.Sprint(tag), ":"), true
	}
}

func tagToken(tag any) (string, bool) {
	switch value := tag.(type) {
	case string:
		return value, true
	case RawHTML:
		return string(value), true
	default:
		rv := reflect.ValueOf(tag)
		if !rv.IsValid() {
			return "", false
		}
		if rv.Kind() == reflect.String {
			return rv.String(), true
		}
		return fmt.Sprint(tag), true
	}
}
