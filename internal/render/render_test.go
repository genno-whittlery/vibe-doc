package render

import (
	"strings"
	"testing"
)

// Regression: $50 and $60 in prose must NOT be wrapped as math.
func TestRenderInlineMathDoesNotEatPrices(t *testing.T) {
	r := New()
	html, _, err := r.Render([]byte("cost was $50 and $60\n"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(html), `class="math"`) {
		t.Errorf("prices wrongly wrapped as math: %s", html)
	}
}

// Mermaid fence with trailing whitespace on close still matches.
func TestRenderMermaidTrailingSpaceOnClose(t *testing.T) {
	r := New()
	src := "```mermaid\ngraph TD;\n  A-->B;\n```   \n"
	html, _, err := r.Render([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), `class="mermaid"`) {
		t.Errorf("expected mermaid pre even with trailing close-fence space: %s", html)
	}
}

// Heading with inline code should produce TOC text containing the code.
func TestTOCWithInlineCodeInHeading(t *testing.T) {
	r := New()
	_, toc, err := r.Render([]byte("## Use `sdk.world`\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(toc) != 1 {
		t.Fatalf("expected 1 TOC entry, got %d", len(toc))
	}
	if !strings.Contains(toc[0].Text, "sdk.world") {
		t.Errorf("expected sdk.world in TOC text, got %q", toc[0].Text)
	}
}

func TestRenderBasic(t *testing.T) {
	r := New()
	html, toc, err := r.Render([]byte("# Title\n\nBody **bold**.\n"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), "<h1") {
		t.Errorf("expected <h1>, got: %s", html)
	}
	if !strings.Contains(string(html), "<strong>bold</strong>") {
		t.Errorf("expected bold: %s", html)
	}
	if len(toc) != 1 || toc[0].Text != "Title" || toc[0].Level != 1 {
		t.Errorf("TOC = %v", toc)
	}
}

func TestRenderMermaid(t *testing.T) {
	r := New()
	src := "```mermaid\ngraph TD;\n  A-->B;\n```\n"
	html, _, err := r.Render([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), `class="mermaid"`) {
		t.Errorf("expected <pre class=\"mermaid\">: %s", html)
	}
	if !strings.Contains(string(html), "graph TD;") {
		t.Errorf("expected raw diagram source preserved: %s", html)
	}
}

func TestRenderMath(t *testing.T) {
	r := New()
	src := "Inline $E = mc^2$ block: $$ \\int f $$\n"
	html, _, err := r.Render([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(html), `class="math"`) {
		t.Errorf("expected math wrapping: %s", html)
	}
}

func TestTOCMultiHeader(t *testing.T) {
	r := New()
	src := "# One\n## Two\n### Three\n## Four\n"
	_, toc, err := r.Render([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		Text  string
		Level int
	}{
		{"One", 1}, {"Two", 2}, {"Three", 3}, {"Four", 2},
	}
	if len(toc) != len(want) {
		t.Fatalf("len=%d want %d", len(toc), len(want))
	}
	for i := range want {
		if toc[i].Text != want[i].Text || toc[i].Level != want[i].Level {
			t.Errorf("toc[%d] = %+v, want %+v", i, toc[i], want[i])
		}
	}
}
