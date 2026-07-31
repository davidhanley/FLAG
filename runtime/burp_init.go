package runtime

import burp "flag-lang/libraries/burp"

func init() {
	RegisterGoFunction("burp/html", burp.Html)
	RegisterGoFunction("burp/html5", burp.Html5)
	RegisterGoFunction("burp/escape", burp.Escape)
	RegisterGoFunction("burp/raw", burp.Raw)
}
