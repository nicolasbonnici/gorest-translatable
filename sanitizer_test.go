package translatable

import (
	"encoding/json"
	"strings"
	"testing"
)

func contentField(t *testing.T, raw string) (title, content string) {
	t.Helper()
	var doc struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		t.Fatalf("result is not valid JSON: %v\n%s", err, raw)
	}
	return doc.Title, doc.Content
}

func TestSanitizeContent_StripsScriptAndHandlers(t *testing.T) {
	in := `{"title":"Hi <script>alert(1)</script>","content":"ok <img src=x onerror=alert(1)> <a href=\"javascript:alert(1)\">x</a>"}`
	title, content := contentField(t, sanitizeContent(in))

	for _, bad := range []string{"<script", "onerror", "javascript:"} {
		if strings.Contains(title+content, bad) {
			t.Errorf("expected %q to be stripped; got title=%q content=%q", bad, title, content)
		}
	}
	if !strings.HasPrefix(title, "Hi") {
		t.Errorf("safe text should survive, got title=%q", title)
	}
}

func TestSanitizeContent_KeepsTrustedIframe(t *testing.T) {
	in := `{"title":"t","content":"<iframe src=\"https://codepen.io/editor/nicolasbonnici/embed/abc\" style=\"width: 100%;\" scrolling=\"no\" title=\"LOLCat\"></iframe>"}`
	_, content := contentField(t, sanitizeContent(in))

	if !strings.Contains(content, "<iframe") || !strings.Contains(content, "codepen.io") {
		t.Errorf("trusted codepen iframe should be preserved, got %q", content)
	}
}

func TestSanitizeContent_DropsUntrustedIframeSrc(t *testing.T) {
	in := `{"title":"t","content":"<iframe src=\"https://evil.example.com/x\"></iframe>"}`
	_, content := contentField(t, sanitizeContent(in))

	if strings.Contains(content, "evil.example.com") {
		t.Errorf("untrusted iframe src should be dropped, got %q", content)
	}
}

func TestSanitizeContent_PreservesMarkdownAndCode(t *testing.T) {
	// Code fence containing angle brackets must not be mangled by the HTML sanitizer.
	in := `{"title":"t","content":"## Heading\n\n[link](https://nbonnici.info)\n\n` + "```" + `typescript\nconst x: Array<Map<string>> = []\n` + "```" + `"}`
	_, content := contentField(t, sanitizeContent(in))

	for _, want := range []string{"## Heading", "[link](https://nbonnici.info)", "Array<Map<string>>"} {
		if !strings.Contains(content, want) {
			t.Errorf("expected markdown/code %q preserved, got %q", want, content)
		}
	}
}

func TestSanitizeContent_PreservesQuotesInProse(t *testing.T) {
	in := `{"title":"t","content":"les bornes d'arcade de mon enfance"}`
	_, content := contentField(t, sanitizeContent(in))
	if !strings.Contains(content, "d'arcade") {
		t.Errorf("apostrophe should not be entity-encoded, got %q", content)
	}
	if strings.Contains(content, "&#39;") {
		t.Errorf("quote entity leaked into output: %q", content)
	}
}

func TestSanitizeContent_DoesNotRebuildPreEscapedScript(t *testing.T) {
	// An author-supplied, already-escaped payload must NOT be turned back into live markup.
	in := `{"title":"t","content":"harmless &lt;script&gt;alert(1)&lt;/script&gt; text"}`
	_, content := contentField(t, sanitizeContent(in))
	if strings.Contains(content, "<script") {
		t.Errorf("must not reconstruct a live script tag, got %q", content)
	}
}

func TestSanitizeContent_OutputAlwaysValidJSON(t *testing.T) {
	// Even a nasty payload must yield storable JSON (the original bug: invalid jsonb).
	in := `{"title":"a \"quote\" & <b>bold</b>","content":"x"}`
	out := sanitizeContent(in)
	var doc map[string]any
	if err := json.Unmarshal([]byte(out), &doc); err != nil {
		t.Fatalf("sanitized output must be valid JSON: %v\n%s", err, out)
	}
}
