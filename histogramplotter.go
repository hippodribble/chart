package chart

import (
	"errors"
	"fmt"
	"image/color"
	"log"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"github.com/hippodribble/chart/colors"
)

// For plotting histograms
type histogramPlotter struct {
	Plotter
	strokecolor, fillcolor color.Color
	strokewidth            float32
	Y2                     []float64
	fillwidth              float64
	binwidth               float64
	showStats              bool
	statsLabel             *canvas.Text
}

func NewHistogramPlotter(limits []float64, counts []float64, ox, oy axisOrientation) *histogramPlotter {
	if len(limits)-len(counts) != 1 {
		log.Printf("%d limits for %d values\n", len(limits), len(counts))
	}
	pl := &histogramPlotter{}

	pl.Y = counts
	pl.ox = ox
	pl.oy = oy

	pl.X = []float64{}

	pl.X = limits

	pl.binwidth = limits[1] - limits[0]

	// defaults

	pl.strokecolor = theme.Color(theme.ColorNameForeground)
	pl.strokewidth = 0
	pl.fillwidth = 90
	pl.fillcolor = colors.Grey

	// Makes the list of objects to be drawn by the plot renderer
	//
	//	In this case, a series of rectangles of fixed width and
	//	varying height and vertical position

	pl.shapes = make([]fyne.CanvasObject, len(pl.X))

	for i := 1; i < len(pl.X); i++ {

		r := canvas.NewRectangle(pl.fillcolor)
		// r.value = x
		r.FillColor = pl.fillcolor
		r.StrokeWidth = pl.strokewidth
		r.StrokeWidth = 0
		r.StrokeColor = theme.Color(theme.ColorNameForeground)
		pl.shapes[i] = r
	}

	return pl
}

func (s *histogramPlotter) SetFillColor(c color.Color) {
	s.fillcolor = c
}
func (s *histogramPlotter) SetFillWidth(percent float64) {
	if percent < 5 {
		percent = 5
	}
	// if percent > 100 {
	// 	percent = 100
	// }
	s.fillwidth = percent
}

func (s *histogramPlotter) SetStrokeColor(c color.Color) {
	s.strokecolor = c
}

func (s *histogramPlotter) SetStrokeWidth(w float32) {
	s.strokewidth = w
}

func (s *histogramPlotter) ShowStatistics(b bool) {
	s.showStats = b
}

// for a histogram Ymin is always zero
func (s *histogramPlotter) dataRange() [2]axisLimits {

	minx := 1.0e+20
	maxx := -1.0e+20
	maxy := -1.0e+20

	for _, y := range s.Y {

		if y > maxy {
			maxy = y
		}
	}

	for _, x := range s.X {

		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
	}
	return [2]axisLimits{{minx, maxx}, {0, maxy}}
}

func (h *histogramPlotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

func (s *histogramPlotter) positionPlotObjects(cv *plot) error {

	// drawing means we need to know where the axes are
	// each Plotter can only have one X and one Y axis (multi-axis plots are composed of multiple Plotters)
	var xaxes, yaxes []*axis

	for _, axis := range cv.ch.axes {
		if axis.orientation == s.ox {
			xaxes = append(xaxes, axis)
		}
		if axis.orientation == s.oy {
			yaxes = append(yaxes, axis)
		}
	}

	if len(xaxes) != 1 || len(yaxes) != 1 {
		// log.Println(len(xaxes), s.ox)
		// log.Println(len(yaxes), s.oy)
		log.Fatalln("There appear to be missing axes for the data to be drawn on")
	}

	xaxis := xaxes[0]
	yaxis := yaxes[0]

	// we need to know where the data is as a fraction of the plot size
	xmin := xaxis.limits.min
	xmax := xaxis.limits.max

	ymin := yaxis.limits.min
	ymax := yaxis.limits.max

	var w float64 = float64(cv.Size().Width)
	var h float64 = float64(cv.Size().Height)

	width := w / (xmax - xmin) * s.fillwidth / 100 * s.binwidth

	for i := 1; i < len(s.X); i++ {

		x := (s.X[i] + s.X[i-1]) / 2
		yTop := s.Y[i-1]

		r := s.shapes[i].(*canvas.Rectangle)

		xplot := w*(x-xmin)/(xmax-xmin) - width/2
		hTop := h * (ymax - yTop) / (ymax - ymin)
		r.Move(fyne.NewPos(float32(xplot), float32(hTop)))
		r.Resize(fyne.NewSize(float32(width), float32(h-hTop)))
		r.Refresh()
	}

	x := w * .05
	y := h * .05

	if s.statsLabel == nil {
		log.Println("HISTGRAM PLOTTER: no label to move")
		return errors.New("HISTOGRAM PLOTTER: no histogram label to midfy")
	}

	s.statsLabel.Move(fyne.NewPos(float32(x), float32(y)))

	return nil
}

func (s *histogramPlotter) allShapes() []fyne.CanvasObject {
	return append(s.shapes, s.statsLabel)
}

func (s *histogramPlotter) allAxes() []axisOrientation {
	return []axisOrientation{s.ox, s.oy}
}

func MakeHistogram(values []float64) (*histogramPlotter, error) {

	limits, err := MakeHistogramBinLimits(values)
	if err != nil {
		return nil, err
	}

	low := limits[0]
	dx := limits[1] - limits[0]
	bins := make([]float64, len(limits)-1)

	for _, v := range values {
		x := int((v - low) / dx)
		bins[x]++
	}

	pl := NewHistogramPlotter(limits, bins, Bottom, Left)

	N := float64(len(values))
	var sx, sxx float64

	for _, v := range values {
		sx += v
		sxx += v * v
	}

	mean := sx / float64(N)

	sxx /= N
	sx /= N

	sd := math.Sqrt(sxx - sx*sx)
	pl.statsLabel = canvas.NewText(fmt.Sprintf("μ=%.3g σ=%.3g N=%d", mean, sd, int(N)), theme.Color(theme.ColorNameForeground))
	pl.statsLabel.TextSize = 10
	pl.statsLabel.TextStyle.Bold = true
	pl.statsLabel.TextStyle.Italic = true

	return pl, nil
}
