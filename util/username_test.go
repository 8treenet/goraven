package util

import "testing"

func TestIsValidUsername(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"too short", "abc", false},
		{"too short at boundary", "1234567", false},
		{"empty", "", false},
		{"too long", "a234567890123456X", false},

		{"simple alphanumeric", "john1234", true},
		{"all letters", "abcdefgh", true},
		{"all digits", "12345678", true},
		{"with underscore", "john_doe", true},
		{"with dash", "john-doe", true},
		{"boundary min len", "12345678", true},
		{"boundary max len", "1234567890123456", true},

		{"leading dot", ".john1234", false},
		{"trailing dot", "john1234.", false},
		{"leading underscore", "_john1234", false},
		{"trailing underscore", "john1234_", false},
		{"leading dash", "-john1234", false},
		{"trailing dash", "john1234-", false},

		{"contains dot", "john.doe", false},
		{"path traversal in middle", "abc..def", false},
		{"path traversal with prefix", "abc..", false},
		{"path traversal with suffix", "..abc", false},

		{"contains slash", "john/doe", false},
		{"contains backslash", `john\doe`, false},
		{"contains space", "john doe", false},
		{"contains tab", "john\tdoe", false},
		{"contains newline", "john\ndoe", false},
		{"contains null byte", "john\x00doe", false},
		{"contains colon", "john:doe", false},
		{"contains asterisk", "john*doe", false},
		{"contains question", "john?doe", false},
		{"contains quote", `john"doe`, false},
		{"contains angle bracket", "john<doe", false},
		{"contains pipe", "john|doe", false},
		{"contains semicolon", "john;doe", false},
		{"contains ampersand", "john&doe", false},
		{"contains dollar", "john$doe", false},

		{"unicode chinese", "张三abcdef", false},
		{"unicode emoji", "john😀1234", false},

		{"windows reserved", "CON", false},
		{"windows reserved lower", "admin1234", true},
		{"absolute posix", "/etc/pass", false},
		{"tilde", "john~doe", false},
		{"at sign", "john@doe", false},
		{"plus", "john+doe", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsValidUsername(tc.input)
			if got != tc.want {
				t.Errorf("IsValidUsername(%q) = %v, want %v", tc.input, got, tc.want)
			}
		})
	}
}
