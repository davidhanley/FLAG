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
	if !isType[*stringImpl](got) || got.StringValue() != want {
		t.Fatalf("unexpected burp html: %q", got.String())
	}
}

func TestBurpHtmlFlattensForLikeSequences(t *testing.T) {
	got := Call(
		GoBind_burp_Html,
		NewArray(
			NewKeyword("tr"),
			NewArray(NewKeyword("th"), NewString("Name")),
			NewArray(
				NewArray(NewKeyword("th"), NewString("Event 1")),
				NewArray(NewKeyword("th"), NewString("Event 2")),
			),
		),
	)
	want := `<tr><th>Name</th><th>Event 1</th><th>Event 2</th></tr>`
	if !isType[*stringImpl](got) || got.StringValue() != want {
		t.Fatalf("unexpected nested burp html: %q", got.String())
	}
}
