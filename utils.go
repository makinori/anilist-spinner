package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"regexp"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
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

func lerpF32(a, b, t float32) float32 {
	return a + (b-a)*t
}

func lerpF64(a, b, t float64) float64 {
	return a + (b-a)*t
}

func invLerpF64(a, b, v float64) float64 {
	return (v - a) / (b - a)
}

func randFloat(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func clamp(v, min, max float64) float64 {
	return math.Max(math.Min(v, max), min)
}
