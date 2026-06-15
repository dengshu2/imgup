package cmd

import "testing"

func TestKeyFromArg(t *testing.T) {
	const base = "https://img.example.com"
	cases := []struct{ arg, want string }{
		{"https://img.example.com/2026/06/a.png", "2026/06/a.png"},
		{"2026/06/a.png", "2026/06/a.png"},
		{"/2026/06/a.png", "2026/06/a.png"},
		{"https://img.example.com/x.png", "x.png"},
	}
	for _, c := range cases {
		if got := keyFromArg(c.arg, base); got != c.want {
			t.Errorf("keyFromArg(%q) = %q, want %q", c.arg, got, c.want)
		}
	}
}

func TestKeyFromArgTrailingSlashBase(t *testing.T) {
	// A public_base with a trailing slash must still strip cleanly.
	if got := keyFromArg("https://img.example.com/p/a.png", "https://img.example.com/"); got != "p/a.png" {
		t.Errorf("got %q, want p/a.png", got)
	}
}

func TestJoinURL(t *testing.T) {
	cases := []struct{ base, key, want string }{
		{"https://img.example.com", "a/b.png", "https://img.example.com/a/b.png"},
		{"https://img.example.com/", "a/b.png", "https://img.example.com/a/b.png"},
		{"https://img.example.com", "/a/b.png", "https://img.example.com/a/b.png"},
	}
	for _, c := range cases {
		if got := joinURL(c.base, c.key); got != c.want {
			t.Errorf("joinURL(%q,%q) = %q, want %q", c.base, c.key, got, c.want)
		}
	}
}

func TestAltFor(t *testing.T) {
	cases := []struct{ src, want string }{
		{"photo.png", "photo"},
		{"dir/sub/shot.jpeg", "shot"},
		{"-", ""},
		{"noext", "noext"},
	}
	for _, c := range cases {
		if got := altFor(c.src); got != c.want {
			t.Errorf("altFor(%q) = %q, want %q", c.src, got, c.want)
		}
	}
}
