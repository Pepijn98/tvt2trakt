package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
		m.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			switch m.state {
			case idle_state:
				m.state = show_options_state
			case show_options_state:
				selected_item, ok := m.list.SelectedItem().(item)
				if ok {
					m.choice = string(selected_item)
					m.state = show_filepicker_state
				}
				return m, nil
			}
		}
	case clear_error_msg:
		m.err = nil
	}

	// TODO: No files are shown
	if m.choice != "" {
		m.filepicker, cmd = m.filepicker.Update(msg)
		cmds = append(cmds, cmd)

		if did_select, path := m.filepicker.DidSelectFile(msg); did_select {
			m.selected_file = path
		}

		if did_select, path := m.filepicker.DidSelectDisabledFile(msg); did_select {
			m.err = fmt.Errorf("file %s is disabled", path)
			m.selected_file = ""
			// return m, tea.Batch(cmd, clear_error_after(2*time.Second))
			cmds = append(cmds, clear_error_after(2*time.Second))
		}
	} else {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
