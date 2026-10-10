package envfile

import (
	"reflect"
	"testing"
)

func TestParseQuotedComments(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{"double quote in comment", `X="value" # say "hello"`, "X=value"},
		{"apostrophe in comment", `X='value' # don't edit`, "X=value"},
		{"unmatched double quote in comment", `X="value" # "`, "X=value"},
		{"empty double quoted value", `X="" # "example"`, "X="},
		{"empty single quoted value", `X='' # 'example'`, "X="},
		{"hash inside double quotes", `X="a#b" # "comment"`, "X=a#b"},
		{"hash inside single quotes", `X='a#b' # don't edit`, "X=a#b"},
		{"opposite quote in value", `X="it's fine" # "example"`, "X=it's fine"},
		{"literal backslash", `X="C:\temp\dir" # "example"`, `X=C:\temp\dir`},
		{"export", `export X='value' # don't edit`, "X=value"},
		{"no comment", `X="value"`, "X=value"},
		{"unquoted comment", `X=value # "example"`, "X=value"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse([]byte(tc.line + "\n"))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, []string{tc.want}) {
				t.Fatalf("Parse(%q) = %q, want %q", tc.line, got, tc.want)
			}
		})
	}
}

func TestParseUnterminatedQuotes(t *testing.T) {
	for _, line := range []string{`X="`, `X='`, `X="value`, `X='value`} {
		t.Run(line, func(t *testing.T) {
			if _, err := Parse([]byte(line)); err == nil {
				t.Fatalf("Parse(%q) should reject an unterminated quote", line)
			}
		})
	}
}
