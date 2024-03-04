package tui

import (
	"fmt"
	"log"
	"strings"
)

const (
	bullet   = "•"
	ellipsis = "…"
)

func (m model) View() string {
	switch m.state {
	case splash_state:
		return splash_view()
	case login_state:
		return login_view()
	case options_state:
		return options_view(m)
	case filepicker_state:
		return filepicker_view(m)
	case exit_state:
		return exit_view()
	default:
		return "Unknown state"
	}
}

func splash_view() string {
	return "Press enter to login to your trakt account"
}

func login_view() string {
	return "Logging in..."
}

func options_view(m model) string {
	return "\n" + m.list.View()
}

func filepicker_view(m model) string {
	var s strings.Builder
	s.WriteString("\n  ")
	if m.err != nil {
		log.Panic(m.err)
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	} else if m.selected_file == "" {
		s.WriteString("Pick a file:")
	} else {
		s.WriteString("Selected file: " + m.filepicker.Styles.Selected.Render(m.selected_file))
	}
	s.WriteString("\n\n" + m.filepicker.View() + "\n")
	s.WriteString("\n" + help_style.Render(fmt.Sprintf("%s %s • %s %s • %s %s • %s %s", m.filepicker.KeyMap.Open.Help().Key, m.filepicker.KeyMap.Open.Help().Desc, m.filepicker.KeyMap.Back.Help().Key, m.filepicker.KeyMap.Back.Help().Desc, m.filepicker.KeyMap.Select.Help().Key, m.filepicker.KeyMap.Select.Help().Desc, "q/ctrl+c", "quit")))
	return s.String()
}

func exit_view() string {
	return "Goodbye!"
}
