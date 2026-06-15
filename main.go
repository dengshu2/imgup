// Command imgup is a lightweight, dependency-free CLI that uploads images to a
// Cloudflare R2 (S3-compatible) bucket and prints the resulting public URL.
//
// It is designed to be scriptable: the public URL is written to stdout while all
// diagnostics go to stderr, so it composes cleanly in pipelines and thin wrappers.
package main

import (
	_ "embed"

	"github.com/dengshu2/imgup/cmd"
)

// skillMD is the companion agent skill, bundled into the binary so
// `imgup skill install` works offline with no extra files.
//
//go:embed skill/SKILL.md
var skillMD string

func main() {
	cmd.SkillAsset = skillMD
	cmd.Execute()
}
