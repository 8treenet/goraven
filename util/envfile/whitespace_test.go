package envfile

import (
	"reflect"
	"testing"
)

func TestSerializePreservesEdgeWhitespace(t *testing.T) {
	values := []string{
		"\u00a0value",
		"value\u00a0",
		"\u2003value\u2003",
		"\u3000",
		"\vvalue",
		"value\f",
		"\tvalue\t",
	}
	for _, value := range values {
		t.Run(value, func(t *testing.T) {
			entries := []string{"VALUE=" + value}
			data, err := Serialize(entries)
			if err != nil {
				t.Fatal(err)
			}
			got, err := Parse(data)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, entries) {
				t.Fatalf("round-trip mismatch: got %q, want %q; data %q", got, entries, data)
			}
		})
	}
}
