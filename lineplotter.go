package chart

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

type LinePlotter struct {
	Plotter
	strokecolor color.Color
	strokewidth float32
	seriesname  string
	HideLegend  bool
	Legend      *LineLegendEntry
}

func NewLinePlotter(X, Y []float64, ox, oy axisOrientation, seriesname string) *LinePlotter {
	pl := &LinePlotter{seriesname: seriesname, strokecolor: theme.Color(theme.ColorNameForeground), strokewidth: 1}
	pl.X = X
	pl.Y = Y
	pl.ox = ox
	pl.oy = oy

	// defaults


	pl.Legend = NewLineLegendEntry(seriesname, pl.strokewidth, pl.strokecolor)

	// line segments that make up the line
	pl.shapes = make([]fyne.CanvasObject, len(pl.X)-1)
	for i := 1; i < len(pl.X); i++ {
		line := canvas.NewLine(pl.strokecolor)
		line.StrokeWidth = pl.strokewidth
		pl.shapes[i-1] = line
	}
	return pl
}

func (s *LinePlotter) SetStrokeColor(c color.Color) {
	s.strokecolor = c
	for i := range s.shapes {
		if sh, ok := s.shapes[i].(*canvas.Line); ok {
			sh.StrokeColor = c
		}

	}
	s.Legend = NewLineLegendEntry(s.seriesname, s.strokewidth, s.strokecolor)
}

func (s *LinePlotter) SetStrokeWidth(w float32) {
	s.strokewidth = w
	for i := range s.shapes {
		if sh, ok := s.shapes[i].(*canvas.Circle); ok {
			sh.StrokeWidth = w
		}
		if sh, ok := s.shapes[i].(*canvas.Rectangle); ok {
			sh.StrokeWidth = w
		}
	}
	s.Legend = NewLineLegendEntry(s.seriesname, s.strokewidth, s.strokecolor)
}

func (s *LinePlotter) dataRange() [2]axisLimits {
	minx := 1.0e+20
	maxx := -1.0e+20
	miny := 1.0e+20
	maxy := -1.0e+20
	for i, x := range s.X {
		y := s.Y[i]
		if x < minx {
			minx = x
		}
		if x > maxx {
			maxx = x
		}
		if y < miny {
			miny = y
		}
		if y > maxy {
			maxy = y
		}
	}
	return [2]axisLimits{{minx, maxx}, {miny, maxy}}
}

func (s *LinePlotter) legendEntry() fyne.CanvasObject {
	if s.HideLegend {
		return canvas.NewRectangle(color.Transparent)
	}
	return s.Legend
}

// Makes the list of objects to be drawn by the plot renderer
func (s *LinePlotter) positionPlotObjects(cv *plot) error {

	if len(s.X) != len(s.Y) || len(s.X) == 0 {
		return fmt.Errorf("the X and Y data must be of the same length , %d <> %d", len(s.X), len(s.Y))
	}

	// log.Printf("    LinePlotter makePlotObjects()")

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

	for i := 1; i < len(s.X); i++ {
		x1 := s.X[i-1]
		y1 := s.Y[i-1]
		x2 := s.X[i]
		y2 := s.Y[i]
		xplot1 := w * (x1 - xmin) / (xmax - xmin)
		yplot1 := h * (ymax - y1) / (ymax - ymin)
		xplot2 := w * (x2 - xmin) / (xmax - xmin)
		yplot2 := h * (ymax - y2) / (ymax - ymin)
		if l, ok := s.shapes[i-1].(*canvas.Line); ok {
			l.Position1 = fyne.NewPos(float32(xplot1), float32(yplot1))
			l.Position2 = fyne.NewPos(float32(xplot2), float32(yplot2))
		}
	}

	return nil
}

func (s *LinePlotter) allShapes() []fyne.CanvasObject {
	return s.shapes
}

func (s *LinePlotter) allAxes() []axisOrientation {
	return []axisOrientation{s.ox, s.oy}
}
