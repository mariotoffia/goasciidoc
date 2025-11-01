package goparser

import "testing"

func TestIsExportedNormalization(t *testing.T) {
	cases := map[string]bool{
		"Foo":                     true,
		"*Foo":                    true,
		"[]*Foo":                  true,
		"map[string]*pkg.Foo":     true,
		"chan<- Foo":              true,
		"<-chan Foo":              true,
		"(...Foo)":                true,
		"pkg.Bar":                 true,
		"[]string":                false,
		"map[int]*pkg.unexported": false,
		"func()":                  false,
	}

	for input, expected := range cases {
		if got := isExported(input); got != expected {
			t.Fatalf("expected isExported(%q) to be %v, got %v", input, expected, got)
		}
	}
}
