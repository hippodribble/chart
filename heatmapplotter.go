package chart

import (
	"image/color"
	"log"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

type heatmapplotter struct {
	Plotter

	heatmapShapes  []fyne.CanvasObject
	colourRange    *[2]color.RGBA
	dx, dy, aspect float32
	vLabels        []string
	colours        []color.RGBA
}

func NewHeatMapPlotter(X []float64, ox, oy axisOrientation, vLabels []string, dx, dy, aspect float32, colours *[2]color.RGBA) *heatmapplotter {

	pl := &heatmapplotter{
		vLabels:     vLabels,
		dx:          dx,
		dy:          dy,
		aspect:      aspect,
		colourRange: colours,
	}
	pl.X = X
	pl.ox = ox
	pl.oy = oy

	pl.makeColors(len(X))
	pl.makeCoordinates(len(X))
	// pl.adjustLimits()
	pl.createPlotObjects()

	return pl
}

func (s *heatmapplotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

// Makes the list of objects to be drawn by the plot renderer
func (s *heatmapplotter) createPlotObjects() error {

	s.heatmapShapes = make([]fyne.CanvasObject, len(s.X))

	for i := range s.X {
		r := canvas.NewRectangle(s.colours[i])
		r.StrokeColor = theme.Color(theme.ColorNameBackground)
		r.StrokeWidth = 2.5
		s.heatmapShapes[i] = r
	}
	return nil
}

func (s *heatmapplotter) makeColors(N int) {
	if s.colourRange == nil {
		log.Println("no colours defined")
		cc := [2]color.RGBA{}
		cc[0] = color.RGBA{0, 0, 255, 255}
		cc[1] = color.RGBA{255, 0, 0, 255}
		s.colourRange = &cc
	}
	mn := +1.0e+20
	mx := -1.0e-20
	for _, v := range s.X {
		mx = max(v, mx)
		mn = min(mn, v)
	}
	d := mx - mn

	s.colours = make([]color.RGBA, N)
	for i, v := range s.X {
		f := (v - mn) / d
		b := uint8(f * 255)
		r := uint8((1 - f) * 255)
		g := uint8((math.Abs(f - .5)) * 255)
		s.colours[i] = color.RGBA{r, g, b, 255}
		// fmt.Println(i, s.colours[i], f)
	}
	// fmt.Println(len(s.colours), "colours")
}

// generate X,Y from the data
func (s *heatmapplotter) makeCoordinates(N int) {
	s.X = make([]float64, N)
	s.Y = make([]float64, N)
	for i := range N { // for each point
		s.Y[i] = float64(i % len(s.vLabels))
		s.X[i] = float64(i / len(s.vLabels))
	}
}

// Repositions the list of objects to be drawn by the plot renderer INTERFACE METHOD
func (s *heatmapplotter) positionPlotObjects(cv *plot) error {

	// if len(s.X) != len(s.Y) || len(s.X) == 0 {
	// 	return fmt.Errorf("the X and Y data must be of the same length , %d <> %d", len(s.X), len(s.Y))
	// }

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
	dx := xmax - xmin
	dy := ymax - ymin
	var w float64 = float64(cv.Size().Width)
	var h float64 = float64(cv.Size().Height)

	wPer := w / dx
	hPer := h / dy

	// log.Println("SCATTERPLOTTER: marker size:",s.markersize,s.strokecolor,s.fillcolor,s.strokewidth)

	// where do we put the rectangles?
	//  - remember to subtract half the size of the shape
	// - rectangle size depends on window size

	for i, x := range s.X {
		markersize := fyne.NewSize(float32(wPer), float32(hPer))

		y := s.Y[i]
		xplot := w * (x - xmin) / (xmax - xmin)
		yplot := h * (ymax - y) / (ymax - ymin)
		// fmt.Println(xplot,yplot)
		if _, ok := s.heatmapShapes[i].(*canvas.Rectangle); ok {
			s.heatmapShapes[i].Resize(markersize)
			s.heatmapShapes[i].Move(fyne.NewPos(float32(xplot)-markersize.Width/2, float32(yplot)-markersize.Height/2))
			// fmt.Println(s.heatmapShapes[i].Position(),s.heatmapShapes[i].(*canvas.Rectangle).FillColor)
		}
	}

	return nil
}

func (s *heatmapplotter) allShapes() []fyne.CanvasObject {
	return s.heatmapShapes
}

func (s *heatmapplotter) allAxes() []axisOrientation {
	return []axisOrientation{s.ox, s.oy}
}

func (s *heatmapplotter) dataRange() [2]axisLimits {
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
