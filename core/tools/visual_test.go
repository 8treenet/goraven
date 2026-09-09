package tools

import "testing"

func TestClassifyMediaType(t *testing.T) {
	cases := []struct {
		ext  string
		want string
	}{
		{".jpg", MediaTypeImage},
		{".JPG", MediaTypeImage},
		{".png", MediaTypeImage},
		{".webp", MediaTypeImage},
		{".svg", MediaTypeImage},
		{".mp4", MediaTypeVideo},
		{".mov", MediaTypeVideo},
		{".mkv", MediaTypeVideo},
		{".mp3", MediaTypeAudio},
		{".wav", MediaTypeAudio},
		{".pdf", ""},
		{".zip", ""},
		{"", ""},
	}
	for _, c := range cases {
		if got := ClassifyMediaType(c.ext); got != c.want {
			t.Errorf("ClassifyMediaType(%q) = %q, want %q", c.ext, got, c.want)
		}
	}
}
