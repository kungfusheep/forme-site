package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

	. "github.com/kungfusheep/glyph"
)

type example struct {
	view any
	w, h int
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
			w: 16, h: 6,
			view: VBox.Gap(1)(
				Text("hello"),
				Text("world"),
				Text("!"),
			),
		}
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
	td := termData{Width: w, Height: h}
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
		fmt.Fprintf(os.Stderr, "usage: termsvg [--html] <name|all>\n\navailable examples:\n")
		for name := range all {
			fmt.Fprintf(os.Stderr, "  %s\n", name)
		}
		os.Exit(1)
	}

	name := args[0]

	if name == "all" {
		for n, ex := range all {
			var td termData
			if n == "modal" {
				td = renderModal()
			} else {
				tmpl := Build(ex.view)
				buf := NewBuffer(ex.w, ex.h)
				tmpl.Execute(buf, int16(ex.w), int16(ex.h))
				td = renderTerm(buf, ex.w, ex.h)
			}
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

	if name == "modal" {
		td := renderModal()
		data, _ := json.Marshal(td)
		os.Stdout.Write(data)
		return
	}

	ex, ok := all[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown example: %s\n", name)
		os.Exit(1)
	}

	tmpl := Build(ex.view)
	buf := NewBuffer(ex.w, ex.h)
	tmpl.Execute(buf, int16(ex.w), int16(ex.h))
	td := renderTerm(buf, ex.w, ex.h)
	data, _ := json.Marshal(td)
	if html {
		fmt.Print(htmlEscape(string(data)))
	} else {
		os.Stdout.Write(data)
	}
}
