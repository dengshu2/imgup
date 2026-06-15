---
name: imgup
description: Upload a local image (a screenshot, a generated image, any image file) to a Cloudflare R2 image host via the imgup CLI and get back a public URL — for blog images, notes, docs. Use when the user asks to upload an image / 上传图片 / 传图 / 把图传图床 / get a public link for this image / put this image on the CDN, or right after generating or capturing an image that needs to be embedded somewhere. The CLI prints the URL to stdout and logs to stderr.
---

# imgup

`imgup` is a standalone CLI that uploads images to a Cloudflare R2 bucket and prints the public URL. This skill tells an agent when and how to call it; all logic lives in the CLI.

## Prerequisites

1. **Binary on PATH** — check with `command -v imgup`. If missing, build/install per the project README.
2. **Credentials configured** — verify with `imgup doctor`; the last line should read `Checking R2 access... OK`. If it fails, the user must run `imgup init` once (the secret is stored in the OS keychain by default, not on disk).

Convention: **the public URL goes to stdout, progress/warnings go to stderr** — capture stdout as the result.

## Usage

```bash
imgup photo.png                  # bare URL (default)
imgup -f md photo.png            # ![photo](https://.../<md5>.png)  ← use when embedding in Markdown
imgup -f html photo.png          # <img src="..." alt="photo">
imgup --copy photo.png           # also copy the URL to the clipboard
imgup -f md a.png b.jpg c.gif     # multiple files, one URL per line
cat shot.png | imgup -            # read a single image from stdin (screenshot pipe)
```

## Typical flow: produce → upload → embed

When the user wants a freshly generated or captured image inserted into a document:

1. Produce the image to a local path (e.g. `/tmp/cover.png`).
2. `imgup -f md /tmp/cover.png` and capture stdout.
3. Insert that Markdown where it belongs.

Uploads are **idempotent**: the object key embeds the file's md5, so re-uploading identical bytes returns the same URL without re-transferring. Safe to call repeatedly.

## Troubleshooting

- `imgup doctor` — resolves the active config and does a real round-trip to R2; pinpoints whether the problem is credentials, endpoint, or network.
- A non-zero exit means at least one file failed; the reason is on stderr, while successful URLs are still printed to stdout.

See the project README for configuration, profiles, and the secret-resolution order.
