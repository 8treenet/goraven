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
		{name: "plain", data: "DB_HOST=localhost\nPORT=8080\n", want: []string{"DB_HOST=localhost", "PORT=8080"}},
		{name: "blank and comment lines", data: "\n # comment\n\t\n", want: []string{}},
		{name: "export and unquoted comment", data: "export NAME = raven # comment\n", want: []string{"NAME=raven"}},
		{name: "quoted values", data: "A=\"hello # world\"\nB='another value'\nEMPTY=\n", want: []string{"A=hello # world", "B=another value", "EMPTY="}},
		{name: "duplicate keys preserve order", data: "KEY=first\nKEY=second", want: []string{"KEY=first", "KEY=second"}},
		{name: "invalid assignment", data: "NOT_AN_ASSIGNMENT\n", wantErr: true},
		{name: "invalid key", data: "1BAD=value\n", wantErr: true},
		{name: "unterminated quote", data: "KEY=\"value\n", wantErr: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse([]byte(tc.data))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected parse error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Parse: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("Parse: got %v, want %v", got, tc.want)
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
