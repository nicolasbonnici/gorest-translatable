package translatable

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

// Translation content is a rich text blob authored by trusted users (writers/admins).
// It is Markdown that may embed a small amount of raw HTML — notably <iframe> embeds
// for code samples and videos. We must strip XSS vectors (<script>, event handlers,
// javascript: URIs, ...) while preserving that legitimate markup, and we must not
// corrupt the surrounding JSON envelope ({"title":..,"content":..}).

// iframeSrc allowlists the hosts whose embeds we trust. An iframe whose src does not
// match loses its src attribute and renders as an inert empty frame.
var iframeSrc = regexp.MustCompile(
	`(?i)^https://(?:[a-z0-9-]+\.)*(?:codepen\.io|youtube\.com|youtube-nocookie\.com|player\.vimeo\.com)/`,
)

// fenced matches Markdown fenced code blocks; inlineCode matches inline code spans.
// Their contents may legitimately contain angle brackets (e.g. TypeScript generics),
// which an HTML sanitizer would otherwise mangle, so we shield them during sanitizing.
var (
	fenced     = regexp.MustCompile("(?s)```.*?```")
	inlineCode = regexp.MustCompile("`[^`\n]*`")
)

var (
	contentPolicyOnce sync.Once
	contentPolicy     *bluemonday.Policy
)

func policy() *bluemonday.Policy {
	contentPolicyOnce.Do(func() {
		p := bluemonday.UGCPolicy()
		p.AllowElements("iframe")
		p.AllowAttrs("src").Matching(iframeSrc).OnElements("iframe")
		p.AllowAttrs(
			"width", "height", "frameborder", "scrolling", "title",
			"allow", "allowfullscreen", "allowtransparency", "loading", "style",
		).OnElements("iframe")
		contentPolicy = p
	})
	return contentPolicy
}

// sanitizeRichText strips dangerous HTML from a single rich-text string while leaving
// Markdown (including code spans/blocks) intact.
func sanitizeRichText(s string) string {
	shield := map[string]string{}
	i := 0
	stash := func(m string) string {
		key := fmt.Sprintf("\x00cb%d\x00", i)
		i++
		shield[key] = m
		return key
	}

	s = fenced.ReplaceAllStringFunc(s, stash)
	s = inlineCode.ReplaceAllStringFunc(s, stash)

	s = policy().Sanitize(s)
	s = quoteUnescaper.Replace(s)

	for key, orig := range shield {
		s = strings.ReplaceAll(s, key, orig)
	}
	return s
}

// bluemonday entity-encodes quotes in text nodes, which mangles ordinary prose
// (e.g. "d'arcade" -> "d&#39;arcade"). Restoring the quote entities is safe because
// quote characters cannot form a tag; we deliberately leave &amp; &lt; &gt; encoded
// so a pre-escaped payload like "&lt;script&gt;" can never be reconstituted.
var quoteUnescaper = strings.NewReplacer(
	"&#39;", "'", "&#x27;", "'",
	"&#34;", `"`, "&#x22;", `"`, "&quot;", `"`,
)

// sanitizeContent sanitizes every string leaf of a translation content payload. The
// payload is expected to be a JSON object ({"title":..,"content":..}); if it does not
// parse as JSON it is treated as a single rich-text string. The returned value is always
// valid for storage in the jsonb column.
func sanitizeContent(raw string) string {
	var doc any
	if err := json.Unmarshal([]byte(raw), &doc); err != nil {
		return sanitizeRichText(raw)
	}
	cleaned := sanitizeNode(doc)
	out, err := json.Marshal(cleaned)
	if err != nil {
		return sanitizeRichText(raw)
	}
	return string(out)
}

func sanitizeNode(node any) any {
	switch v := node.(type) {
	case string:
		return sanitizeRichText(v)
	case []any:
		for i, item := range v {
			v[i] = sanitizeNode(item)
		}
		return v
	case map[string]any:
		for k, item := range v {
			v[k] = sanitizeNode(item)
		}
		return v
	default:
		return v
	}
}
