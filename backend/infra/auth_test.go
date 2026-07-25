package infra

import "testing"

func TestMatchSkipPath(t *testing.T) {
	cases := []struct {
		name    string
		urlPath string
		v       string
		want    bool
	}{
		{"hit /ak as segment", "/api/hfs/ak/rvnt_x/docs/index.html", "/ak", true},
		{"hit /public as segment", "/api/hfs/public/abc", "/public", true},
		{"hit /login as tail", "/api/user/login", "/login", true},
		{"hit /captcha as tail", "/api/user/captcha", "/captcha", true},
		{"hit /api/install prefix", "/api/install/init", "/api/install", true},
		{"v without leading slash", "/api/hfs/public/abc", "public", true},
		{"public in filename must NOT match", "/api/hfs/private/x/public_y.md", "/public", false},
		{"public_dir must NOT match", "/api/hfs/private/public_dir/z", "/public", false},
		{"ak in filename must NOT match", "/api/hfs/private/x/ak_backup.txt", "/ak", false},
		{"login in path middle must NOT match", "/api/user/login_log/x", "/login", false},
		{"empty v", "/api/anything", "", false},
		{"empty needle after trim", "/api/anything", "/", false},
		{"v not present", "/api/hfs/private/x", "/public", false},
		{"exact equal", "/public", "/public", true},
		{"v longer than path", "/pu", "/public", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := matchSkipPath(c.urlPath, c.v)
			if got != c.want {
				t.Fatalf("matchSkipPath(%q, %q) = %v, want %v", c.urlPath, c.v, got, c.want)
			}
		})
	}
}
