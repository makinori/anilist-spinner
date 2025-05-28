package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// https://studio.apollographql.com/sandbox/explorer

const searchQuery = `
query($userName: String, $mediaId: Int)  {
  MediaList(userName: $userName, mediaId: $mediaId) {
    progress
  }
  Media(id: $mediaId) {
    episodes
    title {
      english
      romaji
    }
	synonyms
	bannerImage
	coverImage {
	  color
	  extraLarge
	}
	duration
  }
}`

type SearchQueryVars struct {
	UserName string `json:"userName"`
	MediaID  int    `json:"mediaId"`
}

type SearchQuery struct {
	Query     string          `json:"query"`
	Variables SearchQueryVars `json:"variables"`
}

type SearchQueryResult struct {
	Errors []struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"errors,omitempty"`
	Data struct {
		MediaList struct {
			Progress int `json:"progress"`
		} `json:"MediaList"`
		Media struct {
			Episodes int `json:"episodes"`
			Title    struct {
				English string `json:"english"`
				Romaji  string `json:"romaji"`
			} `json:"title"`
			Synonyms    []string `json:"synonyms"`
			BannerImage string   `json:"bannerImage"`
			CoverImage  struct {
				Color      string `json:"color"`
				ExtraLarge string `json:"extraLarge"`
			} `json:"coverImage"`
			Duration int `json:"duration"`
		} `json:"Media"`
	} `json:"data"`
}

type AnimeResult struct {
	Progress int
	Episodes int
	Title    string
	Color    string
	Duration int
	// helpers
	EpisodesLeft int
	MinutesLeft  int
	Weight       float32
}

func getAnime(username string, id int) (AnimeResult, error) {
	searchQueryJson, err := json.Marshal(SearchQuery{
		Query: searchQuery,
		Variables: SearchQueryVars{
			UserName: username,
			MediaID:  id,
		},
	})

	if err != nil {
		return AnimeResult{}, err
	}

	req, err := http.NewRequest(
		"POST", "https://graphql.anilist.co",
		bytes.NewBuffer(searchQueryJson),
	)

	if err != nil {
		return AnimeResult{}, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return AnimeResult{}, err
	}

	defer res.Body.Close()

	resBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return AnimeResult{}, err
	}

	var result SearchQueryResult
	err = json.Unmarshal(resBytes, &result)
	if err != nil {
		return AnimeResult{}, err
	}

	if len(result.Errors) > 0 {
		return AnimeResult{}, fmt.Errorf("%s", result.Errors)
	}

	title := result.Data.Media.Title.English
	if title == "" {
		if len(result.Data.Media.Synonyms) > 0 {
			title = result.Data.Media.Synonyms[0]
		} else {
			title = result.Data.Media.Title.Romaji
		}
	}

	anime := AnimeResult{
		Progress: result.Data.MediaList.Progress,
		Episodes: result.Data.Media.Episodes,
		Title:    title,
		Color:    result.Data.Media.CoverImage.Color,
		Duration: result.Data.Media.Duration,
	}

	anime.EpisodesLeft = anime.Episodes - anime.Progress
	anime.MinutesLeft = anime.Duration * anime.EpisodesLeft

	return anime, nil
}
