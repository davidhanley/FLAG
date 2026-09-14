package compiler

import "testing"

var astCannedCases = []struct {
	source string
	want   string
}{
	{"", ""},
	{"foo", "(:symbol :name foo :line 1 :col 1)"},
	{"(+ x 10)", "(:list :line 1 :col 1 :elements [(:symbol :name + :line 1 :col 2) (:symbol :name x :line 1 :col 4) (:int :value 10 :line 1 :col 6)])"},
	{"()\n[]\n{}", "(:list :line 1 :col 1 :elements [])\n(:vector :line 2 :col 1 :elements [])\n(:map :line 3 :col 1 :entries [])"},
	{"(outer [inner {:key value}] {:pair [left right] :call (f g)} \"text\")", "(:list :line 1 :col 1 :elements [(:symbol :name outer :line 1 :col 2) (:vector :line 1 :col 8 :elements [(:symbol :name inner :line 1 :col 9) (:map :line 1 :col 15 :entries [(:keyword :name key :line 1 :col 16) (:symbol :name value :line 1 :col 21)])]) (:map :line 1 :col 29 :entries [(:keyword :name pair :line 1 :col 30) (:vector :line 1 :col 36 :elements [(:symbol :name left :line 1 :col 37) (:symbol :name right :line 1 :col 42)]) (:keyword :name call :line 1 :col 49) (:list :line 1 :col 55 :elements [(:symbol :name f :line 1 :col 56) (:symbol :name g :line 1 :col 58)])]) (:string :value \"text\" :line 1 :col 62)])"},
	{"(println \"hi\")", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:string :value \"hi\" :line 1 :col 10)])"},
	{"(println \"\"\"hello\nworld\"\"\")", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:string :value \"hello\\nworld\" :line 1 :col 10)])"},
	{"(println 'abc :xyz)", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:quoted-symbol :name abc :line 1 :col 10) (:keyword :name xyz :line 1 :col 15)])"},
	{"(println '(1 2 3))", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:quoted-list :line 1 :col 10 :elements [(:int :value 1 :line 1 :col 12) (:int :value 2 :line 1 :col 14) (:int :value 3 :line 1 :col 16)])])"},
	{"(println 5/6)", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:ratio :numerator 5 :denominator 6 :line 1 :col 10)])"},
	{"(println 10N -999999999999999999999N)", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:bigint :value 10 :line 1 :col 10) (:bigint :value -999999999999999999999 :line 1 :col 14)])"},
	{"(println \\M \\space)", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:char :value \"M\" :line 1 :col 10) (:char :value \" \" :line 1 :col 13)])"},
	{"(println 1.5 1e2)", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:float :value 1.5 :raw 1.5 :line 1 :col 10) (:float :value 100 :raw 1e2 :line 1 :col 14)])"},
	{"(println #{1 2 3})", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:set :line 1 :col 10 :elements [(:int :value 1 :line 1 :col 12) (:int :value 2 :line 1 :col 14) (:int :value 3 :line 1 :col 16)])])"},
	{"(println #(* % 3))", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:hash-fn :line 1 :col 10 :body (:list :line 1 :col 10 :elements [(:symbol :name * :line 1 :col 12) (:symbol :name % :line 1 :col 14) (:int :value 3 :line 1 :col 16)]))])"},
	{"(def v | 1 2 3 |)", "(:list :line 1 :col 1 :elements [(:symbol :name def :line 1 :col 2) (:symbol :name v :line 1 :col 6) (:pipe-vector :line 1 :col 8 :elements [(:int :value 1 :line 1 :col 10) (:int :value 2 :line 1 :col 12) (:int :value 3 :line 1 :col 14)])])"},
	{"^long value", "(:meta :line 1 :col 1 :meta (:symbol :name long :line 1 :col 2) :target (:symbol :name value :line 1 :col 7))"},
	{"(let [^{:volatile true} seen {}] seen)", "(:list :line 1 :col 1 :elements [(:symbol :name let :line 1 :col 2) (:vector :line 1 :col 6 :elements [(:meta :line 1 :col 7 :meta (:map :line 1 :col 8 :entries [(:keyword :name volatile :line 1 :col 9) (:symbol :name true :line 1 :col 19)]) :target (:symbol :name seen :line 1 :col 25)) (:map :line 1 :col 30 :entries [])]) (:symbol :name seen :line 1 :col 34)])"},
	{"(println 1)\n(comment (println 2) (+ 1 2))\n(println 3)", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:int :value 1 :line 1 :col 10)])\n(:list :line 3 :col 1 :elements [(:symbol :name println :line 3 :col 2) (:int :value 3 :line 3 :col 10)])"},
	{"(do 1 (comment (println 2) {:a 1}) 3)", "(:list :line 1 :col 1 :elements [(:symbol :name do :line 1 :col 2) (:int :value 1 :line 1 :col 5) (:int :value 3 :line 1 :col 36)])"},
	{"true false nil", "(:symbol :name true :line 1 :col 1)\n(:symbol :name false :line 1 :col 6)\n(:symbol :name nil :line 1 :col 12)"},
	{"(println \"a\\nb\")", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:string :value \"a\\nb\" :line 1 :col 10)])"},
	{"(println \"hi\\n\\r\")", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:string :value \"hi\\n\\r\" :line 1 :col 10)])"},
	{"(println \"\\\"text\\\"\")", "(:list :line 1 :col 1 :elements [(:symbol :name println :line 1 :col 2) (:string :value \"\\\"text\\\"\" :line 1 :col 10)])"},
	{"(ok) (nope", "ERROR: parse error at 1:6: missing closing \")\""},
	{"(println \"x\"", "ERROR: parse error at 1:1: missing closing \")\""},
	{"'", "ERROR: parse error at 1:2: unexpected end of input"},
}

func TestFLAGASTBuilderMatchesCanned(t *testing.T) {
	for i, tc := range astCannedCases {
		got := astCanonical(tc.source)
		if got != tc.want {
			t.Fatalf("case %d source %q\n got: %s\nwant: %s", i, tc.source, got, tc.want)
		}
	}
}
