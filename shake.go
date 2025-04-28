package main

import (
	"time"

	"github.com/KEINOS/go-noise"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type OneShotShake struct {
	Noise    noise.Generator
	Amount   float64
	Speed    float64
	Duration float64

	// when started
	CurrentTimestamp float64
}

func NewOneShotShake() (OneShotShake, error) {
	var shake OneShotShake

	var err error
	shake.Noise, err = noise.New(noise.OpenSimplex, time.Now().Unix())
	if err != nil {
		return shake, err
	}

	shake.CurrentTimestamp = -1

	return shake, nil
}

func (shake *OneShotShake) DoShake() {
	shake.CurrentTimestamp = rl.GetTime()
}

func (shake *OneShotShake) GetPosOnFrame() rl.Vector2 {
	time := rl.GetTime()

	t := time * float64(shake.Speed)
	x := shake.Noise.Eval64(t)
	y := shake.Noise.Eval64(1234.567 + t)

	var intensity float64

	if shake.CurrentTimestamp >= 0 {
		if time >= shake.CurrentTimestamp+shake.Duration {
			shake.CurrentTimestamp = -1
		} else {
			// could add easing
			intensity = invLerpF64(
				shake.CurrentTimestamp+shake.Duration,
				shake.CurrentTimestamp,
				time,
			)
		}
	}

	amount := shake.Amount * intensity

	return rl.Vector2{
		X: float32(x * amount),
		Y: float32(y * amount),
	}
}
