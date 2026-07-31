package packages

import "testing"

func TestToUppercase(t *testing.T) {
	if got := ToUppercase("hello"); got != "HELLO" {
		t.Fatalf("expected uppercased string, got %q", got)
	}
}

func TestRegisterCharacterAliases(t *testing.T) {
	registry := make(map[string]any)
	RegisterCharacter(func(name string, fn any) {
		registry[name] = fn
	})
	for _, name := range []string{"character/toUppercase", "character/toUpperCase", "Character/toUpperCase"} {
		if _, ok := registry[name]; !ok {
			t.Fatalf("expected registered alias %q", name)
		}
	}
}
