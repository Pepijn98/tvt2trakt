package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"

	"github.com/jszwec/csvutil"
)

type TvShow struct {
	CreatedAt           string `csv:"created_at"`
	TvShowName          string `csv:"tv_show_name"`
	EpisodeSeasonNumber int    `csv:"episode_season_number"`
	EpisodeNumber       int    `csv:"episode_number"`
	EpisodeID           string `csv:"episode_id"`
	UpdatedAt           string `csv:"updated_at"`
}

func main() {
	csv_file, _ := os.Open("./data/seen_episode_NoAnimeVer_2.csv")
	reader := csv.NewReader(csv_file)
	reader.Comma = ','

	headers, _ := csvutil.Header(TvShow{}, "csv")
	dec, _ := csvutil.NewDecoder(reader, headers...)

	var shows []TvShow
	for {
		var show TvShow
		if err := dec.Decode(&show); err == io.EOF {
			break
		}
		shows = append(shows, show)
	}
	shows = shows[1:]

	fmt.Println(shows[0])
}
