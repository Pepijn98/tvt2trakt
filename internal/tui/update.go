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
		next bool
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
		m, cmd, next = update_splash(m, msg)
		if next {
			m.state = options_state
		}
	case options_state:
		m, cmd, next = update_options(m, msg)
		if next {
			m.state = filepicker_state
		}
	case filepicker_state:
		m, cmd, _ = update_filepicker(m, msg)
	}

	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// TODO: Separate keymsg logic from update logic
// Check key input per state
// Update last

/* EXAMPLE FOR TOMORROW SO I DON'T FORGET

switch msg := msg.(type) {
case tea.KeyMsg:
	if m.state == STATE {
		switch msg.String() {
		case "KEY":
			// keypress
		}
	}

	if m.state == STATE {
		switch msg.String() {
		case "KEY":
			// keypress
		}
	}
}

switch m.state {
case STATE:
	// UPDATE
case STATE2:
	// Update
}

if cmd != nil {
	cmds = append(cmds, cmd)
}

return m, tea.Batch(cmds...)
*/

func update_splash(m model, msg tea.Msg) (model, tea.Cmd, bool) {
	var next bool = false

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// TODO: Go to login state instead
			next = true
		}
	}

	return m, nil, next
}

func update_options(m model, msg tea.Msg) (model, tea.Cmd, bool) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
		next bool = false
	)

	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

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
				next = true
			}
		}
	}

	return m, tea.Batch(cmds...), next
}

// Cna't get filepicker to properly work, it finally shows files but only shows 1 line
// Maybe try filetree instead
func update_filepicker(m model, msg tea.Msg) (model, tea.Cmd, bool) {
	var (
		cmd  tea.Cmd
		next bool = false
	)

	m.filepicker, cmd = m.filepicker.Update(msg)

	switch msg.(type) {
	case clear_error_msg:
		m.err = nil
	}

	if did_select, path := m.filepicker.DidSelectFile(msg); did_select {
		m.selected_file = path
		next = true
	}

	if did_select, path := m.filepicker.DidSelectDisabledFile(msg); did_select {
		m.err = fmt.Errorf("file %s is disabled", path)
		m.selected_file = ""
		return m, tea.Batch(cmd, clear_error_after(2*time.Second)), next
	}

	return m, cmd, next
}
