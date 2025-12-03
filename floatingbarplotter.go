package chart

import (
	"fmt"
	"image/color"
	"log"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
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
type floatingbarplotter struct {
	Plotter
	strokecolor, fillcolor color.Color
	strokewidth            float32
	Y2                     []float64
	fillwidth              float64
}

func NewFloatingBarplotter(X, Y1, Y2 []float64, ox, oy axisOrientation) *floatingbarplotter {
	pl := &floatingbarplotter{}
	pl.X = X
	pl.Y = Y1
	pl.Y2 = Y2
	pl.ox = ox
	pl.oy = oy

	// defaults

	pl.strokecolor = theme.Color(theme.ColorNameForeground)
	pl.strokewidth = 0
	pl.fillwidth = 100
	pl.fillcolor = color.RGBA{240, 240, 240, 255}

	return pl
}

func (s *floatingbarplotter) SetFillColor(c color.Color) {
	s.fillcolor = c
}
func (s *floatingbarplotter) SetFillWidth(percent float64) {
	if percent < 5 {
		percent = 5
	}
	// if percent > 100 {
	// 	percent = 100
	// }
	s.fillwidth = percent
}

func (s *floatingbarplotter) SetStrokeColor(c color.Color) {
	s.strokecolor = c
}

func (s *floatingbarplotter) SetStrokeWidth(w float32) {
	s.strokewidth = w
}

func (s *floatingbarplotter) dataRange() [2]axisLimits {
	minx := 1.0e+20
	maxx := -1.0e+20
	miny := 1.0e+20
	maxy := -1.0e+20
	for i, x := range s.X {
		y := s.Y[i]
		z := s.Y2[i]
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
		if z < miny {
			miny = z
		}
		if z > maxy {
			maxy = z
		}
	}
	return [2]axisLimits{{minx, maxx}, {miny, maxy}}
}

func (h *floatingbarplotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

// Makes the list of objects to be drawn by the plot renderer
//
//	In this case, a series of rectangles of fixed width and
//	varying height and vertical position
func (s *floatingbarplotter) positionPlotObjects(cv *plot) error {

	// log.Println("Floating Bar Plotter")

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

	s.shapes = make([]fyne.CanvasObject, len(s.X)*2)
	wg := sync.WaitGroup{}
	wg.Add(len(s.X))
	concurrencyLimiter := make(chan int, 4)
	width := w / (xmax - xmin) * s.fillwidth / 100
	for i := 0; i < len(s.X); i++ {
		go func(i int, wg *sync.WaitGroup) {
			concurrencyLimiter <- 1
			defer wg.Done()
			x1 := s.X[i]
			y1 := s.Y[i]
			y2 := s.Y2[i]
			r := NewRectangle()
			r.value = x1
			r.FillColor = s.fillcolor
			r.StrokeWidth = float32(s.fillwidth)
			r.StrokeColor = s.fillcolor

			if y1 > y2 {
				y1, y2 = y2, y1
			}

			xplot1 := w*(x1-xmin)/(xmax-xmin) - width/2
			yplot1 := h * (ymax - y1) / (ymax - ymin)
			yplot2 := h * (ymax - y2) / (ymax - ymin)

			r.Move(fyne.NewPos(float32(xplot1), float32(yplot2)))
			r.Resize(fyne.NewSize(float32(width), float32(yplot1-yplot2)))

			// dd1 := fyne.CanvasObject(r)
			s.shapes[2*i] = r
			// dd2 := fyne.CanvasObject(&r.Rectangle)
			s.shapes[2*i+1] = &r.Rectangle

			r.Refresh()

			<-concurrencyLimiter
		}(i, &wg)
	}
	wg.Wait()

	return nil
}

func (s *floatingbarplotter) allShapes() []fyne.CanvasObject {
	return s.shapes
}

func (s *floatingbarplotter) allAxes() []axisOrientation {
	return []axisOrientation{s.ox, s.oy}
}
