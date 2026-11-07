package actions

// Action is one command.
type Action struct {
	ID       string
	Title    string
	Category string
	// Keys are the default shortcuts, written portably: "Mod" is ⌘ on
	// macOS and Ctrl elsewhere ("Mod+O", "Shift+Right", "Space").
	Keys []string
	Run  func()
	// Enabled reports whether the command can run now (nil: always).
	Enabled func() bool
	// Checked reports an on/off state to show (nil: not a toggle).
	Checked func() bool
	// Repeat lets a held key repeat the command (seeking, volume).
	Repeat bool
	// Hidden keeps it out of the command palette.
	Hidden bool
}

// IsEnabled reports whether the action can run now.
func (a *Action) IsEnabled() bool { return a.Enabled == nil || a.Enabled() }

// IsChecked reports the action's on/off state.
func (a *Action) IsChecked() bool { return a.Checked != nil && a.Checked() }
