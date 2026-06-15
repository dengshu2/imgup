# imgup

[English](README.md) | [简体中文](README.zh-CN.md)

A small, dependency-free CLI that uploads images to a **Cloudflare R2** bucket and prints the public URL. It speaks the S3 API directly via `aws-sdk-go-v2` — no Node, no Python, no GUI, no background daemon. Drop the single static binary into your `PATH` and go.

`imgup` is built to be scriptable: the public URL goes to **stdout**, everything else (progress, warnings) goes to **stderr**, so it composes cleanly in pipelines and thin wrappers.

```console
$ imgup screenshot.png
https://img.example.com/2026/06/9f86d081884c7d659a2feaa0c55ad015.png
```

## Features

- Direct R2 / S3 upload (`region=auto` + custom endpoint), no third-party uploader runtime.
- Content-addressed, **idempotent** naming: the object key embeds the file's md5, so re-uploading the same bytes just returns the existing URL.
- Public URL is always `public_base + "/" + key` (custom domain), never `r2.dev`.
- Multiple named **profiles** for multiple buckets.
- Secret resolution chain: **flag → env → config file → macOS Keychain**.
- Output formats: bare URL, Markdown, or HTML; optional `--copy` to clipboard.

## Install

```sh
go install github.com/dengshu2/imgup@latest
```

Or build from source:

```sh
git clone https://github.com/dengshu2/imgup
cd imgup
go build -o imgup .
```

## Quick start

```sh
imgup init       # interactive: writes ~/.config/imgup/config.toml
imgup doctor     # resolves config and does a real round-trip to R2
imgup photo.jpg  # upload, print the URL
```

## Configuration

Config lives at `~/.config/imgup/config.toml` (override with `--config` or `$IMGUP_CONFIG`; `$XDG_CONFIG_HOME` is respected). It is written with mode `0600`.

```toml
default_profile = "default"

[profiles.default]
backend      = "r2"
endpoint     = "https://<account-id>.r2.cloudflarestorage.com"
bucket       = "img-disk"
public_base  = "https://img.example.com"
key_template = "{year}/{month}/{md5}.{ext}"
access_key_id = "<access-key-id>"

# Secret handling — pick ONE:
use_keychain = true            # secret read from the OS keychain (default on macOS)
# secret_access_key = "..."    # OR stored inline (portable; keep the file 0600)
```

### Where the secret lives

`imgup` never requires the secret in the config file. It is resolved in this order, first hit wins:

| Precedence | Source | How |
|-----------|--------|-----|
| 1 | flag | `--secret-access-key` |
| 2 | env | `IMGUP_SECRET_ACCESS_KEY` |
| 3 | config file | inline `secret_access_key` |
| 4 | OS keychain | when `use_keychain = true` (service `imgup`, account = profile) |

**Recommended:** on macOS, let `imgup init` store the secret in the Keychain. Your config file then contains no secret and is safe to sync or back up. For Linux/CI, use env vars or an inline secret in a `0600` file (the same model as `~/.aws/credentials`). `imgup doctor` warns if it finds an inline secret in a group/world-readable file.

### Key template tokens

`{year}` `{month}` `{day}` `{md5}` `{ext}` `{filename}`. The default `{year}/{month}/{md5}.{ext}` produces `2026/06/<md5>.png`.

### Environment variables

`IMGUP_CONFIG`, `IMGUP_PROFILE`-equivalent via `--profile`, `IMGUP_ENDPOINT`, `IMGUP_BUCKET`, `IMGUP_PUBLIC_BASE`, `IMGUP_KEY_TEMPLATE`, `IMGUP_ACCESS_KEY_ID`, `IMGUP_SECRET_ACCESS_KEY`, `IMGUP_BACKEND`.

## Usage

```sh
imgup <file>                 # upload, print URL to stdout
imgup -f md photo.png        # ![photo](https://.../photo.png)
imgup -f html photo.png      # <img src="..." alt="photo">
imgup --copy photo.png       # also copy the result to the clipboard
imgup --profile blog a.png   # use a non-default profile
imgup a.png b.jpg c.gif      # multiple files, one URL per line
cat shot.png | imgup -       # read a single image from stdin (screenshot pipe)
```

Flags: `-f, --format {url|md|html}` · `--copy` · `--profile` · `--config` · setting overrides (`--endpoint`, `--bucket`, `--public-base`, `--key-template`, `--access-key-id`, `--secret-access-key`).

### Commands

- `imgup init` — create/update a profile interactively.
- `imgup doctor` — resolve the active profile, show where each value came from, and verify R2 access with a real request.
- `imgup <file>...` — upload (the default action). `-` reads one image from stdin.
- `imgup ls [prefix]` — list objects; `-l` for size + mtime, `--url` to print public URLs, `--limit N` to cap.
- `imgup rm <key|url>...` — delete objects (accepts keys or full public URLs); `-y` skips the confirmation prompt.
- `imgup exists <file>` — check if a file's content is already stored; prints the URL and exits 0 if present, non-zero otherwise.
- `imgup skill install|print` — manage the bundled agent skill (see below).

## Companion skill

imgup ships an **agent skill** — a short instruction file that teaches an AI coding agent (Claude Code, Cursor, …) when and how to call imgup. It is embedded in the binary, so installing it needs no network and no extra files:

```sh
imgup skill install                    # install into every detected agent skills dir
imgup skill install --to ~/.agents/skills   # or target a specific dir (repeatable)
imgup skill print                      # dump SKILL.md to stdout
```

The source lives at [`skill/SKILL.md`](skill/SKILL.md). `install` writes it to `<dir>/imgup/SKILL.md`; with no `--to`, it fans out to every well-known agent skills directory that exists (`~/.agents/skills`, `~/.claude/skills`).

## Roadmap

The storage layer is an interface (`storage.Uploader`), leaving room for non-R2 / S3 backends without touching the command layer. Possible next steps: glob patterns, per-profile default output format, and a thin AI wrapper built on `imgup skill`.

## License

[MIT](LICENSE)
