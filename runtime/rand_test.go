package runtime

import "testing"

func TestRandRange(t *testing.T) {
	seedRand(1)
	got := Rand()
	if got.tag != TagDouble || got.Double() < 0 || got.Double() >= 1 {
		t.Fatalf("rand: expected [0, 1), got %#v", got)
	}
	got = Rand(NewLong(10))
	if got.tag != TagDouble || got.Double() < 0 || got.Double() >= 10 {
		t.Fatalf("rand 10: expected [0, 10), got %#v", got)
	}
	assertPanics(t, func() { Rand(NewKeyword("a")) })
	assertPanics(t, func() { Rand(NewLong(1), NewLong(2)) })
}

func TestRandNth(t *testing.T) {
	if got := RandNth(NewArray(NewKeyword("a"))); got.tag != TagSymbol || got.SymbolObject().Name != "a" {
		t.Fatalf("rand-nth singleton: %#v", got)
	}
	seedRand(2)
	got := RandNth(NewArray(NewLong(1), NewLong(2), NewLong(3)))
	if got.tag != TagLong || got.Long() < 1 || got.Long() > 3 {
		t.Fatalf("rand-nth: %#v", got)
	}
	assertPanics(t, func() { RandNth(NewArray()) })
	assertPanics(t, func() { RandNth(NilValue()) })
}

func TestShuffle(t *testing.T) {
	if got := Shuffle(NewArray()); got.tag != TagArray || got.ArrayLen() != 0 {
		t.Fatalf("shuffle empty: %#v", got)
	}
	if got := Shuffle(NewArray(NewLong(1))); got.ArrayLen() != 1 || ArrayGet(got, 0).Long() != 1 {
		t.Fatalf("shuffle singleton: %#v", got)
	}
	seedRand(3)
	original := NewArray(NewLong(1), NewLong(2), NewLong(3), NewLong(4))
	shuffled := Shuffle(original)
	if shuffled.ArrayLen() != 4 {
		t.Fatalf("shuffle length: %#v", shuffled)
	}
	if ArrayGet(original, 0).Long() != 1 || ArrayGet(original, 3).Long() != 4 {
		t.Fatal("shuffle mutated original")
	}
	seen := map[int64]bool{}
	for i := 0; i < shuffled.ArrayLen(); i++ {
		seen[ArrayGet(shuffled, i).Long()] = true
	}
	if len(seen) != 4 || !seen[1] || !seen[2] || !seen[3] || !seen[4] {
		t.Fatalf("shuffle lost elements: %#v", shuffled)
	}
}
