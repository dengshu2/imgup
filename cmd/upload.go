package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dengshu2/imgup/internal/clip"
	"github.com/dengshu2/imgup/internal/config"
	"github.com/dengshu2/imgup/internal/naming"
	"github.com/dengshu2/imgup/internal/output"
	"github.com/dengshu2/imgup/internal/storage"
)

var (
	flagFormat string
	flagCopy   bool
)

func init() {
	rootCmd.Flags().StringVarP(&flagFormat, "format", "f", "url", "output format: url, md, html")
	rootCmd.Flags().BoolVar(&flagCopy, "copy", false, "copy the resulting URL(s) to the clipboard")
}

// runUpload uploads each path argument. "-" reads a single image from stdin.
func runUpload(ctx context.Context, paths []string) error {
	format, err := output.Parse(flagFormat)
	if err != nil {
		return err
	}

	resolved, up, err := resolvedUploader(ctx)
	if err != nil {
		return err
	}

	var rendered []string
	var failed int
	for _, p := range paths {
		url, err := uploadOne(ctx, up, resolved, p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "imgup: %s: %v\n", p, err)
			failed++
			continue
		}
		line := output.Render(format, url, altFor(p))
		fmt.Println(line) // stdout: the machine-readable result
		rendered = append(rendered, line)
	}

	if flagCopy && len(rendered) > 0 {
		if err := clip.Copy(strings.Join(rendered, "\n")); err != nil {
			fmt.Fprintf(os.Stderr, "imgup: clipboard: %v\n", err)
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d of %d upload(s) failed", failed, len(paths))
	}
	return nil
}

// uploadOne reads, names, and (idempotently) uploads a single source, returning
// its public URL.
func uploadOne(ctx context.Context, up storage.Uploader, r *config.Resolved, src string) (string, error) {
	content, origName, err := readSource(src)
	if err != nil {
		return "", err
	}

	obj, err := naming.Resolve(r.KeyTemplate, content, origName, time.Now())
	if err != nil {
		return "", err
	}

	exists, err := up.Exists(ctx, obj.Key)
	if err != nil {
		return "", err
	}
	if exists {
		fmt.Fprintf(os.Stderr, "imgup: %s already exists, skipping upload\n", obj.Key)
	} else {
		if err := up.Put(ctx, obj.Key, content, obj.ContentType); err != nil {
			return "", err
		}
		fmt.Fprintf(os.Stderr, "imgup: uploaded %s (%d bytes, %s)\n", obj.Key, len(content), obj.ContentType)
	}

	return joinURL(r.PublicBase, obj.Key), nil
}

// readSource returns the bytes and original filename for a path, or stdin when
// src is "-".
func readSource(src string) (content []byte, origName string, err error) {
	if src == "-" {
		b, err := io.ReadAll(os.Stdin)
		return b, "", err
	}
	b, err := os.ReadFile(src)
	if err != nil {
		return nil, "", err
	}
	return b, src, nil
}

// altFor derives alt text from a path; empty for stdin.
func altFor(src string) string {
	if src == "-" {
		return ""
	}
	base := src
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}
	if i := strings.LastIndexByte(base, '.'); i > 0 {
		base = base[:i]
	}
	return base
}

// joinURL concatenates the public base and key with exactly one slash.
func joinURL(base, key string) string {
	return strings.TrimRight(base, "/") + "/" + strings.TrimLeft(key, "/")
}
