package runtime

import "testing"

func TestBurpHtmlThroughStaticBind(t *testing.T) {
	// Burp is a libraries/*.lib module for FLAG code; the runtime still exposes
	// the static GoBind_* adapters used by that library.
	got := Call(
		GoBind_burp_Html,
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
