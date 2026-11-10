package envfile

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    []string
		wantErr bool
	}{
		{name: "plain", data: "DB_HOST=localhost\nAPI_KEY=123\n", want: []string{"DB_HOST=localhost", "API_KEY=123"}},
		{name: "comments and blank lines", data: "# header\n\nA=1\n   # indented comment\nB=2\n", want: []string{"A=1", "B=2"}},
		{name: "export prefix", data: "export A=1\n", want: []string{"A=1"}},
		{name: "empty values", data: "A=\nB=\"\"\n", want: []string{"A=", "B="}},
		{name: "crlf line endings", data: "A=1\r\nB=2\r\n", want: []string{"A=1", "B=2"}},
		{name: "double quoted keeps the hash", data: "A=\"x # y\"\n", want: []string{"A=x # y"}},
		{name: "trailing comment after a quoted value", data: "A=\"x\" # note\n", want: []string{"A=x"}},
		{name: "trailing comment containing the quote", data: "A=\"x\" # it's \"quoted\"\n", want: []string{"A=x"}},
		{name: "single quoted value with an apostrophe in the comment", data: "A='x' # it's fine\n", want: []string{"A=x"}},
		{name: "unquoted trailing comment", data: "A=x # note\n", want: []string{"A=x"}},
		{name: "text after the closing quote", data: "A=\"x\" junk\n", wantErr: true},
		{name: "unterminated quote", data: "A=\"x\n", wantErr: true},
		{name: "invalid key", data: "1BAD=x\n", wantErr: true},
		{name: "missing equals", data: "NOVALUE\n", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse([]byte(tc.data))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Parse(%q) = %v, want %v", tc.data, got, tc.want)
			}
		})
	}
}

func TestSerializeRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		entries []string
		wantErr bool
	}{
		{name: "plain", entries: []string{"DB_HOST=localhost", "API_KEY=123"}},
		{name: "with space", entries: []string{"MSG=hello world"}},
		{name: "with hash", entries: []string{"X=a#b"}},
		{name: "with single quote", entries: []string{"X=it's"}},
		{name: "empty value", entries: []string{"EMPTY="}},
		{name: "invalid key", entries: []string{"1BAD=x"}, wantErr: true},
		{name: "value with quote", entries: []string{`X=he said "hi"`}, wantErr: true},
		{name: "value with newline", entries: []string{"X=a\nb"}, wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			data, err := Serialize(tc.entries)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Serialize: %v", err)
			}
			got, err := Parse(data)
			if err != nil {
				t.Fatalf("Parse after Serialize: %v; data=%q", err, data)
			}
			if !reflect.DeepEqual(got, tc.entries) {
				t.Fatalf("round-trip mismatch:\n  got  %v\n  want %v\n  data %q", got, tc.entries, data)
			}
		})
	}
}
