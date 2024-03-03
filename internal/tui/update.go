package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type switch_view struct {
	new_state session_state
}

func switch_to(new_state session_state) tea.Cmd {
	// TODO this is a hack, we should be able to return a switch_view directly
	return tea.Tick(0, func(time.Time) tea.Msg {
		return switch_view{new_state}
	})
}

type clear_error_msg struct{}

func clear_error_after(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clear_error_msg{}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.state == options_state {
			m.list.SetSize(msg.Width, msg.Height)
		}
	case tea.KeyMsg:
		key := msg.String()

		// Global keybindings
		if key == "q" || key == "ctrl+c" {
			m.state = exit_state
			return m, tea.Quit
		}

		// State-specific keybindings
		if m.state == splash_state {
			switch key {
			case "enter":
				cmds = append(cmds, switch_to(options_state))
			}
		}

		if m.state == options_state {
			switch key {
			case "enter":
				selected_item, ok := m.list.SelectedItem().(item)
				if ok {
					m.choice = string(selected_item)
					cmds = append(cmds, m.filepicker.Init(), switch_to(filepicker_state))
				}
			}
		}

		// if m.state == filepicker_state {

		// }
	case switch_view:
		m.state = msg.new_state
	}

	switch m.state {
	case splash_state:
		// m, cmd = update_splash(m, msg)
	case options_state:
		m.list, cmd = m.list.Update(msg)
	case filepicker_state:
		m.filepicker, cmd = m.filepicker.Update(msg)

		switch msg.(type) {
		case clear_error_msg:
			m.err = nil
		}

		if did_select, path := m.filepicker.DidSelectFile(msg); did_select {
			m.selected_file = path
		}

		if did_select, path := m.filepicker.DidSelectDisabledFile(msg); did_select {
			m.err = fmt.Errorf("file %s is disabled", path)
			m.selected_file = ""
			return m, tea.Batch(cmd, clear_error_after(2*time.Second))
		}
	}

	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
