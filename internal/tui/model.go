package tui

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type session_state int

const (
	splash_state     session_state = iota // initial state
	login_state                           // login to trakt using oauth
	options_state                         // choose what to import
	filepicker_state                      // choose csv file to import
	uploading_state                       // importing data to trakt (show progress bar)
	exit_state                            // exiting the program
)

var (
	title_style         = lipgloss.NewStyle().MarginLeft(2)
	item_style          = lipgloss.NewStyle().PaddingLeft(4)
	selected_item_style = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	pagination_style    = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	help_style          = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1).Foreground(lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#5C5C5C"})
)

type item string

func (i item) FilterValue() string { return "" }

type item_delegate struct{}

func (d item_delegate) Height() int {
	return 1
}

func (d item_delegate) Spacing() int {
	return 0
}

func (d item_delegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d item_delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := item_style.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selected_item_style.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	list          list.Model
	filepicker    filepicker.Model
	choice        string
	selected_file string
	state         session_state
	err           error
}

func New() model {
	items := []list.Item{
		item("Watched Shows"),
		item("Followed Shows"),
		item("Movies"),
	}

	options := list.New(items, item_delegate{}, 20, 14)
	options.Title = "What do you want to import?"
	options.SetShowStatusBar(false)
	options.SetFilteringEnabled(false)
	options.Styles.Title = title_style
	options.Styles.PaginationStyle = pagination_style
	options.Styles.HelpStyle = help_style

	fp_model := filepicker.New()
	fp_model.AllowedTypes = []string{".csv"}
	fp_model.DirAllowed = true
	fp_model.ShowHidden = false
	fp_model.Height = 8

	fp_model.KeyMap = filepicker.KeyMap{
		GoToTop:  key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "first")),
		GoToLast: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "last")),
		Down:     key.NewBinding(key.WithKeys("j", "down", "ctrl+n"), key.WithHelp("↓/j", "down")),
		Up:       key.NewBinding(key.WithKeys("k", "up", "ctrl+p"), key.WithHelp("↑/k", "up")),
		PageUp:   key.NewBinding(key.WithKeys("K", "pgup"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("J", "pgdown"), key.WithHelp("pgdown", "page down")),
		Back:     key.NewBinding(key.WithKeys("h", "backspace", "left"), key.WithHelp("←/h", "back")),
		Open:     key.NewBinding(key.WithKeys("l", "right", "enter"), key.WithHelp("→/l", "open")),
		Select:   key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	}

	dir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	// If the data directory exists, use it else use the user's home directory
	data_dir := filepath.Join(dir, "data")
	if _, err := os.Stat(data_dir); !os.IsNotExist(err) {
		dir = data_dir
	} else {
		dir, err = os.UserHomeDir()
		if err != nil {
			log.Panic(err)
		}
	}

	fp_model.CurrentDirectory = dir

	return model{
		list:          options,
		filepicker:    fp_model,
		choice:        "",
		selected_file: "",
		state:         splash_state,
		err:           nil,
	}
}
