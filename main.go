package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jszwec/csvutil"
)

type Config struct {
	ClientID      string `toml:"client_id"`
	ClientSecret  string `toml:"client_secret"`
	TraktUsername string `toml:"trakt_username"`
}

type TvTimeShow struct {
	CreatedAt           string `csv:"created_at"`
	TvShowName          string `csv:"tv_show_name"`
	EpisodeSeasonNumber string `csv:"episode_season_number"`
	EpisodeNumber       string `csv:"episode_number"`
	EpisodeID           string `csv:"episode_id"`
	UpdatedAt           string `csv:"updated_at"`
}

type Episode struct {
	CreatedAt string
	Number    int
	ID        string
	UpdatedAt string
}

type Season struct {
	Number   int
	Episodes []Episode
}

type Show struct {
	Name    string
	Seasons []Season
}

const listHeight = 14

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int {
	return 1
}
func (d itemDelegate) Spacing() int {
	return 0
}
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type model struct {
	List         list.Model
	Choice       string
	FilePicker   filepicker.Model
	SelectedFile string
	Quitting     bool
	Err          error
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}

func (m model) Init() tea.Cmd {
	return m.FilePicker.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.List.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "esc", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "enter":
			selected_item, ok := m.List.SelectedItem().(item)
			if ok {
				m.Choice = string(selected_item)
			}
			return m, tea.Quit
		}
	case clearErrorMsg:
		m.Err = nil
	}

	var cmd tea.Cmd

	if m.Choice != "" {
		m.FilePicker, cmd = m.FilePicker.Update(msg)
		// Did the user select a file?
		if didSelect, path := m.FilePicker.DidSelectFile(msg); didSelect {
			// Get the path of the selected file.
			m.SelectedFile = path
		}

		// Did the user select a disabled file?
		// This is only necessary to display an error to the user.
		if didSelect, path := m.FilePicker.DidSelectDisabledFile(msg); didSelect {
			// Let's clear the selectedFile and display an error.
			m.Err = errors.New(path + " is not valid.")
			m.SelectedFile = ""
			return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
		}
	} else {
		m.List, cmd = m.List.Update(msg)
	}

	return m, nil
}

func (m model) View() string {
	if m.Quitting {
		return "\n  See you later!\n\n"
	}

	if m.Choice != "" {
		var s strings.Builder
		s.WriteString("\n  ")
		if m.Err != nil {
			s.WriteString(m.FilePicker.Styles.DisabledFile.Render(m.Err.Error()))
		} else if m.SelectedFile == "" {
			s.WriteString("Pick a file:")
		} else {
			s.WriteString("Selected file: " + m.FilePicker.Styles.Selected.Render(m.SelectedFile))
		}
		s.WriteString("\n\n" + m.FilePicker.View() + "\n")
		return s.String()
	} else {
		return "\n" + m.List.View()
	}
}

func main() {
	var config Config
	conf_file, err := os.ReadFile("./config.toml")
	if err != nil {
		// Config file can't be read
		log.Fatal(err)
	}

	_, err = toml.Decode(string(conf_file), &config)
	if err != nil {
		// Invalid toml
		log.Fatal(err)
	}

	loadFile()

	items := []list.Item{
		item("Watched Shows"),
		item("Followed Shows"),
		item("Movies"),
	}

	const defaultWidth = 20

	options := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	options.Title = "What do you want to import?"
	options.SetShowStatusBar(false)
	options.SetFilteringEnabled(false)
	options.Styles.Title = titleStyle
	options.Styles.PaginationStyle = paginationStyle
	options.Styles.HelpStyle = helpStyle

	fp := filepicker.New()
	fp.AllowedTypes = []string{".csv", ".md", ".go"}
	fp.CurrentDirectory, _ = os.UserHomeDir()

	initial_model := model{
		List:         options,
		Choice:       "",
		FilePicker:   fp,
		SelectedFile: "",
		Quitting:     false,
	}

	app := tea.NewProgram(&initial_model, tea.WithOutput(os.Stderr))
	if _, err := app.Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

func loadFile() {
	csv_file, err := os.Open("./data/seen_episode_NoAnimeVer_2.csv")
	if err != nil {
		// CSV file can't be read
		log.Fatal(err)
	}

	reader := csv.NewReader(csv_file)
	reader.Comma = ','

	headers, err := csvutil.Header(TvTimeShow{}, "csv")
	if err != nil {
		// TODO: Handle error properly
		fmt.Println(err)
	}

	dec, _ := csvutil.NewDecoder(reader, headers...)
	if err != nil {
		// TODO: Handle error properly
		fmt.Println(err)
	}

	var shows []Show
	for {
		var tvt_show TvTimeShow
		if err := dec.Decode(&tvt_show); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		// Don't append csv headers
		if tvt_show.TvShowName == "tv_show_name" {
			continue
		}

		// Probably not the most optimal but a quick way to have the shows in a proper data structure
		// Each show has a seasons array and each season has an episodes array
		// In my head this seems the most logical way to do it right now but trakt api might operate different I don't know yet
		idx := slices.IndexFunc(shows, func(s Show) bool { return s.Name == tvt_show.TvShowName })
		if idx == -1 {
			episode_num, err := strconv.Atoi(tvt_show.EpisodeNumber)
			if err != nil {
				log.Fatal(err)
			}

			episode := Episode{
				CreatedAt: tvt_show.CreatedAt,
				Number:    episode_num,
				ID:        tvt_show.EpisodeID,
				UpdatedAt: tvt_show.UpdatedAt,
			}

			season_num, err := strconv.Atoi(tvt_show.EpisodeSeasonNumber)
			if err != nil {
				log.Fatal(err)
			}

			season := Season{
				Number:   season_num,
				Episodes: []Episode{episode},
			}

			show := Show{
				Name:    tvt_show.TvShowName,
				Seasons: []Season{season},
			}

			shows = append(shows, show)
		} else {
			show := &shows[idx]

			episode_num, err := strconv.Atoi(tvt_show.EpisodeNumber)
			if err != nil {
				log.Fatal(err)
			}

			episode := Episode{
				CreatedAt: tvt_show.CreatedAt,
				Number:    episode_num,
				ID:        tvt_show.EpisodeID,
				UpdatedAt: tvt_show.UpdatedAt,
			}

			season_num, err := strconv.Atoi(tvt_show.EpisodeSeasonNumber)
			if err != nil {
				log.Fatal(err)
			}

			idx := slices.IndexFunc(show.Seasons, func(s Season) bool { return s.Number == season_num })
			if idx == -1 {
				season := Season{
					Number:   season_num,
					Episodes: []Episode{episode},
				}
				show.Seasons = append(show.Seasons, season)
			} else {
				episodes := show.Seasons[idx].Episodes
				episodes = append(episodes, episode)
				season := Season{
					Number:   season_num,
					Episodes: episodes,
				}
				show.Seasons[idx] = season
			}
		}
	}
}
