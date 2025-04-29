package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func getAnimeData() []AnimeResult {
	args := flag.Args()

	if len(args) <= 1 {
		flag.CommandLine.Usage()
		os.Exit(0)
	}

	username := args[0]
	animeIds := args[1:]

	var animes []AnimeResult

	for _, idStr := range animeIds {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			log.Fatalln(err)
		}

		res, err := getAnime(username, id)
		if err != nil {
			log.Fatalln(err)
		}

		animes = append(animes, res)
	}

	return animes
}

func getTestAnimeData() []AnimeResult {
	data := []AnimeResult{
		{Title: "anime a",
			Progress: 4, Episodes: 23, Duration: 19, Color: "#ff0000"},
		{Title: "anime b: is really damn cool",
			Progress: 5, Episodes: 20, Duration: 21, Color: "#00ff00"},
		{Title: "anime c - isnt it!!?",
			Progress: 6, Episodes: 22, Duration: 22, Color: "#0000ff"},
		{Title: "anime d: the long lived gopher and its legacy",
			Progress: 7, Episodes: 20, Duration: 30, Color: "#ff00ff"},
	}

	for i := range data {
		anime := data[i]
		data[i].EpisodesLeft = anime.Episodes - anime.Progress
		data[i].MinutesLeft = anime.Duration * data[i].EpisodesLeft
	}

	return data
}

var titleShortenerRegExp = regexp.MustCompile("^(.+)[:-]")

var _, useTestData = os.LookupEnv("TEST_DATA")

type ProgramOptions struct {
	NoSpin  bool
	NoShake bool
	NoSound bool
	Volume  float64
}

func main() {
	flag.CommandLine.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"  [options] <anilist username> <anilist ids...>\n",
		)
		fmt.Fprintln(flag.CommandLine.Output(), "\nOptions:")
		flag.PrintDefaults()
	}

	var options ProgramOptions

	flag.BoolVar(&options.NoSpin, "no-spin", false, "disable spinning, hides button")
	flag.BoolVar(&options.NoShake, "no-shake", false, "disable screen shake")
	flag.BoolVar(&options.NoSound, "no-sound", false, "disable sounds")
	flag.Float64Var(&options.Volume, "volume", 0.8, "between 0 to 1")
	flag.Parse()

	options.Volume = clamp(options.Volume, 0, 1)

	var animes []AnimeResult

	if useTestData {
		animes = getTestAnimeData()
	} else {
		animes = getAnimeData()
	}

	// set weights

	var totalMinutesLeft int

	var totalEpisodesWatched int
	var totalEpisodesLeft int

	for _, anime := range animes {
		totalMinutesLeft += anime.MinutesLeft
		totalEpisodesWatched += anime.Progress
		totalEpisodesLeft += anime.EpisodesLeft
	}

	for i, anime := range animes {
		animes[i].Weight = float32(anime.MinutesLeft) / float32(totalMinutesLeft)
	}

	// shorten titles

	for i, anime := range animes {
		matches := titleShortenerRegExp.FindStringSubmatch(anime.Title)
		if len(matches) == 0 {
			continue
		}

		animes[i].Title = strings.TrimSpace(matches[1])
	}

	// render table

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight, AlignHeader: text.AlignCenter},
		{Number: 2, Align: text.AlignRight, AlignHeader: text.AlignCenter},
	})
	t.AppendHeader(table.Row{"Progress", "Duration", "Left", "Title"})
	for _, anime := range animes {
		t.AppendRow(table.Row{
			fmt.Sprintf("%d / %d", anime.Progress, anime.Episodes),
			fmt.Sprintf("%d min", anime.Duration),
			formatMinutes(anime.MinutesLeft),
			anime.Title,
		})
	}
	t.Render()

	t = table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"Progress", "Left"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Align: text.AlignRight, AlignHeader: text.AlignCenter},
	})
	t.AppendRow(table.Row{
		fmt.Sprintf("%d / %d", totalEpisodesWatched, totalEpisodesLeft),
		formatMinutes(totalMinutesLeft),
	})
	t.Render()

	// run raylib program

	runRaylibProgram(animes, options)
}
