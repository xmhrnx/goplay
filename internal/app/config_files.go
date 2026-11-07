package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var configTemplates = map[string]string{
	"mpv.conf": `# mpv options, one per line: name=value
# The full list: https://mpv.io/manual/stable/#options
# Read when goplay starts; these override goplay's own settings.
#
# Examples:
# sub-font="Segoe UI"
# screenshot-format=png
# deband-iterations=2
`,
	"input.conf": `# Key bindings for mpv commands, one per line: KEY command
# The syntax: https://mpv.io/manual/stable/#input-conf
# goplay's own shortcuts come first (change them in Settings > Shortcuts);
# keys goplay doesn't use reach these bindings.
#
# Examples:
# ctrl+h cycle-values hwdec "auto-safe" "no"
# g cycle-values video-aspect-override "16:9" "4:3" "no"
`,
}

// openConfigFolder opens goplay's configuration folder, where settings,
// mpv.conf, input.conf and scripts live.
func (a *App) openConfigFolder() {
	if a.configDir == "" {
		return
	}
	os.MkdirAll(filepath.Join(a.configDir, "scripts"), 0o755)
	openPath(a.configDir)
}

// editConfigFile opens mpv.conf or input.conf in a text editor, creating
// it with a short guide first.
func (a *App) editConfigFile(name string) {
	if a.configDir == "" {
		return
	}
	path := filepath.Join(a.configDir, name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(a.configDir, 0o755)
		os.WriteFile(path, []byte(configTemplates[name]), 0o644)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("notepad.exe", path)
	case "darwin":
		cmd = exec.Command("open", "-t", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	startDetached(cmd)
	a.toast(tr("{file} is read when GoPlay starts", "file", name))
}

func openPath(p string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", p)
	case "darwin":
		cmd = exec.Command("open", p)
	default:
		cmd = exec.Command("xdg-open", p)
	}
	startDetached(cmd)
}

func openWebPage(u string) { openPath(u) }

func startDetached(cmd *exec.Cmd) {
	if cmd.Start() == nil {
		go cmd.Wait()
	}
}
