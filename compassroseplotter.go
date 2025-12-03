package chart

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/hippodribble/chart/colors"
)

type compassroseplotter struct {
	Plotter
	strokecolor     color.Color
	strokewidth     float32
	data            map[string]int
	seriesname      string
	cardinalLabels  []fyne.CanvasObject // cardinal directions
	cardinalRadials []fyne.CanvasObject // radial lines to cardinal directions
	outerCircle     []fyne.CanvasObject // r
	sides           []fyne.CanvasObject // sides of sectors
	curves          []fyne.CanvasObject // arcs of sectors
	lastsize        fyne.Size
}

func NewCompassRosePlotter(mapper map[string]int) *compassroseplotter {
	pl := &compassroseplotter{data: mapper}

	// defaults

	pl.strokecolor = color.Black
	pl.strokewidth = 2
	pl.seriesname = " "

	pl.makePlotObjects()

	return pl
}

func (s *LinePlotter) name() string {
	return s.seriesname
}

func (s *compassroseplotter) SetStrokeColor(c color.Color) {
	s.strokecolor = c
}

func (s *compassroseplotter) SetStrokeWidth(w float32) {
	s.strokewidth = w
}

func (s *compassroseplotter) dataRange() [2]axisLimits {
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

func (s *compassroseplotter) allShapes() []fyne.CanvasObject {
	zoid := append(s.outerCircle, s.cardinalLabels...)
	zoid = append(zoid, s.cardinalRadials...)
	zoid = append(zoid, s.sides...)
	zoid = append(zoid, s.curves...)
	return zoid
}

func (s *compassroseplotter) allAxes() []axisOrientation {
	// log.Println("Rose Plot - all axes")
	return []axisOrientation{}
}

func (h *compassroseplotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

func (s *compassroseplotter) makePlotObjects() error {

	// log.Println("Rose Plot - make shapes")

	// we need to know where the data is as a fraction of the plot size

	s.cardinalRadials = []fyne.CanvasObject{}
	s.cardinalLabels = []fyne.CanvasObject{}

	dphi := math.Pi / 180

	// add some labels and radial lines to cardinal points

	for _, dir := range RoseDirections {
		t := canvas.NewText(dir, color.Black)
		t.TextSize = 16
		t.TextStyle.Bold = true

		l := canvas.NewLine(colors.LightGrey)

		s.cardinalRadials = append(s.cardinalRadials, l)

		s.cardinalLabels = append(s.cardinalLabels, t)
	}

	for phi := 0.0; phi < 2*math.Pi; phi += dphi {
		seg := canvas.NewLine(s.strokecolor)
		seg.StrokeColor = colors.LightGrey
		s.outerCircle = append(s.outerCircle, seg)
	}

	// make the rose data objects, which are radial lines and a curve in between (made of tiny line segments)
	for i, _ := range RoseDirections {
		// n := s.data[dir]
		// if n == 0 {
		// 	continue
		// }
		// log.Println("Rose Plot - make sides", i,dir)
		p0 := float64(i)*math.Pi/4 + math.Pi/8
		p0 = math.Pi/2 - p0
		p1 := p0 + math.Pi/4

		// log.Println(p0,p1,dir)

		ccwside := canvas.NewLine(s.strokecolor)
		ccwside.StrokeWidth = s.strokewidth

		// a := fyne.CanvasObject(ccwside)
		s.sides = append(s.sides, ccwside)

		cwside := canvas.NewLine(s.strokecolor)
		cwside.StrokeWidth = s.strokewidth
		s.sides = append(s.sides, cwside)

		// log.Println("Now there are sides", len(s.sides),dir)

		// curved part
		dp := p1 - p0
		for dp < 0 {
			dp += 2 * math.Pi
		}
		for dp > 2*math.Pi {
			dp -= 2 * math.Pi
		}
		dphi := math.Pi / 180
		k := math.Floor(dp / dphi)
		dphi = dp / k

		p1 = p0 + dp

		// log.Println(p0,p1,dir)

		for phi := p0; phi < p1; phi += dphi {
			seg := canvas.NewLine(s.strokecolor)
			seg.StrokeWidth = s.strokewidth * 1.5
			s.curves = append(s.curves, seg)
		}
	}
	// log.Println(len(s.sides), "sides")
	return nil
}

func (s *compassroseplotter) positionPlotObjects(cv *plot) error {

	if cv.Size() == s.lastsize {
		// log.Println("Rose Plot - No change in size")
		return nil
	}
	// log.Println("Rose Plot - Position")
	s.lastsize = cv.Size()

	// we need to know where the data is as a fraction of the plot size

	var w float32 = float32(cv.Size().Width)
	var h float32 = float32(cv.Size().Height)

	cx := w / 2
	cy := h / 2
	center := fyne.NewPos(cx, cy)

	maxval := -1

	for _, dir := range RoseDirections {
		n := s.data[dir]
		if n > maxval {
			maxval = n
		}
	}

	rmax := cx
	if cy < rmax {
		rmax = cy
	}

	scale := rmax / float32(maxval)

	dphi := math.Pi / 180

	// add some labels
	for i := range RoseDirections {
		t := s.cardinalLabels[i].(*canvas.Text)
		t.TextSize = 16
		t.TextStyle.Bold = true
		X := t.MinSize().Width / 2
		Y := t.MinSize().Height / 2

		phi := math.Pi/4*float64(i) - math.Pi/2
		r := rmax * 1.05

		x := float32(math.Cos(phi)*float64(r)) + cx
		y := float32(math.Sin(phi)*float64(r)) + cy

		t.Move(fyne.NewPos(x-X, y-Y))

		l := s.cardinalRadials[i].(*canvas.Line)
		l.Position1 = fyne.NewPos(cx, cy)
		r = rmax * .9

		x = float32(math.Cos(phi)*float64(r)) + cx
		y = float32(math.Sin(phi)*float64(r)) + cy
		l.Position2 = fyne.NewPos(x, y)
	}

	for i, seg := range s.outerCircle {
		phi := dphi * float64(i)
		seg.(*canvas.Line).Position1 = fyne.NewPos(
			float32(math.Cos(phi)*float64(rmax*.9))+cx,
			float32(-math.Sin(phi)*float64(rmax*.9))+cy,
		)
		seg.(*canvas.Line).Position2 = fyne.NewPos(
			float32(math.Cos(phi+dphi)*float64(rmax*.9))+cx,
			float32(-math.Sin(phi+dphi)*float64(rmax*.9))+cy,
		)
	}
	counter := 0

	for i, dir := range RoseDirections {
		// log.Println(i, dir)
		n := s.data[dir]
		if n == 0 {
			// log.Println("ignoring", dir)
			continue
		}
		p0 := float64(i)*math.Pi/4 + math.Pi/8
		p0 = math.Pi/2 - p0
		p1 := p0 + math.Pi/4

		r := scale * float32(n) * .8
		// log.Printf("%7d %.3g  %.3g\n", i, p0/math.Pi*180, p1/math.Pi*180)

		ccwside := s.sides[i].(*canvas.Line)
		ccwside.Position1 = center
		ccwside.Position2 = fyne.NewPos(
			float32(math.Cos(p0)*float64(r))+cx,
			float32(-math.Sin(p0)*float64(r))+cy,
		)
		// a := fyne.CanvasObject(ccwside)
		// log.Println("SIDE", i+len(RoseDirections),len(s.sides))
		cwside := s.sides[i+len(RoseDirections)].(*canvas.Line)
		cwside.Position1 = center
		cwside.Position2 = fyne.NewPos(
			float32(math.Cos(p1)*float64(r))+cx,
			float32(-math.Sin(p1)*float64(r))+cy,
		)
		// b := fyne.CanvasObject(cwside)

		// curved part
		dp := p1 - p0
		for dp < 0 {
			dp += 2 * math.Pi
		}
		for dp > 2*math.Pi {
			dp -= 2 * math.Pi
		}
		dphi := math.Pi / 180
		k := math.Floor(dp / dphi)
		dphi = dp / k

		for phi := p0; phi < p0+dp; phi += dphi {
			seg := s.curves[counter].(*canvas.Line)
			counter++
			seg.Position1 = fyne.NewPos(
				float32(math.Cos(phi)*float64(r))+cx,
				float32(-math.Sin(phi)*float64(r))+cy,
			)
			seg.Position2 = fyne.NewPos(
				float32(math.Cos(phi+dphi)*float64(r))+cx,
				float32(-math.Sin(phi+dphi)*float64(r))+cy,
			)
		}
	}

	return nil
}
