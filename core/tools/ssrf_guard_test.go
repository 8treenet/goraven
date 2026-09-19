package tools

import "testing"

func TestValidatePublicHTTPURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"https to public ip", "https://1.1.1.1/index.html", false},
		{"http to public ip", "http://8.8.8.8:8080/path?q=1", false},
		{"file scheme", "file:///etc/passwd", true},
		{"ftp scheme", "ftp://example.com/file", true},
		{"gopher scheme", "gopher://127.0.0.1:6379/_info", true},
		{"empty host", "http:///path", true},
		{"missing scheme", "/etc/passwd", true},
		{"loopback v4", "http://127.0.0.1/admin", true},
		{"loopback alt v4", "http://127.1.0.1/", true},
		{"private v4 10/8", "http://10.0.0.1/", true},
		{"private v4 172.16/12", "http://172.16.0.1/", true},
		{"private v4 192.168/16", "http://192.168.1.1/", true},
		{"link local", "http://169.254.169.254/latest/meta-data", true},
		{"loopback v6", "http://[::1]/", true},
		{"private v6 ula", "http://[fd00::1]/", true},
		{"link local v6", "http://[fe80::1]/", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePublicHTTPURL(tc.url)
			if (err != nil) != tc.wantErr {
				t.Errorf("validatePublicHTTPURL(%q) error = %v, wantErr %v", tc.url, err, tc.wantErr)
			}
		})
	}
}

func TestValidatePublicHTTPURLResolvesHost(t *testing.T) {
	// localhost 在标准环境下解析为 127.0.0.1，应被拒绝
	if err := validatePublicHTTPURL("http://localhost:8000/api"); err == nil {
		t.Errorf("validatePublicHTTPURL(localhost) should fail, got nil")
	}
}
