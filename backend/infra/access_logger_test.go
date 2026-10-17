package infra

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMaskSensitiveJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "mask login password",
			input: `{"username":"admin","password":"secret123"}`,
			want:  `{"username":"admin","password":"***"}`,
		},
		{
			name:  "mask current and new password",
			input: `{"currentPassword":"a","newPassword":"b"}`,
			want:  `{"currentPassword":"***","newPassword":"***"}`,
		},
		{
			name:  "mask nested password",
			input: `{"user":{"name":"u","password":"p"},"token":"x"}`,
			want:  `{"user":{"name":"u","password":"***"},"token":"x"}`,
		},
		{
			name:  "mask password in array",
			input: `{"list":[{"password":"p"},{"name":"n"}]}`,
			want:  `{"list":[{"password":"***"},{"name":"n"}]}`,
		},
		{
			name:  "keep normal body",
			input: `{"content":"hello","sessionId":"abc"}`,
			want:  `{"content":"hello","sessionId":"abc"}`,
		},
		{
			name:  "unparseable json",
			input: `{"password": "broken`,
			want:  "[unparseable json body omitted]",
		},
		{
			name:  "empty body",
			input: ``,
			want:  ``,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := maskSensitiveJSON([]byte(tc.input))
			if tc.want == "" {
				if got != tc.want {
					t.Errorf("maskSensitiveJSON(%q) = %q, want %q", tc.input, got, tc.want)
				}
				return
			}
			// JSON 对象 key 顺序不保证，做结构化比较
			if tc.want != "[unparseable json body omitted]" {
				var gotVal, wantVal interface{}
				if err := json.Unmarshal([]byte(got), &gotVal); err != nil {
					t.Fatalf("output is not valid json: %s", got)
				}
				if err := json.Unmarshal([]byte(tc.want), &wantVal); err != nil {
					t.Fatalf("want is not valid json: %s", tc.want)
				}
				if !reflect.DeepEqual(gotVal, wantVal) {
					t.Errorf("maskSensitiveJSON(%q) = %q, want %q", tc.input, got, tc.want)
				}
				return
			}
			if got != tc.want {
				t.Errorf("maskSensitiveJSON(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestIsSensitiveField(t *testing.T) {
	for _, key := range []string{"password", "Password", "currentPassword", "newPassword", "user_passwd", "clientSecret"} {
		if !isSensitiveField(key) {
			t.Errorf("isSensitiveField(%q) = false, want true", key)
		}
	}
	for _, key := range []string{"username", "content", "sessionId"} {
		if isSensitiveField(key) {
			t.Errorf("isSensitiveField(%q) = true, want false", key)
		}
	}
}
