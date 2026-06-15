// Package clip copies text to the system clipboard by shelling out to the
// platform's standard tool. This avoids any cgo dependency, keeping imgup a
// single static binary.
package clip

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Copy writes s to the system clipboard.
func Copy(s string) error {
	name, args, err := command()
	if err != nil {
		return err
	}
	cmd := exec.Command(name, args...)
	cmd.Stdin = strings.NewReader(s)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w: %s", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// command picks the clipboard tool for the current OS. On Linux it prefers
// Wayland's wl-copy, then X11's xclip, then xsel.
func command() (string, []string, error) {
	switch runtime.GOOS {
	case "darwin":
		return "pbcopy", nil, nil
	case "windows":
		return "clip", nil, nil
	case "linux":
		if p, err := exec.LookPath("wl-copy"); err == nil {
			return p, nil, nil
		}
		if p, err := exec.LookPath("xclip"); err == nil {
			return p, []string{"-selection", "clipboard"}, nil
		}
		if p, err := exec.LookPath("xsel"); err == nil {
			return p, []string{"--clipboard", "--input"}, nil
		}
		return "", nil, fmt.Errorf("no clipboard tool found (install wl-clipboard, xclip, or xsel)")
	default:
		return "", nil, fmt.Errorf("clipboard not supported on %s", runtime.GOOS)
	}
}
