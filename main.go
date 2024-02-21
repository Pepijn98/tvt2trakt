package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/jszwec/csvutil"
)

type Config struct {
	ClientID      string `toml:"client_id"`
	ClientSecret  string `toml:"client_secret"`
	TraktUsername string `toml:"trakt_username"`
}

type TvShow struct {
	CreatedAt           string `csv:"created_at"`
	TvShowName          string `csv:"tv_show_name"`
	EpisodeSeasonNumber string `csv:"episode_season_number"`
	EpisodeNumber       string `csv:"episode_number"`
	EpisodeID           string `csv:"episode_id"`
	UpdatedAt           string `csv:"updated_at"`
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

	csv_file, err := os.Open("./data/seen_episode_NoAnimeVer_2.csv")
	if err != nil {
		// CSV file can't be read
		log.Fatal(err)
	}

	reader := csv.NewReader(csv_file)
	reader.Comma = ','

	headers, err := csvutil.Header(TvShow{}, "csv")
	if err != nil {
		// TODO: Handle error properly
		fmt.Println(err)
	}

	dec, _ := csvutil.NewDecoder(reader, headers...)
	if err != nil {
		// TODO: Handle error properly
		fmt.Println(err)
	}

	var shows []TvShow
	for {
		var show TvShow
		if err := dec.Decode(&show); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}

		// Don't append csv headers
		if show.TvShowName == "tv_show_name" {
			continue
		}

		shows = append(shows, show)
	}

	fmt.Println(shows[0])
}
