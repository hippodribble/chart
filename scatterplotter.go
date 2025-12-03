package chart

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

type scatterplotter struct {
	Plotter
	marker                  markertype
	strokecolor, fillcolor  color.Color
	markersize, strokewidth float32
	showCorrelation         bool
	pearson                 float64
	pearsonlabel            fyne.CanvasObject
	scatterShapes []fyne.CanvasObject
}

func NewScatterPlotter(X, Y []float64, ox, oy axisOrientation, seriesname string) *scatterplotter {

	pl := &scatterplotter{}
	pl.X = X
	pl.Y = Y
	pl.ox = ox
	pl.oy = oy

	pl.strokewidth = 1

	R, err := PearsonCorrelation(X, Y)
	if err != nil {
		pl.showCorrelation = false
	}
	pl.pearson = R

	// defaults
	pl.marker = Circle
	pl.markersize = 5
	pl.fillcolor = theme.Color(theme.ColorNameForeground)
	pl.strokecolor = theme.Color(theme.ColorNameBackground)
	// pl.strokewidth = 1

	pl.createPlotObjects()

	return pl
}

func (s *scatterplotter) SetMarkerType(m markertype) {
	s.marker = m
	switch s.marker {
	case Circle:
		for i := range s.scatterShapes {
			c := canvas.NewCircle(s.fillcolor)
			c.StrokeColor = s.strokecolor
			c.StrokeWidth = s.strokewidth
			c.Resize(fyne.NewSize(s.markersize, s.markersize))
			s.scatterShapes[i] = c
		}

	case Square:
		for i := range s.scatterShapes {
			r := canvas.NewRectangle(s.fillcolor)
			r.StrokeColor = s.strokecolor
			r.StrokeWidth = s.strokewidth
			s.scatterShapes[i] = r
			s.scatterShapes[i].Resize(fyne.NewSize(s.markersize, s.markersize))
		}
	}
}

func (s *scatterplotter) SetMarkerSize(size float32) {
	s.markersize = size
	for _, sh := range s.scatterShapes[:len(s.X)] {
		switch s.marker {
		case Circle:
			sh.(*canvas.Circle).Resize(fyne.NewSize(size, size))
		case Square:
			sh.(*canvas.Rectangle).Resize(fyne.NewSize(size, size))
		}
	}
}

func (s *scatterplotter) SetStrokeColor(c color.Color) {
	s.strokecolor = c
	for _, sh := range s.scatterShapes[:len(s.X)] {
		switch s.marker {
		case Circle:
			sh.(*canvas.Circle).StrokeColor = c
		case Square:
			sh.(*canvas.Rectangle).StrokeColor = c
		}
	}
}
func (s *scatterplotter) SetFillColor(c color.Color) {
	s.fillcolor = c
	for _, sh := range s.scatterShapes[:len(s.X)] {
		switch s.marker {
		case Circle, Square:
			sh.(*canvas.Circle).FillColor = c
		}
	}
}
func (s *scatterplotter) SetStrokeWidth(w float32) {
	s.strokewidth = w
	for _, sh := range s.scatterShapes[:len(s.X)] {
		switch s.marker {
		case Circle, Square:
			sh.(*canvas.Circle).StrokeWidth = w
		}
	}
}

func (s *scatterplotter) ShowCorrelation(b bool) {
	s.showCorrelation = b
}

func (s *scatterplotter) dataRange() [2]axisLimits {
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

func (s *scatterplotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

// Makes the list of objects to be drawn by the plot renderer
func (s *scatterplotter) createPlotObjects() error {

	s.scatterShapes = make([]fyne.CanvasObject, len(s.X))

	for i := range s.X {
		switch s.marker {
		case Circle:
			c := canvas.NewCircle(s.fillcolor)
			c.StrokeColor = s.strokecolor
			c.StrokeWidth = s.strokewidth
			s.scatterShapes[i] = c
		case Square:
			sq := canvas.NewRectangle(s.fillcolor)
			sq.StrokeColor = s.strokecolor
			sq.StrokeWidth = s.strokewidth
			s.scatterShapes[i] = sq
		}
	}

	t := canvas.NewText(fmt.Sprintf("R=%.3g", s.pearson), theme.Color(theme.ColorNameForeground))
	t.TextSize = 10
	t.TextStyle.Bold = true
	t.TextStyle.Italic = true
	s.pearsonlabel = t

	return nil
}

// Repositions the list of objects to be drawn by the plot renderer
func (s *scatterplotter) positionPlotObjects(cv *plot) error {

	if len(s.X) != len(s.Y) || len(s.X) == 0 {
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

	// log.Println("SCATTERPLOTTER: marker size:",s.markersize,s.strokecolor,s.fillcolor,s.strokewidth)

	for i, x := range s.X {
		y := s.Y[i]
		xplot := w * (x - xmin) / (xmax - xmin)
		yplot := h * (ymax - y) / (ymax - ymin)
		if _, ok := s.scatterShapes[i].(*canvas.Circle); ok {
			s.scatterShapes[i].Resize(fyne.NewSize(s.markersize, s.markersize))
			s.scatterShapes[i].Move(fyne.NewPos(float32(xplot)-s.markersize/2, float32(yplot)-s.markersize/2))
		} else if _, ok := s.scatterShapes[i].(*canvas.Rectangle); ok {
			s.scatterShapes[i].Resize(fyne.NewSize(s.markersize, s.markersize))
			s.scatterShapes[i].Move(fyne.NewPos(float32(xplot)-s.markersize/2, float32(yplot)-s.markersize/2))
		}
	}

	t := s.pearsonlabel.(*canvas.Text)
	xplot := w * .05
	yplot := h * .05
	t.Move(fyne.NewPos(float32(xplot), float32(yplot)))

	return nil
}

func (s *scatterplotter) allShapes() []fyne.CanvasObject {
	if s.showCorrelation {
		return append(s.scatterShapes, s.pearsonlabel)
	}
	return s.scatterShapes
}

func (s *scatterplotter) allAxes() []axisOrientation {
	return []axisOrientation{s.ox, s.oy}
}
