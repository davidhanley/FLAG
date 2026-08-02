package runtime

import "testing"

func TestBurpHtmlThroughGoFunctionBridge(t *testing.T) {
	fn := GoFunction("burp/html")
	got := Call(
		fn,
		NewArray(
			NewKeyword("div#app.hero"),
			NewMap(NewKeyword("data-role"), NewSymbol("main")),
			NewArray(NewKeyword("span"), NewSymbol("Hello & <world>")),
			NewArray(NewKeyword("br")),
		),
	)

	want := `<div id="app" class="hero" data-role="main"><span>Hello &amp; &lt;world&gt;</span><br></div>`
	if got.tag != TagString || got.StringValue() != want {
		t.Fatalf("unexpected burp html: %q", ValueToString(got))
	}
}
