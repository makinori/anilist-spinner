package main

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func formatMinutes(m int) string {
	if m < 60 {
		return fmt.Sprintf("%dm", m)
	}
	h := m / 60
	m = m % 60
	return fmt.Sprintf("%dh %dm", h, m)
}

var hexColorRegExp = regexp.MustCompile("(?i)#?([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})?")

func hexStrColor(hexStr string) color.RGBA {
	matches := hexColorRegExp.FindStringSubmatch(hexStr)
	if len(matches) == 0 {
		return rl.Black
	}

	r, _ := strconv.ParseInt(matches[1], 16, 16)
	g, _ := strconv.ParseInt(matches[2], 16, 16)
	b, _ := strconv.ParseInt(matches[3], 16, 16)
	a, err := strconv.ParseInt(matches[4], 16, 16)
	if err != nil {
		a = 255
	}

	return rl.Color{
		R: uint8(r),
		G: uint8(g),
		B: uint8(b),
		A: uint8(a),
	}
}

func getAnimeData() []AnimeResult {
	if len(os.Args) <= 1 {
		fmt.Println("usage: <anilist username> <anilist id...>")
		os.Exit(1)
	}

	username := os.Args[1]
	animeIds := os.Args[2:]

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
	return []AnimeResult{
		{Title: "anime a",
			Progress: 4, Episodes: 24, Color: "#ff0000",
			Duration: 20, EpisodesLeft: 20, MinutesLeft: 400},
		{Title: "anime b: is really damn cool",
			Progress: 5, Episodes: 24, Color: "#00ff00",
			Duration: 20, EpisodesLeft: 18, MinutesLeft: 360},
		{Title: "anime c - isnt it!!?",
			Progress: 6, Episodes: 24, Color: "#0000ff",
			Duration: 20, EpisodesLeft: 16, MinutesLeft: 320},
		{Title: "anime d: the long lived gopher and its legacy",
			Progress: 7, Episodes: 24, Color: "#ff00ff",
			Duration: 20, EpisodesLeft: 14, MinutesLeft: 280},
	}
}

func fitTextToWidth(
	text string, fontSize float32, spacing float32, maxWidth float32,
) (float32, float32, rl.Vector2) {
	inputFontSize := fontSize
	inputSpacing := spacing

	var textSize rl.Vector2

	textSize = rl.MeasureTextEx(
		rl.GetFontDefault(), text, fontSize, spacing,
	)

	for textSize.X > maxWidth {
		fontSize -= 1
		spacing = inputSpacing * (fontSize / inputFontSize)

		textSize = rl.MeasureTextEx(
			rl.GetFontDefault(), text, fontSize, spacing,
		)
	}

	return fontSize, spacing, textSize
}

func drawAnimeCircle(animes []AnimeResult, rotation float32) {
	rotation = 360 - float32(math.Mod(float64(rotation), 360))

	circleDiameter := float32(rl.GetScreenWidth()) - 50
	circleRadius := circleDiameter * 0.5

	screenCenter := rl.Vector2{
		X: float32(rl.GetScreenWidth()) * 0.5,
		Y: float32(rl.GetScreenHeight()) * 0.5,
	}

	lastAngle := -90 + rotation

	var animeTextRotations []float32

	for _, anime := range animes {
		var endAngle float32 = lastAngle + (360.0 * anime.Weight)

		rl.DrawCircleSector(
			screenCenter, circleRadius,
			lastAngle, endAngle,
			512, hexStrColor(anime.Color),
		)

		animeTextRotation := lastAngle + ((endAngle - lastAngle) * 0.5)
		animeTextRotations = append(animeTextRotations, animeTextRotation)

		lastAngle = endAngle
	}

	for i, anime := range animes {

		fontSize, spacing, textSize := fitTextToWidth(
			anime.Title, 64, 8, circleRadius-50,
		)

		origin := rl.Vector2{
			X: (textSize.X * 0.5) - (circleRadius * 0.5),
			Y: textSize.Y * 0.5,
		}

		rl.DrawTextPro(
			rl.GetFontDefault(), anime.Title,
			screenCenter,
			origin,
			animeTextRotations[i], // rotation
			fontSize,
			spacing,
			rl.White,
		)
	}
}

func getAnimeFromRotation(animes []AnimeResult, rotation float32) AnimeResult {
	rotationNormalized := float32(math.Mod(float64(rotation), 360)) / 360

	var seenWeight float32 = 0

	for _, anime := range animes {
		seenWeight += anime.Weight
		if rotationNormalized < seenWeight {
			return anime
		}
	}

	log.Fatalln("weights didn't add up properly")

	return AnimeResult{}
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

var titleShortenerRegExp = regexp.MustCompile("^(.+)[:-]")

var _, useTestData = os.LookupEnv("TEST_DATA")

func main() {
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

	var windowSize int32 = 800

	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.SetTraceLogLevel(rl.LogNone)

	rl.InitWindow(windowSize, windowSize, "AniList Spinner")
	defer rl.CloseWindow()

	rl.SetTargetFPS(-1)

	// var lastTitle string
	// updateTitle := func(anime AnimeResult) {
	// 	newTitle := "AniList Spinner - " + anime.Title
	// 	if lastTitle == newTitle {
	// 		return
	// 	}
	// 	rl.SetWindowTitle(newTitle)
	// 	lastTitle = newTitle
	// }

	// state
	var rotation float32 = 0
	var targetRotation float32 = 0

	spinning := false

	doSpin := func() {
		rotation = 0 // cause we were idle spinning
		targetRotation = rotation + (360 * float32(randFloat(20, 30)))
		spinning = true
	}

	const rotationSharpness float32 = 0.5

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		// update rotation

		if !spinning {
			// idle spinning
			rotation = float32(rl.GetTime() * 10)
		}

		if rotation != targetRotation {
			rotation = lerp(
				rotation, targetRotation, rotationSharpness*rl.GetFrameTime(),
			)
		}

		// draw

		screenWidth := float32(rl.GetScreenWidth())
		screenHeight := float32(rl.GetScreenHeight())

		rl.ClearBackground(rl.Black)

		dimmer := rl.RayWhite
		dimmer.A = 32
		rl.DrawCircleSector(
			rl.Vector2{X: screenWidth * 0.5, Y: screenHeight * 0.5},
			screenWidth*0.5, 0, 360, 512, dimmer,
		)

		drawAnimeCircle(animes, rotation)

		// draw triangle
		{
			const triangleWidth float32 = 50
			const triangleHeight float32 = 50
			rl.DrawTriangle(
				rl.Vector2{
					X: (screenWidth + triangleWidth) * 0.5,
					Y: 0,
				},
				rl.Vector2{
					X: (screenWidth - triangleWidth) * 0.5,
					Y: 0,
				},
				rl.Vector2{
					X: screenWidth * 0.5,
					Y: triangleHeight,
				},
				rl.RayWhite,
			)
		}

		currentAnime := getAnimeFromRotation(animes, rotation)
		// updateTitle(currentAnime)

		// draw text
		{
			var padding float32 = 8

			fontSize, spacing, textSize := fitTextToWidth(
				currentAnime.Title, 64, 4, screenWidth-50,
			)

			rl.DrawRectangle(
				0,
				int32(screenHeight-padding*2-textSize.Y),
				int32(screenWidth),
				int32(textSize.Y+padding*2),
				rl.Color{R: 0, G: 0, B: 0, A: 128},
			)

			rl.DrawTextPro(
				rl.GetFontDefault(), currentAnime.Title,
				rl.Vector2{
					X: screenWidth * 0.5,
					Y: screenHeight - padding,
				},
				rl.Vector2{
					X: textSize.X * 0.5,
					Y: textSize.Y,
				},
				0, fontSize, spacing,
				rl.RayWhite,
			)
		}

		if !spinning {
			const circleRadius float32 = 80
			rl.DrawRectangle(
				0, 0, int32(screenWidth), int32(screenHeight),
				rl.Color{R: 0, G: 0, B: 0, A: 128},
			)
			rl.DrawCircle(
				int32(screenWidth*0.5),
				int32(screenHeight*0.5),
				circleRadius, rl.RayWhite,
			)
			dimmer := rl.RayWhite
			dimmer.A = 96
			rl.DrawCircle(
				int32(screenWidth*0.5),
				int32(screenHeight*0.5),
				circleRadius+20, dimmer,
			)
			const fontSize float32 = 64
			const spacing float32 = 4
			textSize := rl.MeasureTextEx(
				rl.GetFontDefault(), "spin", fontSize, spacing,
			)
			rl.DrawTextPro(rl.GetFontDefault(), "spin", rl.Vector2{
				X: screenWidth * 0.5,
				Y: screenHeight * 0.5,
			}, rl.Vector2{
				X: textSize.X * 0.5,
				Y: textSize.Y * 0.5,
			}, 0, fontSize, spacing, rl.Black)

			// check if mouse pressed

			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				mousePos := rl.GetMousePosition()
				d := rl.Vector2Distance(
					mousePos, rl.Vector2{X: screenWidth * 0.5, Y: screenHeight * 0.5},
				)
				if d < circleRadius {
					doSpin()
				}
			}
		}

		rl.EndDrawing()
	}
}
