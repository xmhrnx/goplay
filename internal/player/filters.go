package player

import "strings"

// SetFilter installs (or with "" removes) a video filter under a label,
// leaving other filters - the user's own from mpv.conf - alone.
//
// FFmpeg's filters work on frames in system memory, so while any of
// goplay's filters is active, hardware decoding switches to its
// copy-back mode (still decoding on the GPU), and returns to the user's
// mode once they're all removed.
func (p *Player) SetFilter(label, filter string) error {
	if p.filters[label] == filter {
		return nil
	}
	if _, ok := p.filters[label]; ok {
		p.h.Command("vf", "remove", "@"+label)
		delete(p.filters, label)
	}
	if filter != "" {
		p.copyBackHwdec(true)
		if err := p.h.Command("vf", "add", "@"+label+":"+filter); err != nil {
			p.copyBackHwdec(len(p.filters) > 0)
			return err
		}
		p.filters[label] = filter
	}
	p.copyBackHwdec(len(p.filters) > 0)
	return nil
}

// copyBackHwdec switches hardware decoding to copy-back mode, or back.
func (p *Player) copyBackHwdec(on bool) {
	switch {
	case on && p.hwdecSaved == "":
		cur, _ := p.h.String("hwdec")
		if cur == "" || cur == "no" || strings.HasSuffix(cur, "-copy") {
			return
		}
		p.hwdecSaved = cur
		p.h.SetString("hwdec", "auto-copy")
	case !on && p.hwdecSaved != "":
		p.h.SetString("hwdec", p.hwdecSaved)
		p.hwdecSaved = ""
	}
}

// FilterCommand sends a runtime command to a labeled filter, e.g. new
// angles to the 360° projection without rebuilding it.
func (p *Player) FilterCommand(label, cmd, arg string) error {
	return p.h.Command("vf-command", label, cmd, arg)
}
