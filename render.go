package main

import (
	_ "embed"
	"log"
	"math"
	"math/rand/v2"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	//go:embed sounds/click.wav
	clickSoundWav []byte
	clickSound    rl.Sound
)

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

func drawAnimePie(animes []AnimeResult, rotation float32, offset rl.Vector2) {
	rotation = 360 - float32(math.Mod(float64(rotation), 360))

	circleDiameter := float32(rl.GetScreenWidth()) - 50
	circleRadius := circleDiameter * 0.5

	screenCenter := rl.Vector2{
		X: float32(rl.GetScreenWidth()) * 0.5,
		Y: float32(rl.GetScreenHeight()) * 0.5,
	}

	screenCenter = rl.Vector2Add(screenCenter, offset)

	// pie shadow
	dimmer := rl.RayWhite
	dimmer.A = 32
	rl.DrawCircleSector(
		screenCenter, float32(rl.GetScreenWidth())*0.5, 0, 360, 512, dimmer,
	)

	lastAngle := -90 + rotation

	var animeTextRotations []float32

	for _, anime := range animes {
		var animeStartAngle float32 = lastAngle
		var animeEndAngle float32 = animeStartAngle + (360.0 * anime.Weight)

		// seperate sectors by episodes left

		angleWidth := animeEndAngle - animeStartAngle
		epAngleWidth := angleWidth / float32(anime.EpisodesLeft)

		animeColor := hexStrColor(anime.Color)
		animeAltColor := rl.ColorBrightness(animeColor, -0.05)

		for i := range anime.EpisodesLeft {
			var epStartAngle float32 = animeStartAngle + (epAngleWidth * float32(i))
			var epEndAngle float32 = epStartAngle + epAngleWidth

			epColor := animeColor
			if i%2 == 1 {
				epColor = animeAltColor
			}

			rl.DrawCircleSector(
				screenCenter, circleRadius,
				epStartAngle, epEndAngle,
				16, epColor,
			)
		}

		// add text rotation

		animeTextRotation := animeStartAngle + (angleWidth * 0.5)
		animeTextRotations = append(animeTextRotations, animeTextRotation)

		lastAngle = animeEndAngle
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

func playClickSound() {
	var pitchRange float32 = 0.3
	var minPitch float32 = 1 - pitchRange
	var maxPitch float32 = 1 + pitchRange
	pitch := minPitch + rand.Float32()*(maxPitch-minPitch)
	rl.SetSoundPitch(clickSound, pitch)
	rl.PlaySound(clickSound)
}

func runRaylibProgram(animes []AnimeResult, noSpin bool) {
	var windowSize int32 = 800

	rl.SetConfigFlags(rl.FlagMsaa4xHint)
	rl.SetTraceLogLevel(rl.LogNone)

	rl.InitWindow(windowSize, windowSize, "AniList Spinner")
	defer rl.CloseWindow()

	rl.InitAudioDevice()
	defer rl.CloseAudioDevice()

	rl.SetTargetFPS(-1)

	{
		clickSoundWave := rl.LoadWaveFromMemory(
			".wav", clickSoundWav, int32(len(clickSoundWav)),
		)
		defer rl.UnloadWave(clickSoundWave)
		clickSound = rl.LoadSoundFromWave(clickSoundWave)
	}
	defer rl.UnloadSound(clickSound)

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
		rotation = 0 // cause we were idle rotating
		targetRotation = rotation + (360 * float32(randFloat(20, 30)))
		spinning = true
	}

	const rotationSharpness float32 = 0.5

	var lastAnime AnimeResult = animes[0]

	cameraShake, err := NewOneShotShake()
	if err != nil {
		panic(err)
	}

	cameraShake.Amount = 16
	cameraShake.Speed = 16
	cameraShake.Duration = 0.2

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()

		// update rotation

		if !spinning && !noSpin {
			// idle spinning
			rotation = float32(rl.GetTime() * 10)
		}

		if rotation != targetRotation {
			rotation = lerpF32(
				rotation, targetRotation, rotationSharpness*rl.GetFrameTime(),
			)
		}

		// draw

		screenWidth := float32(rl.GetScreenWidth())
		screenHeight := float32(rl.GetScreenHeight())

		rl.ClearBackground(rl.Black)

		drawAnimePie(animes, rotation, cameraShake.GetPosOnFrame())

		// return early
		if noSpin {
			rl.EndDrawing()
			continue
		}

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

		if lastAnime != currentAnime {
			playClickSound()
			cameraShake.DoShake()
			lastAnime = currentAnime
		}

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
