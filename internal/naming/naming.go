// Package naming turns an image's bytes and original filename into a
// deterministic object key, using a small template language. The md5 of the
// content is the basis for deduplication: identical bytes always map to the
// same key.
package naming

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// Object describes a resolved upload target derived from content + template.
type Object struct {
	Key         string // e.g. "2026/06/<md5>.png"
	MD5         string // lowercase hex digest of the content
	Ext         string // extension without leading dot, lowercased (e.g. "png")
	ContentType string // best-effort MIME type
}

// Resolve computes the object key for the given content.
//
// origName is the source filename (used to derive the extension); it may be
// empty for piped/stdin input, in which case the extension is sniffed from the
// content. now is injected so the result is testable.
func Resolve(template string, content []byte, origName string, now time.Time) (Object, error) {
	sum := md5.Sum(content)
	digest := hex.EncodeToString(sum[:])

	ext := normalizeExt(filepath.Ext(origName))
	contentType := contentTypeFor(ext, content)
	if ext == "" {
		ext = extFromContentType(contentType)
	}

	key := expand(template, map[string]string{
		"year":     fmt.Sprintf("%04d", now.Year()),
		"month":    fmt.Sprintf("%02d", int(now.Month())),
		"day":      fmt.Sprintf("%02d", now.Day()),
		"md5":      digest,
		"ext":      ext,
		"filename": strings.TrimSuffix(filepath.Base(origName), filepath.Ext(origName)),
	})
	key = strings.TrimPrefix(key, "/")
	if key == "" {
		return Object{}, fmt.Errorf("empty object key after expanding template %q", template)
	}

	return Object{Key: key, MD5: digest, Ext: ext, ContentType: contentType}, nil
}

// expand replaces every {token} in s using the provided map. Unknown tokens are
// left untouched so a template typo is visible in the output rather than silent.
func expand(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "{"+k+"}", v)
	}
	return s
}

// normalizeExt lowercases and strips the leading dot from a file extension.
func normalizeExt(ext string) string {
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}

// contentTypeFor resolves a MIME type from the extension first (stable), then
// falls back to sniffing the content.
func contentTypeFor(ext string, content []byte) string {
	if ext != "" {
		if ct := mime.TypeByExtension("." + ext); ct != "" {
			return ct
		}
	}
	return http.DetectContentType(content)
}

// extFromContentType maps a handful of common image MIME types back to an
// extension, used only when the source name lacked one.
func extFromContentType(ct string) string {
	base := ct
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		base = ct[:i]
	}
	switch strings.TrimSpace(base) {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/webp":
		return "webp"
	case "image/svg+xml":
		return "svg"
	case "image/avif":
		return "avif"
	case "image/bmp":
		return "bmp"
	case "image/tiff":
		return "tiff"
	default:
		return "bin"
	}
}
