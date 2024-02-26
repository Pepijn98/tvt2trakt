package tui

import "strings"

func (m model) View() string {
	switch m.state {
	case idle_state:
		// Prompt user to login
		return "Press enter to login to your trakt account"
	case show_login_state:
		// TODO: Login through oauth
		return ""
	case show_options_state:
		return "\n" + m.list.View()
	case show_filepicker_state:
		var s strings.Builder
		s.WriteString("\n  ")
		if m.err != nil {
			s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
		} else if m.selected_file == "" {
			s.WriteString("Pick a file:")
		} else {
			s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selected_file))
		}
		s.WriteString("\n\n" + m.filepicker.View() + "\n")
		return s.String()
	}
	return ""
}
