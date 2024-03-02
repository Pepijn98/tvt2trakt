package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"tvt2trakt/internal/tui"

	"github.com/BurntSushi/toml"
	tea "github.com/charmbracelet/bubbletea"
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

	model := tui.New()

	if _, err := tea.NewProgram(&model, tea.WithOutput(os.Stderr)).Run(); err != nil {
		fmt.Println("could not start program:", err)
	}
}

// TODO: Implement properly (playing last epoch now fuck it)
//
//lint:ignore U1000 Ignore unused function temporarily
func load_file() {
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
