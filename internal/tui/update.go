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

	if msg, ok := msg.(tea.KeyMsg); ok {
		key := msg.String()
		if key == "q" || key == "ctrl+c" {
			m.state = exit_state
			return m, tea.Quit
		}
	}

	switch m.state {
	case splash_state:
		m, cmd = update_splash(m, msg)
	case options_state:
		m, cmd = update_options(m, msg)
	case filepicker_state:
		m, cmd = update_filepicker(m, msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func update_splash(m model, msg tea.Msg) (model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// TODO: Go to login state instead
			m.state = options_state
		}
	}

	return m, nil
}

func update_options(m model, msg tea.Msg) (model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selected_item, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(selected_item)

				cmd = m.filepicker.Init()
				cmds = append(cmds, cmd)

				m.state = filepicker_state
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Cna't get filepicker to properly work, it finally shows files but only shows 1 line
// Maybe try filetree instead
func update_filepicker(m model, msg tea.Msg) (model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.filepicker.Height = msg.Height
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

	m.filepicker, cmd = m.filepicker.Update(msg)

	return m, cmd
}
