package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"

	. "github.com/kungfusheep/glyph"
)

type example struct {
	view   any
	w, h   int
	pad    int // extra rows below content for fade clearance
}

func examples() map[string]example {
	m := map[string]example{}

	// styled-text
	{
		m["styled-text"] = example{
			w: 30, h: 4,
			view: VBox(
				Text("error").FG(Red).Bold(),
				Text("muted").FG(BrightBlack).Dim(),
				Text("success").FG(Green),
				Text("warning").FG(Yellow).Bold(),
			),
		}
	}

	// theme-demo
	{
		theme := ThemeDark

		m["theme-demo"] = example{
			w: 30, h: 7,
			view: VBox.CascadeStyle(&theme.Base).Border(BorderRounded).BorderFG(theme.Border.FG)(
				Text("normal text"),
				Text("muted").Style(theme.Muted),
				Text("accent").Style(theme.Accent),
				Text("error!").Style(theme.Error),
			),
		}
	}

	// concepts-build
	{
		title := "Hello"
		m["concepts-build"] = example{
			w: 24, h: 3,
			view: VBox(
				Text(&title),
				Text("static label"),
			),
		}
	}

	// concepts-vbox
	{
		m["concepts-vbox"] = example{
			w: 24, h: 6,
			view: VBox.Gap(1)(
				Text("hello"),
				Text("world"),
				Text("!"),
			),
		}
	}

	// concepts-grow1 (header + Grow(1) fills remaining + footer)
	{
		items := []string{"server.go", "handler.go", "config.go"}
		m["concepts-grow1"] = example{
			w: 30, h: 10, pad: 2,
			view: VBox(
				Text("header"),
				VBox.Grow(1).Border(BorderRounded)(
					List(&items),
				),
				Text("footer"),
			),
		}
	}

	// concepts-grow-equal (header + two equal Grow(1))
	{
		items := []string{"server.go", "handler.go"}
		m["concepts-grow-equal"] = example{
			w: 30, h: 12, pad: 2,
			view: VBox(
				Text("header"),
				VBox.Grow(1).Border(BorderRounded).Title("list")(
					List(&items),
				),
				VBox.Grow(1).Border(BorderRounded).Title("log")(
					Text("started on :8080").FG(BrightBlack),
				),
				Text("footer"),
			),
		}
	}

	// concepts-grow-ratio (header + Grow(1) vs Grow(2))
	{
		m["concepts-grow-ratio"] = example{
			w: 40, h: 10, pad: 2,
			view: VBox(
				Text("header"),
				HBox.Grow(1)(
					VBox.Grow(1).Border(BorderRounded).Title("sidebar")(
						Text("nav one"),
						Text("nav two"),
					),
					VBox.Grow(2).Border(BorderRounded).Title("content")(
						Text("main content area"),
						Text("with more detail"),
					),
				),
				Text("footer"),
			),
		}
	}

	// concepts-foreach
	{
		s1, s2, s3 := "running", "healthy", "stopped"
		m["concepts-foreach"] = example{
			w: 30, h: 5,
			view: VBox(
				HBox.Gap(2)(Text("api server"), Text(&s1).FG(
					Switch(&s1).Case("running", Yellow).Case("stopped", Red).Default(BrightBlack),
				)),
				HBox.Gap(2)(Text("database"), Text(&s2).FG(
					Switch(&s2).Case("healthy", Green).Case("stopped", Red).Default(BrightBlack),
				)),
				HBox.Gap(2)(Text("cache"), Text(&s3).FG(
					Switch(&s3).Case("running", Green).Case("healthy", Green).Case("stopped", Red).Default(BrightBlack),
				)),
			),
		}
	}

	// concepts-jump
	{
		m["concepts-jump"] = example{w: 24, h: 7, view: nil}
	}

	// concepts-hbox
	{
		m["concepts-hbox"] = example{
			w: 24, h: 2,
			view: HBox.Gap(2)(
				Text("hello"),
				Text("world"),
				Text("!"),
			),
		}
	}

	// form-demo
	{
		name := "Pete"
		email := "pete@example.com"
		role := 0
		agree := true
		register := func() {}

		m["form-demo"] = example{
			w: 36, h: 9,
			view: VBox.Border(BorderRounded).Title("register")(
				Form.LabelBold().OnSubmit(register)(
					Field("Name", Input(&name)),
					Field("Email", Input(&email)),
					Field("Role", Radio(&role, "Admin", "User", "Guest")),
					Field("Terms", Checkbox(&agree, "I accept")),
				),
			),
		}
	}

	// table-demo
	{
		type svc struct {
			Name   string
			Status string
			CPU    string
		}
		services := []svc{
			{"api", "running", "12%"},
			{"worker", "running", "8%"},
			{"cache", "stopped", "0%"},
			{"db", "running", "23%"},
		}

		m["table-demo"] = example{
			w: 40, h: 7,
			view: AutoTable(&services).Scrollable(10),
		}
	}

	// control-flow
	{
		online := true
		mode := "edit"

		m["control-flow"] = example{
			w: 30, h: 5,
			view: VBox.Border(BorderRounded).Title("status")(
				If(&online).Then(Text("● connected").FG(Green)).Else(Text("● offline").FG(Red)),
				Switch(&mode).
					Case("edit", Text("mode: edit").FG(Yellow)).
					Case("preview", Text("mode: preview").FG(Cyan)).
					Default(Text("mode: idle").FG(BrightBlack)),
			),
		}
	}

	// file-picker
	{
		files := []string{"main.go", "go.mod", "README.md", "config.toml", "Makefile"}

		m["file-picker"] = example{
			w: 24, h: 7,
			view: VBox.Border(BorderRounded).Title("open")(
				List(&files).BindVimNav(),
			),
		}
	}

	// layout-statusbar
	{
		m["layout-statusbar"] = example{
			w: 40, h: 2,
			view: HBox.Gap(2)(
				Text("ready").FG(Green),
				Text("3 tasks").Bold(),
				Space(),
				Text("q: quit").FG(BrightBlack),
			),
		}
	}

	// layout-arrange
	{
		grid := Arrange(func(children []ChildSize, w, h int) []Rect {
			cols := 3
			cellW := w / cols
			cellH := 3
			rects := make([]Rect, len(children))
			for i := range children {
				rects[i] = Rect{
					X: (i % cols) * cellW,
					Y: (i / cols) * cellH,
					W: cellW,
					H: cellH,
				}
			}
			return rects
		})

		m["layout-arrange"] = example{
			w: 42, h: 7,
			view: grid(
				VBox.Border(BorderRounded)(Text("alpha")),
				VBox.Border(BorderRounded)(Text("beta")),
				VBox.Border(BorderRounded)(Text("gamma")),
				VBox.Border(BorderRounded)(Text("delta")),
				VBox.Border(BorderRounded)(Text("epsilon")),
				VBox.Border(BorderRounded)(Text("zeta")),
			),
		}
	}

	// layout-compose (widget demo)
	{
		m["layout-compose"] = example{
			w: 50, h: 10,
			view: VBox.Border(BorderRounded).Title("dashboard")(
				HBox.Gap(2)(
					Text("metrics").Bold(),
					Space(),
					Text("live").FG(Green),
				),
				HRule(),
				HBox(
					VBox.Grow(1).Border(BorderRounded).Title("chart")(
						Widget(
							func(availW int16) (w, h int16) { return availW, 4 },
							func(buf *Buffer, x, y, w, h int16) {
								bars := []int{3, 1, 4, 2, 3, 4, 2, 1, 3, 2}
								barW := int(w) / len(bars)
								for i, v := range bars {
									for row := 0; row < v; row++ {
										bx := int(x) + i*barW
										by := int(y) + int(h) - 1 - row
										for dx := 0; dx < barW-1; dx++ {
											buf.Set(bx+dx, by, Cell{Rune: '█', Style: Style{FG: Green}})
										}
									}
								}
							},
						),
					),
					VBox.Grow(1).Border(BorderRounded).Title("status")(
						Leader("uptime", "99.9%"),
						Leader("errors", "0"),
						Leader("latency", "12ms"),
					),
				),
			),
		}
	}

	// layout-panels
	{
		m["layout-panels"] = example{
			w: 50, h: 5,
			view: HBox(
				VBox.Grow(1).Border(BorderRounded).Title("left")(
					Text("panel one"),
				),
				VBox.Grow(2).Border(BorderRounded).Title("right")(
					Text("panel two takes 2/3 width"),
				),
			),
		}
	}

	// first-app
	{
		count := 7

		m["first-app"] = example{
			w: 30, h: 3,
			view: VBox(
				Text(&count),
				Text("↑/↓ to count, enter to quit"),
			),
		}
	}

	// hero
	{
		cpu, mem := 72, 48
		online := true
		history := []float64{3, 5, 2, 7, 4, 6, 3, 5, 8, 4}
		requests := "1,247"
		latency := "42ms"
		errRate := "0.1%"
		uptime := "99.9%"
		status := "all systems operational"
		green := Style{FG: Green}
		cyan := Style{FG: Cyan}

		m["hero"] = example{
			w: 50, h: 11,
			view: VBox.Border(BorderRounded).BorderFG(Green).Title("glyph")(
				HBox(
					VBox.Border(BorderDouble).BorderFG(Green).Title("system").Width(16).CascadeStyle(&green)(
						If(&online).Then(Text("● ONLINE")).Else(Text("● OFFLINE").FG(Red)),
						HRule(),
						Leader("CPU", &cpu),
						Leader("MEM", &mem),
						Sparkline(&history),
					),
					SpaceW(1),
					VBox.Grow(1).CascadeStyle(&cyan)(
						Text("metrics").Bold(),
						HRule(),
						Leader("requests", &requests),
						Leader("latency", &latency),
						Leader("errors", &errRate),
						Leader("uptime", &uptime),
					),
				),
				HRule(),
				Text(&status).FG(Green),
			),
		}
	}

	// file-browser
	{
		files := []string{"main.go", "go.mod", "go.sum", "README.md", "config.toml", "Makefile"}
		preview := "package main\n\nimport (\n\t\"fmt\"\n)\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"

		m["file-browser"] = example{
			w: 60, h: 10,
			view: HBox(
				VBox.Grow(1).Border(BorderRounded)(
					List(&files).BindVimNav(),
				),
				VBox.Grow(2).Border(BorderRounded)(
					TextView(&preview).Grow(1),
				),
			),
		}
	}

	// process-monitor
	{
		type proc struct {
			PID     int
			Command string
			CPU     string
			Mem     string
		}
		procs := []proc{
			{1234, "api-server", "7.2%", "240M"},
			{1235, "postgres", "3.8%", "1.2G"},
			{1236, "redis", "19.5%", "380M"},
			{1237, "worker", "4.9%", "190M"},
			{1238, "nginx", "1.1%", "42M"},
		}
		cpuPct := 72
		memPct := 45

		m["process-monitor"] = example{
			w: 60, h: 10,
			view: VBox.Gap(1)(
				HBox.Gap(4)(
					Text("CPU"), Progress(&cpuPct).Width(30),
					Text("Mem"), Progress(&memPct).Width(30),
				),
				AutoTable(&procs).Sortable().Scrollable(20).BindVimNav(),
			),
		}
	}

	// deploy-log
	{
		frame := 0
		status := "deploying..."
		pct := 65

		m["deploy-log"] = example{
			w: 50, h: 8,
			view: VBox.Border(BorderRounded).Title("deploy")(
				HBox.Gap(2)(
					Spinner(&frame).FG(Cyan),
					Text(&status).Bold(),
					Progress(&pct).Width(20),
				),
				Text("uploading build artifacts...").FG(BrightBlack),
				Text("starting containers...").FG(BrightBlack),
				Text("running health checks...").FG(BrightBlack),
				Text("migrating database (3/5)...").FG(BrightBlack),
			),
		}
	}

	// fuzzy-finder
	{
		type pkg struct {
			Name string
			Desc string
		}
		packages := []pkg{
			{"glyph", "terminal UI framework"},
			{"cobra", "CLI command framework"},
			{"viper", "configuration management"},
			{"zap", "structured logging"},
			{"chi", "HTTP router"},
			{"sqlc", "type-safe SQL"},
			{"ent", "entity framework"},
		}

		m["fuzzy-finder"] = example{
			w: 50, h: 11,
			view: FilterList(&packages, func(p *pkg) string { return p.Name }).
				Placeholder("search packages...").
				Render(func(p *pkg) any {
					return HBox.Gap(2)(
						Text(&p.Name).Bold(),
						Text(&p.Desc).FG(BrightBlack),
					)
				}).MaxVisible(15).Border(BorderRounded).Title("packages"),
		}
	}

	// live-dashboard
	{
		reqData := make([]float64, 30)
		latData := make([]float64, 30)
		for i := range reqData {
			reqData[i] = 100 + 30*math.Sin(float64(i)*0.3)
			latData[i] = 20 + 8*math.Sin(float64(i)*0.25)
		}
		reqRate := "1,204 req/s"
		p99 := "12ms"

		m["live-dashboard"] = example{
			w: 60, h: 8,
			view: VBox(
				HBox.Gap(1)(
					VBox.Grow(1).Border(BorderRounded).Title("requests/s")(
						Sparkline(&reqData).FG(Green),
						Text(&reqRate).FG(BrightBlack),
					),
					VBox.Grow(1).Border(BorderRounded).Title("p99 latency")(
						Sparkline(&latData).FG(Yellow),
						Text(&p99).FG(BrightBlack),
					),
				),
				Text("18:42:05  GET  /api/users        200  11ms").FG(BrightBlack),
				Text("18:42:06  POST /api/sessions     200  23ms").FG(BrightBlack),
				Text("18:42:06  GET  /api/health       200   4ms").FG(BrightBlack),
			),
		}
	}

	// registration-form
	{
		name := "Pete"
		email := "pete@example.com"
		role := 0
		agree := true
		register := func() {}

		m["registration-form"] = example{
			w: 40, h: 10,
			view: VBox.Border(BorderRounded).Title("register")(
				Form.LabelBold().OnSubmit(register)(
					Field("Name", Input(&name)),
					Field("Email", Input(&email)),
					Field("Role", Radio(&role, "Admin", "User", "Guest")),
					Field("Terms", Checkbox(&agree, "I accept")),
				),
			),
		}
	}

	// concepts-effect-flip
	{
		flip := map[rune]rune{
			'a': 'ɐ', 'b': 'q', 'c': 'ɔ', 'd': 'p', 'e': 'ǝ', 'f': 'ɟ',
			'g': 'ƃ', 'h': 'ɥ', 'i': 'ᴉ', 'j': 'ɾ', 'k': 'ʞ', 'l': 'l',
			'm': 'ɯ', 'n': 'u', 'o': 'o', 'p': 'd', 'q': 'b', 'r': 'ɹ',
			's': 's', 't': 'ʇ', 'u': 'n', 'v': 'ʌ', 'w': 'ʍ', 'x': 'x',
			'y': 'ʎ', 'z': 'z',
			'A': '∀', 'B': 'ꓭ', 'C': 'Ɔ', 'D': 'ꓷ', 'E': 'Ǝ', 'F': 'Ⅎ',
			'G': '⅁', 'H': 'H', 'I': 'I', 'J': 'ſ', 'K': 'ꓘ', 'L': '⅂',
			'M': 'W', 'N': 'N', 'O': 'O', 'P': 'Ԁ', 'Q': 'Ό', 'R': 'ꓤ',
			'S': 'S', 'T': '⊥', 'U': '∩', 'V': 'Λ', 'W': 'M', 'X': 'X',
			'Y': '⅄', 'Z': 'Z',
			'1': 'Ɩ', '2': 'ᄅ', '3': 'Ɛ', '4': 'ㄣ', '5': 'ϛ',
			'6': '9', '7': 'ㄥ', '8': '8', '9': '6', '0': '0',
			'.': '˙', ',': '\'', '!': '¡', '?': '¿',
		}
		w := 34
		pad := func(s string) string {
			for len(s) < w {
				s += " "
			}
			return s[:w]
		}
		m["concepts-effect-flip"] = example{
			w: w, h: 9, pad: 2,
			view: VBox(
				Text(pad("08:42  London Euston    On time")),
				Text(pad("08:47  Bristol Temple    Plat 4")),
				Text(pad("08:55  Birmingham New   Plat 6")),
				Text(pad("09:10  Manchester Pic   Plat 3")),
				Text(pad("09:15  Edinburgh Wav    Delayed")),
				Text(pad("09:22  Leeds Central    Plat 8")),
				Text(pad("09:30  Glasgow Ctl      Plat 1")),
				ScreenEffect(EachCell(func(x, y int, c Cell, ctx PostContext) Cell {
					diag := float64(x)/float64(ctx.Width) + float64(y)/float64(ctx.Height)
					// brand gradient over everything based on diagonal position
					// #682850 (dark) → #983848 → #c44040 → #ff6060 (bright)
					t := diag / 2.0
					var r, g, b float64
					switch {
					case t < 0.33:
						s := t / 0.33
						r = 104 + s*(152-104)
						g = 40 + s*(56-40)
						b = 80 + s*(72-80)
					case t < 0.66:
						s := (t - 0.33) / 0.33
						r = 152 + s*(196-152)
						g = 56 + s*(64-56)
						b = 72 + s*(64-72)
					default:
						s := (t - 0.66) / 0.34
						r = 196 + s*(255-196)
						g = 64 + s*(96-64)
						b = 64 + s*(96-64)
					}
					// flip band shifted left (centered at 0.8)
					dist := math.Abs(diag - 0.7)
					if dist < 0.2 {
						if f, ok := flip[c.Rune]; ok {
							c.Rune = f
						}
						// brighten the flipped zone
						bright := 1.0 - dist/0.2
						r = min(255, r+bright*60)
						g = min(255, g+bright*40)
						b = min(255, b+bright*40)
					}
					c.Style.FG = RGB(uint8(r), uint8(g), uint8(b))
					return c
				})),
			),
		}
	}

	// concepts-effect-gradient
	{
		m["concepts-effect-gradient"] = example{
			w: 30, h: 8, pad: 2,
			view: VBox(
				VBox.Border(BorderRounded).Title("status")(
					Leader("api", "online"),
					Leader("cache", "warm"),
					Leader("queue", "12 pending"),
				),
				ScreenEffect(EachCell(func(x, y int, c Cell, ctx PostContext) Cell {
					t := float64(x) / float64(ctx.Width-1)
					// brand gradient: #c44040 → #ff6060 → #983848 → #682850
					var r, g, b float64
					switch {
					case t < 0.33:
						s := t / 0.33
						r = 196 + s*(255-196)
						g = 64 + s*(96-64)
						b = 64 + s*(96-64)
					case t < 0.66:
						s := (t - 0.33) / 0.33
						r = 255 + s*(152-255)
						g = 96 + s*(56-96)
						b = 96 + s*(72-96)
					default:
						s := (t - 0.66) / 0.34
						r = 152 + s*(104-152)
						g = 56 + s*(40-56)
						b = 72 + s*(80-72)
					}
					c.Style.FG = RGB(uint8(r), uint8(g), uint8(b))
					return c
				})),
			),
		}
	}

	// concepts-effect-redact
	{
		m["concepts-effect-redact"] = example{
			w: 30, h: 8, pad: 2,
			view: VBox(
				VBox.Border(BorderRounded).Title("status")(
					Leader("api", "online"),
					Leader("cache", "warm"),
					Leader("queue", "12 pending"),
				),
				ScreenEffect(EachCell(func(x, y int, c Cell, ctx PostContext) Cell {
					if x > ctx.Width*2/3 && c.Rune != '│' && c.Rune != '┐' && c.Rune != '┘' {
						c.Rune = '█'
						c.Style.FG = RGB(200, 60, 60)
					}
					return c
				})),
			),
		}
	}

	// concepts-effect-wave (rendered manually — needs full-buffer access)
	{
		m["concepts-effect-wave"] = example{w: 34, h: 9, pad: 2, view: nil}
	}

	// concepts-effect-scatter (rendered manually — needs hash-based buffer writes)
	{
		m["concepts-effect-scatter"] = example{w: 34, h: 9, pad: 2, view: nil}
	}

	// concepts-effect-collapse (rendered manually — needs radial distance calc)
	{
		m["concepts-effect-collapse"] = example{w: 34, h: 9, pad: 2, view: nil}
	}

	// concepts-effect-before (no effect applied)
	{
		m["concepts-effect-before"] = example{
			w: 24, h: 7, pad: 2,
			view: VBox(
				Text("  inbox").FG(BrightBlack),
				Text("  drafts").FG(BrightBlack),
				Text("> sent").Bold(),
				Text("  trash").FG(BrightBlack),
				Text("  spam").FG(BrightBlack),
			),
		}
	}

	// concepts-effect-after (focus dim with SelectedRef)
	{
		var selRef NodeRef
		sel := 2
		items := []string{"main.go", "handler.go", "config.go", "routes.go", "auth.go"}
		m["concepts-effect-after"] = example{
			w: 24, h: 5,
			view: VBox(
				List(&items).Selection(&sel).SelectedRef(&selRef).BindVimNav(),
				ScreenEffect(SEFocusDim(&selRef)),
			),
		}
	}

	// modal (rendered manually with vignette + overlay)
	{
		m["modal"] = example{w: 40, h: 10, view: nil}
	}

	return m
}

// renderModal composites a vignetted background with a modal dialog overlay
func renderModal() termData {
	w, h := 40, 10

	// render background — a file list
	items := []string{"server.go", "handler.go", "middleware.go", "routes.go", "config.go", "main.go", "db.go", "auth.go"}
	bgView := VBox.Border(BorderRounded).Title("files")(
		List(&items),
	)
	bgTmpl := Build(bgView)
	buf := NewBuffer(w, h)
	bgTmpl.Execute(buf, int16(w), int16(h))

	// resolve all colors to RGB before post-processing
	defaultFG := RGB(200, 196, 184)
	defaultBG := RGB(26, 26, 24) // match canvas BG (#1a1a18)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := buf.Get(x, y)
			if cell.Style.FG.Mode == ColorDefault {
				cell.Style.FG = defaultFG
			} else if cell.Style.FG.Mode == Color16 {
				cell.Style.FG = RGB(cell.Style.FG.R, cell.Style.FG.G, cell.Style.FG.B)
			}
			if cell.Style.BG.Mode == ColorDefault {
				cell.Style.BG = defaultBG
			} else if cell.Style.BG.Mode == Color16 {
				cell.Style.BG = RGB(cell.Style.BG.R, cell.Style.BG.G, cell.Style.BG.B)
			}
			buf.Set(x, y, cell)
		}
	}

	// render modal dialog — lighter background, no border, just padding
	onConfirm := func() {}
	onCancel := func() {}
	modalBG := RGB(50, 50, 48)
	modalW, modalH := 23, 4
	modalView := VBox.Fill(modalBG).PaddingVH(1, 2)(
		Text("confirm delete?"),
		HBox.Gap(4)(
			Jump(Text("yes").Bold(), onConfirm),
			Jump(Text("no"), onCancel),
		),
	)
	modalTmpl := Build(modalView)
	modalBuf := NewBuffer(modalW, modalH)
	modalTmpl.Execute(modalBuf, int16(modalW), int16(modalH))

	// resolve modal colors to explicit RGB matching the canvas BG
	for y := 0; y < modalH; y++ {
		for x := 0; x < modalW; x++ {
			cell := modalBuf.Get(x, y)
			if cell.Style.FG.Mode == ColorDefault {
				cell.Style.FG = defaultFG
			} else if cell.Style.FG.Mode == Color16 {
				cell.Style.FG = RGB(cell.Style.FG.R, cell.Style.FG.G, cell.Style.FG.B)
			}
			if cell.Style.BG.Mode == ColorDefault {
				cell.Style.BG = defaultBG
			} else if cell.Style.BG.Mode == Color16 {
				cell.Style.BG = RGB(cell.Style.BG.R, cell.Style.BG.G, cell.Style.BG.B)
			}
			modalBuf.Set(x, y, cell)
		}
	}

	// calculate modal position
	ox := (w - modalW) / 2
	oy := (h - modalH) / 2

	// apply vignette with dodge on the modal region
	modalRef := &NodeRef{X: ox, Y: oy, W: modalW, H: modalH}
	SEVignette().Strength(0.35).Dodge(modalRef).Apply(buf, PostContext{
		Width:     w,
		Height:    h,
		DefaultFG: defaultFG,
		DefaultBG: defaultBG,
	})

	// blit modal onto vignetted background
	buf.Blit(modalBuf, 0, 0, ox, oy, modalW, modalH)

	// return all lines (no trim — vignette sets explicit BG on every cell)
	td := termData{Width: w, Height: h + 3}
	for y := 0; y < h; y++ {
		td.Lines = append(td.Lines, buf.GetLineStyled(y))
	}
	return td
}

type termData struct {
	Width  int      `json:"w"`
	Height int      `json:"h"`
	Lines  []string `json:"lines"`
}

func renderTerm(buf *Buffer, w, h int) termData {
	// trim to last non-empty row
	lastRow := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := buf.Get(x, y)
			if cell.Rune != 0 && cell.Rune != ' ' {
				lastRow = y
				break
			}
		}
	}

	trimH := lastRow + 1
	td := termData{Width: w, Height: h}
	for y := 0; y < trimH; y++ {
		td.Lines = append(td.Lines, buf.GetLineStyled(y))
	}
	return td
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}

func renderWaveDemo() termData {
	w, h := 34, 11
	padStr := func(s string) string {
		for len(s) < w {
			s += " "
		}
		return s[:w]
	}
	view := VBox(
		Text(padStr("08:42  London Euston    On time")),
		Text(padStr("08:47  Bristol Temple    Plat 4")),
		Text(padStr("08:55  Birmingham New   Plat 6")),
		Text(padStr("09:10  Manchester Pic   Plat 3")),
		Text(padStr("09:15  Edinburgh Wav    Delayed")),
		Text(padStr("09:22  Leeds Central    Plat 8")),
		Text(padStr("09:30  Glasgow Ctl      Plat 1")),
	)
	tmpl := Build(view)
	buf := NewBuffer(w, h)
	tmpl.Execute(buf, int16(w), 9)

	// wave: shift each row by a sine offset
	tmp := NewBuffer(w, h)
	for y := range h {
		offset := int(3.0 * math.Sin(float64(y)*0.9))
		for x := range w {
			srcX := x - offset
			if srcX >= 0 && srcX < w {
				c := buf.Get(srcX, y)
				tmp.Set(x, y, c)
			}
		}
	}

	return renderTerm(tmp, w, h)
}

func renderScatterDemo() termData {
	w, h := 34, 11
	padStr := func(s string) string {
		for len(s) < w {
			s += " "
		}
		return s[:w]
	}
	hash := func(x, y int) int {
		h := x*374761393 + y*668265263
		h = (h ^ (h >> 13)) * 1274126177
		if h < 0 {
			h = -h
		}
		return h
	}
	view := VBox(
		Text(padStr("08:42  London Euston    On time")),
		Text(padStr("08:47  Bristol Temple    Plat 4")),
		Text(padStr("08:55  Birmingham New   Plat 6")),
		Text(padStr("09:10  Manchester Pic   Plat 3")),
		Text(padStr("09:15  Edinburgh Wav    Delayed")),
		Text(padStr("09:22  Leeds Central    Plat 8")),
		Text(padStr("09:30  Glasgow Ctl      Plat 1")),
	)
	tmpl := Build(view)
	buf := NewBuffer(w, h)
	tmpl.Execute(buf, int16(w), 9)

	// scatter: randomly replace some chars with block elements
	for y := range h {
		for x := range w {
			c := buf.Get(x, y)
			if c.Rune != ' ' && c.Rune != 0 {
				hv := hash(x, y)
				if hv%4 == 0 {
					glitches := []rune{'░', '▒', '▓', '█', '▄', '▀', '▌', '▐'}
					c.Rune = glitches[(hv>>8)%len(glitches)]
					c.Style.FG = RGB(uint8(100+hv%80), uint8(60+hv%40), uint8(80+hv%60))
					buf.Set(x, y, c)
				}
			}
		}
	}

	return renderTerm(buf, w, h)
}

func renderCollapseDemo() termData {
	w, h := 34, 11
	padStr := func(s string) string {
		for len(s) < w {
			s += " "
		}
		return s[:w]
	}
	view := VBox(
		Text(padStr("08:42  London Euston    On time")),
		Text(padStr("08:47  Bristol Temple    Plat 4")),
		Text(padStr("08:55  Birmingham New   Plat 6")),
		Text(padStr("09:10  Manchester Pic   Plat 3")),
		Text(padStr("09:15  Edinburgh Wav    Delayed")),
		Text(padStr("09:22  Leeds Central    Plat 8")),
		Text(padStr("09:30  Glasgow Ctl      Plat 1")),
	)
	tmpl := Build(view)
	buf := NewBuffer(w, h)
	tmpl.Execute(buf, int16(w), 9)

	// collapse: radial block fill from center
	cx := float64(w) / 2
	cy := float64(h) / 2
	for y := range h {
		for x := range w {
			dx := (float64(x) - cx) / cx
			dy := (float64(y) - cy) / cy
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist < 0.6 {
				t := 1.0 - dist/0.6
				blocks := []rune{' ', '░', '▒', '▓', '█'}
				idx := int(t * float64(len(blocks)-1))
				c := buf.Get(x, y)
				c.Rune = blocks[idx]
				c.Style.FG = RGB(
					uint8(min(255, 104+t*151)),
					uint8(min(255, 40+t*56)),
					uint8(min(255, 80+t*16)),
				)
				buf.Set(x, y, c)
			}
		}
	}

	return renderTerm(buf, w, h)
}

func renderJumpDemo() termData {
	w, h := 24, 7
	noop := func() {}
	view := VBox.Gap(1)(
		Jump(Text("Inbox"), noop),
		Jump(Text("Drafts"), noop),
		Jump(Text("Sent"), noop),
	)

	app := NewApp()
	tmpl := Build(view)
	tmpl.SetApp(app)

	// activate jump mode manually and render to collect targets
	jm := app.JumpMode()
	jm.Active = true
	jm.ClearJumpTargets()
	buf := NewBuffer(w, h)
	tmpl.Execute(buf, int16(w), int16(h))

	// assign labels, then re-render to draw them
	jm.AssignLabels()
	buf = NewBuffer(w, h)
	tmpl.Execute(buf, int16(w), int16(h))

	return renderTerm(buf, w, h)
}

func renderExample(id string, ex example) termData {
	if id == "modal" {
		return renderModal()
	}
	if id == "concepts-jump" {
		return renderJumpDemo()
	}
	if id == "concepts-effect-wave" {
		return renderWaveDemo()
	}
	if id == "concepts-effect-scatter" {
		return renderScatterDemo()
	}
	if id == "concepts-effect-collapse" {
		return renderCollapseDemo()
	}
	totalH := ex.h + ex.pad
	tmpl := Build(ex.view)
	buf := NewBuffer(ex.w, totalH)
	tmpl.Execute(buf, int16(ex.w), int16(ex.h))

	// apply screen effects if any were declared in the view tree
	effects := tmpl.ScreenEffects()
	if len(effects) > 0 {
		defaultFG := RGB(200, 196, 184)
		defaultBG := RGB(26, 26, 24)
		ppCtx := PostContext{
			Width:     ex.w,
			Height:    totalH,
			DefaultFG: defaultFG,
			DefaultBG: defaultBG,
		}
		for _, e := range effects {
			e.Apply(buf, ppCtx)
		}
	}

	return renderTerm(buf, ex.w, totalH)
}

func injectHTML(path string, all map[string]example) error {
	html, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	reStrip := regexp.MustCompile(` data-termdata="[^"]*"`)
	html = reStrip.ReplaceAll(html, nil)

	reTerm := regexp.MustCompile(`data-term="([^"]+)"`)
	var missing []string
	count := 0

	result := reTerm.ReplaceAllFunc(html, func(match []byte) []byte {
		sub := reTerm.FindSubmatch(match)
		id := string(sub[1])

		ex, ok := all[id]
		if !ok {
			missing = append(missing, id)
			return match
		}

		td := renderExample(id, ex)
		data, _ := json.Marshal(td)
		escaped := htmlEscape(string(data))
		count++
		return []byte(string(match) + ` data-termdata="` + escaped + `"`)
	})

	for _, id := range missing {
		fmt.Fprintf(os.Stderr, "warning: unknown example %q\n", id)
	}

	if err := os.WriteFile(path, result, 0644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "injected %d examples into %s\n", count, path)
	return nil
}

func main() {
	all := examples()

	html := false
	args := os.Args[1:]
	for i, a := range args {
		if a == "--html" {
			html = true
			args = append(args[:i], args[i+1:]...)
			break
		}
	}

	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "usage: termsvg [--html] <name|all|inject <file>>\n\navailable examples:\n")
		for name := range all {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		os.Exit(1)
	}

	name := args[0]

	if name == "inject" {
		if len(args) < 2 {
			fmt.Fprintf(os.Stderr, "usage: termsvg inject <file>\n")
			os.Exit(1)
		}
		if err := injectHTML(args[1], all); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if name == "all" {
		for n, ex := range all {
			td := renderExample(n, ex)
			data, _ := json.Marshal(td)
			path := fmt.Sprintf("img/%s.json", n)
			if err := os.WriteFile(path, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "error writing %s: %v\n", path, err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "wrote %s\n", path)
		}
		return
	}

	ex, ok := all[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown example: %s\n", name)
		os.Exit(1)
	}

	td := renderExample(name, ex)
	data, _ := json.Marshal(td)
	if html {
		fmt.Print(htmlEscape(string(data)))
	} else {
		os.Stdout.Write(data)
	}
}
