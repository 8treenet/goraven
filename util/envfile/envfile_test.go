package envfile

import (
	"os"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	data, err := os.ReadFile("/Users/ysmini/work/config/goraven/user/90a431bee756432492c134f510bad949/.profile")
	if err != nil {
		panic(err)
	}
	list, err := Parse(data)
	if err != nil {
		panic(err)
	}
	t.Log(list)
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
