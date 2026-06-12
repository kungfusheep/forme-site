// captures one full oscillator cycle of a condensed oscdemo view as
// frame-by-frame styled terminal lines for the site's animated canvas.
// all frequencies are multiples of 0.5hz so the 2s window loops seamlessly.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	. "github.com/kungfusheep/glyph"
)

const (
	width  = 60
	height = 10
	fps    = 15
	cycle  = 2 * time.Second
)

type termAnim struct {
	W      int        `json:"w"`
	H      int        `json:"h"`
	Fps    int        `json:"fps"`
	Frames [][]string `json:"frames"`
}

func main() {
	dim, cyan, green, amber, red := Hex(0x4A4A4A), Hex(0x4DD0E1), Hex(0x66BB6A), Hex(0xD4B87A), Hex(0xD47080)

	// the tree below is reproduced verbatim on the homepage Live motion card;
	// keep the two in lockstep
	wave := make([]Component, 0, 58)
	for i := 0; i < 58; i++ {
		wave = append(wave, Text("█").FG(
			Osc(0.5).Sine().Phase(float64(i)/58).Lerp(dim, cyan)))
	}

	tree := VBox.Gap(1).PaddingVH(0, 1)(
		HBox(wave...),
		HBox.Width(Osc(0.5).Sine().Range(1, 58)).Height(1).Fill(green)(),
		HBox.Width(Osc(0.5).Saw().Range(1, 58)).Height(1).Fill(amber)(),
		HBox.Gap(2)(
			Text("●").FG(Osc(1).Square(0.5).Lerp(dim, red)),
			Text("●").FG(Osc(0.5).Sine().Lerp(dim, green)),
		),
	)

	buf := NewBuffer(width, height)
	tmpl := Build(tree)

	frames := make([][]string, 0, fps*2)
	step := time.Second / fps
	start := time.Now()
	for i := 0; i < int(cycle/step); i++ {
		target := start.Add(time.Duration(i) * step)
		if d := time.Until(target); d > 0 {
			time.Sleep(d)
		}
		buf.Clear()
		tmpl.Execute(buf, width, height)
		lines := make([]string, 0, height)
		for y := 0; y < height; y++ {
			lines = append(lines, buf.GetLineStyled(y))
		}
		frames = append(frames, lines)
	}

	out, err := json.Marshal(termAnim{W: width, H: height, Fps: fps, Frames: frames})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout.Write(out)
}
