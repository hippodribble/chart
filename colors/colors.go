package colors

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/crazy3lf/colorconv"
	eventbus "github.com/dtomasi/go-event-bus/v3"
)

var Red = color.RGBA{255, 0, 0, 255}
var Green = color.RGBA{0, 255, 0, 255}
var Blue = color.RGBA{0, 0, 255, 255}
var Cyan = color.RGBA{0, 255, 255, 255}
var Magenta = color.RGBA{255, 0, 255, 255}
var Yellow = color.RGBA{255, 255, 0, 255}
var Orange = color.RGBA{255, 165, 0, 255}
var Black = color.Black
var White = color.White
var Grey = color.RGBA{128, 128, 128, 255}
var LightGrey = color.RGBA{224, 224, 224, 255}
var DarkGrey = color.RGBA{64, 64, 64, 255}

var PlotBlue = color.RGBA{0, 0, 255, 128}
var PlotRed = color.RGBA{255, 0, 0, 128}
var PlotGreen = color.RGBA{0, 255, 0, 128}
var PlotCyan = color.RGBA{0, 255, 255, 128}
var PlotMagenta = color.RGBA{255, 0, 255, 128}
var PlotOrange = color.RGBA{255, 165, 0, 255}
var PlotGrey = color.RGBA{0, 0, 0, 128}
var PlotLightGrey = color.RGBA{0, 0, 0, 64}

var Colorlist = []color.Color{
	Red,
	Green,
	Blue,
	Cyan,
	Magenta,
	Orange,
	LightGrey,
	DarkGrey,
}

var TranslucentColors = []color.Color{
	Black,
	PlotGrey,
	PlotLightGrey,
	PlotRed,
	PlotGreen,
	PlotBlue,
	PlotCyan,
	PlotMagenta,
	PlotOrange,
}

func MakeGreys() []color.Color {
	greys := make([]color.Color, 16)
	for i := 0; i < 16; i++ {
		greys[i] = color.RGBA{uint8(i * 16), uint8(i * 16), uint8(i * 16), 255}
	}
	return greys
}

func MakeSpanningGreys(n int) []color.Color {
	greys := make([]color.Color, n)
	d := 192.0 / float64(n)
	for i := 0; i < n; i++ {
		greys[i] = color.Gray{uint8(float64(i)*d + 32)}
	}
	return greys
}

var Brewer8 = []color.Color{
	color.RGBA{127, 201, 127, 255},
	color.RGBA{190, 174, 212, 255},
	color.RGBA{253, 192, 134, 255},
	color.RGBA{255, 255, 153, 255},
	color.RGBA{56, 108, 176, 255},
	color.RGBA{240, 2, 127, 255},
	color.RGBA{191, 91, 23, 255},
	color.RGBA{102, 102, 102, 255},
}

type colorgen func() color.Color

func NewColor(cc []color.Color) colorgen {
	c := cc
	i := -1
	return func() color.Color {
		i++
		i = i % len(c)
		return c[i]
	}
}

func InvertColor(c color.Color) color.Color {
	// log.Println(c)
	r, g, b, a := c.RGBA()
	return color.RGBA{255 - uint8(r), 255 - uint8(g), 255 - uint8(b), uint8(a)}
}

type ColorRectangle struct {
	widget.BaseWidget
	bH, bL, bS binding.Float
	ras        *canvas.Raster
	outline    *canvas.Rectangle
	savedL     float64
	bus        *eventbus.EventBus
}

func NewColorRectangle(h, l, s binding.Float, bus *eventbus.EventBus) *ColorRectangle {
	r := &ColorRectangle{
		bH:  h,
		bL:  l,
		bS:  s,
		bus: bus,
	}

	r.bH.AddListener(binding.NewDataListener(func() { r.ras.Refresh() }))
	r.bS.AddListener(binding.NewDataListener(func() { r.ras.Refresh() }))
	r.ras = canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		luminance := 1 - float64(y)/float64(h)/1
		hue, _ := r.bH.Get()
		sat, _ := r.bS.Get()

		r, g, b, err := colorconv.HSLToRGB(hue, sat, luminance)
		if err != nil {
			log.Println(err)
		}
		return color.RGBA{r, g, b, 255}
	})
	r.outline = canvas.NewRectangle(color.Transparent)
	r.outline.StrokeColor = color.Black
	r.outline.StrokeWidth = 5
	r.ras.SetMinSize(fyne.NewSize(50, 50))

	r.ExtendBaseWidget(r)
	return r
}

func (r *ColorRectangle) CreateRenderer() fyne.WidgetRenderer {
	c := container.NewPadded(container.NewStack(r.ras, r.outline))
	return widget.NewSimpleRenderer(c)
}

func (r *ColorRectangle) MouseMoved(evt *desktop.MouseEvent) {
	raspos := r.ras.Position()
	mousepos := evt.Position
	rassize := r.ras.Size()
	dpos := mousepos.Subtract(raspos)
	if dpos.X > 0 && dpos.Y > 0 && dpos.X < rassize.Width && dpos.Y < rassize.Height {
		newLuminance := 1 - dpos.Y/rassize.Height
		newLuminance /= 2
		if newLuminance < .01 {
			newLuminance = .01
		}
		r.bL.Set(float64(newLuminance))
	}
}
func (r *ColorRectangle) MouseIn(evt *desktop.MouseEvent) {

}
func (r *ColorRectangle) MouseOut() {
	r.bL.Set(r.savedL)
}

func (r *ColorRectangle) MouseDown(evt *desktop.MouseEvent) {
}
func (r *ColorRectangle) MouseUp(evt *desktop.MouseEvent) {
	l, _ := r.bL.Get()
	r.savedL = l
	h, _ := r.bH.Get()
	s, _ := r.bS.Get()
	r.bus.PublishAsync("CLICK", fmt.Sprintf("H: %.3g S: %.2f L: %.2f", h, s, l))
}

type ColorPicker struct {
	widget.BaseWidget
	desktop.Hoverable
	lumerect       *ColorRectangle
	backingImage   *canvas.Raster
	outercircle    *canvas.Circle
	base           *canvas.Rectangle
	bH, bL, bS     binding.Float
	savedH, savedS float64
	label          *canvas.Text
	bus            *eventbus.EventBus
}

func NewColorPicker() *ColorPicker {
	p := &ColorPicker{
		outercircle: canvas.NewCircle(color.Transparent),
		base:        canvas.NewRectangle(Grey),
		label:       canvas.NewText("Click to Pick...", color.RGBA{192, 192, 192, 255}),
		bus:         eventbus.NewEventBus(),
	}
	p.label.TextSize = 18
	p.label.TextStyle.Bold = true

	p.outercircle.StrokeColor = color.Black
	p.outercircle.StrokeWidth = 5
	p.base = canvas.NewRectangle(Grey)

	p.bL = binding.NewFloat()
	p.bH = binding.NewFloat()
	p.bS = binding.NewFloat()
	p.bL.Set(0.5)
	p.bH.Set(0)
	p.bS.Set(1)
	p.lumerect = NewColorRectangle(p.bH, p.bL, p.bS, p.bus)

	p.backingImage = canvas.NewRasterWithPixels(
		func(x, y, w, h int) color.Color {
			rmax := float64(min(w, h)) / 2
			dx := x - w/2
			dy := h/2 - y
			r := math.Sqrt(float64(dx*dx + dy*dy))
			θ := math.Atan2(float64(dy), float64(dx))
			var C color.Color
			if r < rmax {
				saturation := r / rmax
				hue := θ / math.Pi * 180
				if hue < 0 {
					hue += 360
				}
				ll, _ := p.bL.Get()
				luminance := ll
				r, g, b, _ := colorconv.HSLToRGB(hue, saturation, luminance)
				C = color.RGBA{r, g, b, 255}
				return C
			}
			return color.RGBA{255, 255, 255, 0}
		},
	)

	p.ExtendBaseWidget(p)
	p.listen()
	return p
}

func (p *ColorPicker) CreateRenderer() fyne.WidgetRenderer {
	b := container.NewStack(p.base, container.NewBorder(nil, p.label, nil, p.lumerect, container.NewStack(p.backingImage, p.outercircle)))
	return widget.NewSimpleRenderer(b)
}

func (p *ColorPicker) Refresh() {
	p.BaseWidget.Refresh()
	p.backingImage.Refresh()
}

func (p *ColorPicker) MouseMoved(evt *desktop.MouseEvent) {
	mousepos := evt.Position
	circlepos := p.backingImage.Position()
	circlesize := p.backingImage.Size()
	rmax := min(circlesize.Width, circlesize.Height) / 2
	ccentrex := circlesize.Width/2 + circlepos.X
	ccentrey := circlesize.Height/2 + circlepos.Y
	dx := mousepos.X - ccentrex
	dy := mousepos.Y - ccentrey
	if dx*dx+dy*dy > rmax*rmax {
		return
	}

	r := math.Sqrt(float64(dx*dx + dy*dy))
	θ := math.Atan2(float64(-dy), float64(dx))
	if r < float64(rmax) {
		saturation := r / float64(rmax)
		hue := θ / math.Pi * 180
		if hue < 0 {
			hue += 360
		}
		p.bH.Set(hue)
		p.bS.Set(saturation)
		p.lumerect.Refresh()
	}
}
func (p *ColorPicker) MouseIn(evt *desktop.MouseEvent) {

}
func (p *ColorPicker) MouseOut() {
	p.bH.Set(p.savedH)
	p.bS.Set(p.savedS)
	log.Println("Hue set to", p.savedH)
}

func (p *ColorPicker) MouseDown(evt *desktop.MouseEvent) {

}

func (p *ColorPicker) MouseUp(evt *desktop.MouseEvent) {
	h, _ := p.bH.Get()
	s, _ := p.bS.Get()
	p.savedH = h
	p.savedS = s
	p.lumerect.Refresh()
	l, _ := p.bL.Get()
	p.label.Text = fmt.Sprintf("H: %.3g S: %.2f L: %.2f", h, s, l)
}

func (p *ColorPicker) listen() {
	subClick := p.bus.Subscribe("CLICK")
	go func() {
		for update := range subClick {
			if str, ok := update.Data.(string); ok {
				p.label.Text = str
				p.label.Refresh()
			}
			update.Done()
		}
	}()
}
