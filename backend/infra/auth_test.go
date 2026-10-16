package infra

import "testing"

func TestMatchSkipPath(t *testing.T) {
	cases := []struct {
		name    string
		urlPath string
		v       string
		want    bool
	}{
		// Full-prefix skip paths (HFS / user public routes)
		{"hit /api/hfs/ak prefix", "/api/hfs/ak/rvnt_x/docs/index.html", "/api/hfs/ak", true},
		{"hit /api/hfs/public prefix", "/api/hfs/public/abc", "/api/hfs/public", true},
		{"hit /api/user/login exact", "/api/user/login", "/api/user/login", true},
		{"hit /api/user/captcha exact", "/api/user/captcha", "/api/user/captcha", true},
		{"hit /api/install prefix", "/api/install/init", "/api/install", true},
		{"v without leading slash still prefixes", "/api/hfs/public/abc", "api/hfs/public", true},

		// Nested "public" under a protected route must NOT skip auth (#24)
		{"exploit nested /public under private", "/api/hfs/private/public/../victim/doc.pdf", "/api/hfs/public", false},
		{"nested /public segment with short v must not prefix-match", "/api/hfs/private/public/x", "/api/hfs/public", false},
		{"short /public must not match middle segment via prefix", "/api/hfs/private/public/x", "/public", false},
		{"short /public file-at-end still suffix-matches (legacy)", "/api/hfs/private/x/public", "/public", true},

		// Filename / directory-name lookalikes
		{"public in filename must NOT match full prefix", "/api/hfs/private/x/public_y.md", "/api/hfs/public", false},
		{"public_dir must NOT match full prefix", "/api/hfs/private/public_dir/z", "/api/hfs/public", false},
		{"ak in filename must NOT match", "/api/hfs/private/x/ak_backup.txt", "/api/hfs/ak", false},
		{"login_log must NOT match login prefix", "/api/user/login_log/x", "/api/user/login", false},

		// Share require-auth route: single-segment final match
		{"internalMessages final segment", "/api/share/abc/internalMessages", "/internalMessages", true},
		{"internalMessages not final", "/api/share/abc/internalMessages/extra", "/internalMessages", false},

		// Edge cases
		{"empty v", "/api/anything", "", false},
		{"empty needle after trim", "/api/anything", "/", false},
		{"v not present", "/api/hfs/private/x", "/api/hfs/public", false},
		{"exact equal", "/api/hfs/public", "/api/hfs/public", true},
		{"v longer than path", "/api/hfs", "/api/hfs/public", false},
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
