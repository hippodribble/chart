package chart

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"github.com/hippodribble/chart/colors"
)

type TernaryPoints struct {
	Labels []string
	Values [][3]float64
}

type TernaryPlotter struct {
	Plotter
	DataPoints          [][3]float64        // raw data points
	munged              [][3]float32        // data as fractions
	seriesname          string              // name of series
	Aname, Bname, Cname string              // axis names
	pointcolor          color.Color         // color of line
	pointsize           float32             // width of line
	points              []fyne.CanvasObject // points to be plotted
	triangle            []fyne.CanvasObject // outer triangle
	gridlines           []fyne.CanvasObject // internal grid lines
	titles              []fyne.CanvasObject // Vertex titles
	labeldata           []string            // text of labels associated with each point
	labels              []fyne.CanvasObject // canvas elements showing labels
	showlabels          bool                // whether to show labels
	lastsize            fyne.Size           // maybe we can cut back on force-directed label positioning
}

func NewTernaryPlotter(data TernaryPoints, name string, Aname, Bname, Cname string) *TernaryPlotter {
	pl := &TernaryPlotter{
		seriesname: name,
		pointcolor: theme.Color(theme.ColorNameForeground),
		pointsize:  5,
		Aname:      Aname,
		Bname:      Bname,
		Cname:      Cname,
		labeldata:  data.Labels,
		DataPoints: data.Values,
	}

	pl.munge()
	pl.makeFrame()

	return pl
}

func (pl *TernaryPlotter) makeFrame() {
	pl.points = make([]fyne.CanvasObject, len(pl.DataPoints))
	pl.labels = make([]fyne.CanvasObject, len(pl.DataPoints))
	pl.triangle = make([]fyne.CanvasObject, 3)
	pl.titles = make([]fyne.CanvasObject, 3)
	pl.gridlines = make([]fyne.CanvasObject, 27)

	for i := range pl.DataPoints {
		label := canvas.NewText(pl.labeldata[i], theme.Color(theme.ColorNameForeground))
		label.TextSize = 10
		label.TextStyle.Bold = true
		label.TextStyle.Italic = true
		label.Alignment = fyne.TextAlignCenter
		pl.labels[i] = label
		pl.points[i] = canvas.NewCircle(pl.pointcolor)
		pl.points[i].Resize(fyne.NewSize(pl.pointsize, pl.pointsize))
	}

	for i := 0; i < 3; i++ {
		l := canvas.NewLine(color.Black)
		l.StrokeWidth = 2
		pl.triangle[i] = l
		pl.titles[i] = canvas.NewText("X", theme.Color(theme.ColorNameForeground))
	}

	for i := 0; i < 27; i++ {
		l := canvas.NewLine(colors.LightGrey)
		l.StrokeWidth = 1
		pl.gridlines[i] = l
	}

	pl.titles[0].(*canvas.Text).Text = pl.Aname
	pl.titles[1].(*canvas.Text).Text = pl.Bname
	pl.titles[2].(*canvas.Text).Text = pl.Cname
	pl.titles[0].(*canvas.Text).Alignment = fyne.TextAlignCenter
	pl.titles[1].(*canvas.Text).Alignment = fyne.TextAlignLeading
	pl.titles[2].(*canvas.Text).Alignment = fyne.TextAlignTrailing

	for i := 0; i < 3; i++ {
		pl.titles[i].(*canvas.Text).TextSize = 16
		pl.titles[i].(*canvas.Text).TextStyle.Bold = true
	}
}

// converts raw data to fractions of total
func (s *TernaryPlotter) munge() {
	s.munged = make([][3]float32, len(s.DataPoints))
	for i, x := range s.DataPoints {
		sum := x[0] + x[1] + x[2]
		s.munged[i] = [3]float32{float32(x[0] / sum), float32(x[1] / sum), float32(x[2] / sum)}
	}
}

func (s *TernaryPlotter) name() string {
	return s.seriesname
}

func (s *TernaryPlotter) SetShowLabels(b bool) {
	s.showlabels = b
}

func (s *TernaryPlotter) SetPointColor(c color.Color) {
	s.pointcolor = c
	for _, p := range s.points {
		p.(*canvas.Circle).FillColor = c
	}
}

func (s *TernaryPlotter) SetPointSize(w float32) {
	s.pointsize = w
	for _, p := range s.points {
		p.Resize(fyne.NewSize(w, w))
	}
}

func (s *TernaryPlotter) dataRange() [2]axisLimits {
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
	return [2]axisLimits{{-1, 1}, {-1, 1}}
}

func (s *TernaryPlotter) allShapes() []fyne.CanvasObject {

	return append(append(append(append(s.gridlines, s.triangle...), s.points...), s.labels...), s.titles...)
}

func (s *TernaryPlotter) allAxes() []axisOrientation {
	return []axisOrientation{}
}

func (h *TernaryPlotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

func (s *TernaryPlotter) positionPlotObjects(cv *plot) error {
	// we need to know where the data is as a fraction of the plot size

	if cv.Size() == s.lastsize {
		return nil
	}

	s.lastsize = cv.Size()

	var w float32 = float32(cv.Size().Width)
	var h float32 = float32(cv.Size().Height)

	cx := w / 2
	cy := h / 2
	centre := fyne.NewPos(cx, cy)

	s.shapes = []fyne.CanvasObject{}

	r32 := float32(math.Sqrt(3) / 2)

	// make the outer triangle - base it on available width and height
	var W, H float32
	if w > h*r32 {
		H = h
		W = H / r32
	} else {
		W = w
		H = W * r32
	}

	// locate the outer triangle
	A := fyne.NewPos(cx, cy-.4*H)
	B := fyne.NewPos(cx-W/2, cy+.4*H)
	C := fyne.NewPos(cx+W/2, cy+.4*H)
	s.triangle[0].(*canvas.Line).Position1 = A
	s.triangle[0].(*canvas.Line).Position2 = B
	s.triangle[1].(*canvas.Line).Position1 = B
	s.triangle[1].(*canvas.Line).Position2 = C
	s.triangle[2].(*canvas.Line).Position1 = C
	s.triangle[2].(*canvas.Line).Position2 = A

	// axis gridlines - weighted fraction of distance along triangle edges
	for i := 0; i < 9; i++ {
		perc := float32(i+1) * .1
		// three lines can be drawn
		// vector AB
		p0 := B.Subtract(A)
		p0.X *= perc
		p0.Y *= perc
		p1 := C.Subtract(A)
		p1.X *= perc
		p1.Y *= perc
		s.gridlines[int(3*i)].(*canvas.Line).Position1 = p0.Add(A)
		s.gridlines[int(3*i)].(*canvas.Line).Position2 = p1.Add(A)
		// b := fyne.CanvasObject(s.innerlines[int(3*i)])

		p0 = C.Subtract(B)
		p0.X *= perc
		p0.Y *= perc
		p1 = A.Subtract(B)
		p1.X *= perc
		p1.Y *= perc
		s.gridlines[int(3*i)+1].(*canvas.Line).Position1 = p0.Add(B)
		s.gridlines[int(3*i)+1].(*canvas.Line).Position2 = p1.Add(B)
		// bb := fyne.CanvasObject(s.innerlines[int(3*i)+1])
		s.shapes = append(s.shapes, s.gridlines[int(3*i)+1])

		p0 = A.Subtract(C)
		p0.X *= perc
		p0.Y *= perc
		p1 = B.Subtract(C)
		p1.X *= perc
		p1.Y *= perc
		s.gridlines[int(3*i)+2].(*canvas.Line).Position1 = p0.Add(C)
		s.gridlines[int(3*i)+2].(*canvas.Line).Position2 = p1.Add(C)
		// bc := fyne.CanvasObject(s.innerlines[int(3*i)+2])

	}

	// outer triangle drawn on the gridlines

	// plot the points on the triangle (just a weighted fraction of the vertex locations)
	for i := range s.points {
		fA := s.munged[i][0]
		fB := s.munged[i][1]
		fC := s.munged[i][2]
		xa, ya := A.Components()
		xb, yb := B.Components()
		xc, yc := C.Components()
		p := fyne.NewPos(fA*xa+fB*xb+fC*xc, fA*ya+fB*yb+fC*yc)
		p = p.AddXY(s.pointsize/2, s.pointsize/2)

		s.points[i].Move(p.Subtract(s.points[i].Size()))
		s.labels[i].Move(s.points[i].Position())

		// log.Printf("%.1f,%.1f  %.1f,%.1f  %.1f,%.1f - %.1f,%.1f fractions %.2f, %.2f, %.2f,",A.X,A.Y,B.X,B.Y,C.X,C.Y,s.points[i].Position().X,s.points[i].Position().Y,fA,fB,fC)
		// q := fyne.CanvasObject(s.points[i])
	}

	// axis titles - extend a line from centre to each vertex, put label there.
	da := A.Subtract(centre)
	var k float32 = 1.05
	// sz:=s.titles[0].Size()
	da.X *= k
	da.Y *= k
	da.Y -= 20
	// da.X-=sz.Width/2
	// da.Y-=sz.Height/2
	da = da.Add(centre)
	s.titles[0].Move(A.SubtractXY(0, 30))
	// b0 := fyne.CanvasObject(s.titles[0])

	// sz=s.titles[1].Size()
	db := B.Subtract(centre)
	db.X *= k
	db.Y *= k
	db = db.Add(centre)
	s.titles[1].Move(B.AddXY(0, 10))
	// b1 := fyne.CanvasObject(s.titles[1])

	// sz=s.titles[2].Size()
	dc := C.Subtract(centre)
	dc.X *= k
	dc.Y *= k
	// dc.X-=sz.Width/2
	// dc.Y-=sz.Height/2
	dc = dc.Add(centre)
	s.titles[2].Move(C.AddXY(0, 10))
	// b2 := fyne.CanvasObject(s.titles[2])

	// label text - this stuff will be positioned via an algorithm to avoid overlap
	s.MoveLabels()

	return nil
}
func repel(x float32) float32 {
	return 5 / x / x
}
func norm(p fyne.Position) fyne.Position {
	r := float32(math.Sqrt(float64(p.X)*float64(p.X) + float64(p.Y)*float64(p.Y)))
	return fyne.NewPos(p.X/r, p.Y/r)
}
func length(p fyne.Position) float32 {
	r := float32(math.Sqrt(float64(p.X)*float64(p.X) + float64(p.Y)*float64(p.Y)))
	return r
}

func scale(p fyne.Position, k float32) fyne.Position {
	p.X *= k
	p.Y *= k
	return p
}
func scaleFlipped(p fyne.Position, k float32) fyne.Position {
	p.X *= k
	p.Y *= -k
	return p
}

func (s *TernaryPlotter) MoveLabels() {

	// labelpos := make([]fyne.Position, len(s.DataPoints))
	NettForce := make([]fyne.Position, len(s.DataPoints))

	for i := range s.DataPoints {
		s.labels[i].Move(s.points[i].Position())
		// labelpos[i] = s.points[i].Position()
	}

	// t := float32(100)
	offset := float32(15)

	for iter := 0; iter < 100; iter++ {

		for i := range s.DataPoints {
			for j := range s.DataPoints {
				if i == j {
					continue
				}
				d := s.labels[i].Position().Subtract(s.labels[j].Position())
				NettForce[i] = NettForce[i].Add(scale(norm(d), repel(length(d))))
			}
		}

		for i := range s.points {

			s.labels[i].Move(s.labels[i].Position().Add(NettForce[i]))
			pointToLabelDistance := s.labels[i].Position().Subtract(s.points[i].Position())
			pointToLabelVector := scale(norm(pointToLabelDistance), offset)
			s.labels[i].Move(s.points[i].Position().Add(pointToLabelVector))
			s.labels[i].Move(s.labels[i].Position().AddXY(0, -1))
		}

		// t *= 0.9
	}
}
