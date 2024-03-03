package tui

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type session_state int

const (
	splash_state session_state = iota
	login_state
	options_state
	filepicker_state
	exit_state
)

var (
	title_style         = lipgloss.NewStyle().MarginLeft(2)
	item_style          = lipgloss.NewStyle().PaddingLeft(4)
	selected_item_style = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	pagination_style    = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	help_style          = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
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
	fp_model.ShowHidden = true
	fp_model.Height = 10 // ALL I HAD TO DO TO SHOW MORE THAN 1 FILE WAS TO ADD HEIGHT WOOOOOOOOOOOOOOOOOOOOOOOOOW (totally didn't waste 2 days haha.. hahahahaha ヽ(｀Д´#)ﾉ)

	dir, err := os.Getwd()
	if err != nil {
		log.Panic(err)
	}

	fp_model.CurrentDirectory = dir + "/data"

	return model{
		list:          options,
		filepicker:    fp_model,
		choice:        "",
		selected_file: "",
		state:         splash_state,
		err:           nil,
	}
}
