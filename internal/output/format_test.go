package output

import "testing"

func TestParse(t *testing.T) {
	for _, s := range []string{"url", "md", "html"} {
		if _, err := Parse(s); err != nil {
			t.Errorf("Parse(%q) errored: %v", s, err)
		}
	}
	if _, err := Parse("xml"); err == nil {
		t.Error("Parse(xml) should error")
	}
}

func TestRender(t *testing.T) {
	const url = "https://img.example.com/2026/06/abc.png"
	cases := []struct {
		f    Format
		want string
	}{
		{FormatURL, url},
		{FormatMarkdown, "![pic](" + url + ")"},
		{FormatHTML, `<img src="` + url + `" alt="pic">`},
	}
	for _, c := range cases {
		if got := Render(c.f, url, "pic"); got != c.want {
			t.Errorf("Render(%q) = %q, want %q", c.f, got, c.want)
		}
	}
}
