package burp

import "testing"

func TestHtmlRendersNestedElements(t *testing.T) {
	got := Html(
		[]any{
			":div#app.hero",
			map[any]any{":data-role": "main"},
			[]any{":span", "Hello & <world>"},
			[]any{":br"},
		},
	)

	want := `<div id="app" class="hero" data-role="main"><span>Hello &amp; &lt;world&gt;</span><br></div>`
	if got != want {
		t.Fatalf("unexpected html: %q", got)
	}
}

func TestHtml5PrefixesDoctype(t *testing.T) {
	got := Html5([]any{":html", []any{":body", "ok"}})
	want := `<!DOCTYPE html><html><body>ok</body></html>`
	if got != want {
		t.Fatalf("unexpected html5: %q", got)
	}
}

func TestRawBypassesEscaping(t *testing.T) {
	got := Html([]any{":p", Raw("<strong>raw</strong>")})
	want := `<p><strong>raw</strong></p>`
	if got != want {
		t.Fatalf("unexpected raw html: %q", got)
	}
}
