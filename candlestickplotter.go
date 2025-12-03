package chart

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/hippodribble/chart/colors"
)

// A floating bar plot is used to display a vertical bar disconnected from the origin
//
//	It used an X value and 2 Y values.
//
//	Ideally, the 2nd Y value is larger than the first. If not, the values will be swapped
//	and the colours will be inverted to indicate this.
//
//	The rectangles themselves are hoverable and clickable.
//
//	Setting the Y1 values to zero results in a standard vertical bar chart.
type candlestickplotter struct {
	Plotter
	strokecolor, fillcolor color.Color
	strokewidth            float32
	O, H, L, C             []float64
	fillwidth              float64
}

func NewCandlestickPlotter(X, O, H, L, C []float64, ox, oy axisOrientation) *candlestickplotter {
	pl := &candlestickplotter{}
	pl.X = X
	pl.O = O
	pl.H = H
	pl.L = L
	pl.C = C
	pl.ox = ox
	pl.oy = oy

	// defaults

	pl.strokecolor = color.Black
	pl.strokewidth = 1
	pl.fillwidth = 100
	pl.fillcolor = color.Transparent

	pl.makePlotObjects()

	return pl
}

func (s *candlestickplotter) SetFillColor(c color.Color) {
	s.fillcolor = c
}
func (s *candlestickplotter) SetFillWidth(percent float64) {
	if percent < 5 {
		percent = 5
	}
	// if percent > 100 {
	// 	percent = 100
	// }
	s.fillwidth = percent
}

func (s *candlestickplotter) SetStrokeColor(c color.Color) {
	s.strokecolor = c
}

func (s *candlestickplotter) SetStrokeWidth(w float32) {
	s.strokewidth = w
}

func (s *candlestickplotter) dataRange() [2]axisLimits {
	minx := 1.0e+20
	maxx := -1.0e+20
	miny := 1.0e+20
	maxy := -1.0e+20
	for i, x := range s.X {
		h := s.H[i]
		l := s.L[i]
		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if h > maxy {
			maxy = h
		}
		if l < miny {
			miny = l
		}
	}
	return [2]axisLimits{{minx, maxx}, {miny, maxy}}
}

func (h *candlestickplotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

// Makes the list of objects to be drawn by the plot renderer
//
//	In this case, a series of rectangles of fixed width and
//	varying height and vertical position, as well as two lines from the bax to the extrema

func (s *candlestickplotter) makePlotObjects() {

	s.shapes = make([]fyne.CanvasObject, len(s.X)*4)

	for i := 0; i < len(s.X); i++ {
		// log.Println("candlestick",i)
		x1 := s.X[i]
		r := NewRectangle()
		r.value = x1
		r.StrokeWidth = 0
		r.StrokeColor = s.strokecolor
		if s.O[i] < s.C[i] {
			r.StrokeColor = colors.LightGrey
		} else {
			r.StrokeColor = colors.Grey
		}

		r.FillColor = r.StrokeColor

		s.shapes[4*i] = r
		s.shapes[4*i+1] = &r.Rectangle

		lLo := canvas.NewLine(s.strokecolor)

		lLo.StrokeColor = r.StrokeColor
		lLo.StrokeWidth = 1
		// log.Println(lLo.Position1, lLo.Position2)

		// c := fyne.CanvasObject(lLo)
		s.shapes[4*i+2] = lLo

		lHi := canvas.NewLine(s.strokecolor)

		lHi.StrokeColor = r.StrokeColor
		lHi.StrokeWidth = 1
		// log.Println(lHi.Position1, lHi.Position2)

		// d := fyne.CanvasObject(lHi)
		s.shapes[4*i+3] = lHi

	}

}
func (s *candlestickplotter) positionPlotObjects(cv *plot) error {
	// log.Println("Make candlesticks called")

	if len(s.X) != len(s.O) || len(s.X) == 0 {
		return fmt.Errorf("the X and Y data must be of the same length , %d <> %d", len(s.X), len(s.Y))
	}

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
		log.Println(len(xaxes), s.ox)
		log.Println(len(yaxes), s.oy)
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

	width := w / (xmax - xmin) * s.fillwidth / 100
	if width < 0 {
		width = -width
	}
	for width < 10 {
		width *= 1.1
	}

	for i := 0; i < len(s.X); i++ {
		x1 := s.X[i]
		yRectLow := s.O[i]
		yRectHigh := s.C[i]
		r := s.shapes[4*i].(*rectangle)
		r.value = x1
		if s.O[i] < s.C[i] {
			r.StrokeColor = colors.LightGrey
		} else {
			r.StrokeColor = colors.Grey
		}

		r.FillColor = r.StrokeColor

		if yRectLow > yRectHigh {
			yRectLow, yRectHigh = yRectHigh, yRectLow
		}

		xplot1 := w*(x1-xmin)/(xmax-xmin) - width/2
		rectBottom := h * (ymax - yRectLow) / (ymax - ymin)
		rectTop := h * (ymax - yRectHigh) / (ymax - ymin)
		lineMinimumY := h * (ymax - s.L[i]) / (ymax - ymin)
		lineMaximumY := h * (ymax - s.H[i]) / (ymax - ymin)

		r.Move(fyne.NewPos(float32(xplot1), float32(rectTop)))
		r.Resize(fyne.NewSize(float32(width), float32(rectBottom-rectTop)))

		// s.shapes[4*i] = r
		s.shapes[4*i+1] = &r.Rectangle

		// lLo := canvas.NewLine(s.strokecolor)
		lLo := s.shapes[4*i+2].(*canvas.Line)
		lLo.Position1 = fyne.NewPos(
			float32(w*(x1-xmin)/(xmax-xmin)),
			float32(h*(ymax-yRectLow)/(ymax-ymin)),
		)
		lLo.Position2 = fyne.NewPos(
			float32(w*(x1-xmin)/(xmax-xmin)),
			float32(lineMinimumY),
		)
		lLo.StrokeColor = r.StrokeColor
		lLo.StrokeWidth = 1
		// log.Println(lLo.Position1, lLo.Position2)

		// c := fyne.CanvasObject(lLo)
		s.shapes[4*i+2] = lLo

		lHi := s.shapes[4*i+3].(*canvas.Line)
		lHi.Position1 = fyne.NewPos(
			float32(w*(x1-xmin)/(xmax-xmin)),
			float32(h*(ymax-yRectHigh)/(ymax-ymin)),
		)
		lHi.Position2 = fyne.NewPos(
			float32(w*(x1-xmin)/(xmax-xmin)),
			float32(lineMaximumY),
		)
		lHi.StrokeColor = r.StrokeColor
		lHi.StrokeWidth = 1
		// log.Println(lHi.Position1, lHi.Position2)

		// d := fyne.CanvasObject(lHi)
		s.shapes[4*i+3] = lHi

	}

	// log.Println("made",len(s.shapes))

	return nil
}

func (s *candlestickplotter) allShapes() []fyne.CanvasObject {
	return s.shapes
}

func (s *candlestickplotter) allAxes() []axisOrientation {
	return []axisOrientation{s.ox, s.oy}
}
