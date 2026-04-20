package chart

import (
	"image"
	"image/png"
	"log"
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	eventbus "github.com/dtomasi/go-event-bus/v3"
)

const (
	axismargin float32 = 0.1
)

type Chart struct {
	widget.BaseWidget
	canvas     *plot
	axes       []*axis
	plottables []Plottable
	name       string
	bus        *eventbus.EventBus
}

// a chart has a plot canvas and some axis, but no title - add one in a border layout if required. They just take up space
// the plot canvas listens to changes to the axes, as these can be manually set independently of data
// the canvas has a reference back to the chart (maybe not necessary)
func NewChart(plotters ...Plottable) *Chart {
	ch := &Chart{
		plottables: plotters,
	}

	canvas := newPlot(ch, plotters)
	ch.canvas = canvas

	for o := range []axisOrientation{Bottom, Left, Right, Top} {
		ax := newAxis(axisOrientation(o))
		ax.Hide()
		ch.axes = append(ch.axes, ax)
		ax.addObserver(ch.canvas)
	}
	ch.ResetAxes()
	ch.ExtendBaseWidget(ch)
	return ch
}

func (c *Chart) CreateRenderer() fyne.WidgetRenderer {
	return chartRenderer{c: c}
}

// scan the data and re-estimate the axis ranges so that all data are visible
//
//	As this depends on which axis is being used (left, right y axis etc), this
//	needs to be taken into account
func (c *Chart) ResetAxes() {
	if c.axes == nil {
		return
	}

	var leftrange, rightrange, bottomrange, toprange *axisLimits

	for _, p := range c.plottables {
		limits := p.dataRange()
		for _, ori := range p.allAxes() {
			switch ori {
			case Bottom:
				if bottomrange == nil {
					bottomrange = &limits[0]
				} else {
					if bottomrange.min > limits[0].min {
						bottomrange.min = limits[0].min
					}
					if bottomrange.max < limits[0].max {
						bottomrange.max = limits[0].max
					}
				}
			case Top:
				if toprange == nil {
					toprange = &limits[0]
				} else {
					if toprange.min > limits[0].min {
						toprange.min = limits[0].min
					}
					if toprange.max < limits[0].max {
						toprange.max = limits[0].max
					}
				}
			case Left:
				if leftrange == nil {
					leftrange = &limits[1]
				} else {
					if leftrange.min > limits[1].min {
						leftrange.min = limits[1].min
					}
					if leftrange.max < limits[1].max {
						leftrange.max = limits[1].max
					}
				}
			case Right:
				if rightrange == nil {
					rightrange = &limits[1]
				} else {
					if rightrange.min > limits[1].min {
						rightrange.min = limits[1].min
					}
					if rightrange.max < limits[1].max {
						rightrange.max = limits[1].max
					}
				}
			}
		}
	}

	for _, ax := range c.axes {
		if ax.orientation == Bottom && bottomrange != nil {
			bottomrange = bottomrange.pad(5)
			ax.setLimits(bottomrange.min, bottomrange.max)
			ax.Show()
		}
		if ax.orientation == Top && toprange != nil {
			toprange = toprange.pad(5)
			ax.setLimits(toprange.min, toprange.max)
			ax.Show()
		}
		if ax.orientation == Left && leftrange != nil {
			leftrange = leftrange.pad(5)
			ax.setLimits(leftrange.min, leftrange.max)
			ax.Show()
		}
		if ax.orientation == Right && rightrange != nil {
			rightrange = rightrange.pad(5)
			ax.setLimits(rightrange.min, rightrange.max)
			ax.Show()
		}
	}
}

func (c *Chart) SetName(name string) {
	c.name = name
}

func (c *Chart) SetBus(bus *eventbus.EventBus) {
	c.bus = bus
}

func (c *Chart) PadAxes() {
	if c.axes==nil{return}
	if len(c.axes)==0{return}

	var xaxes, yaxes []*axis

	for _, axis := range c.axes {
		if axis.orientation == Bottom || axis.orientation == Top {
			xaxes = append(xaxes, axis)
		}
		if axis.orientation == Left || axis.orientation == Right {
			yaxes = append(yaxes, axis)
		}
	}

	if len(xaxes) ==0 || len(yaxes) ==0 {
		// log.Println(len(xaxes), s.ox)
		// log.Println(len(yaxes), s.oy)
		log.Fatalln("There appear to be missing axes for the data to be drawn on")
	}

	xaxis := xaxes[0]
	yaxis := yaxes[0]

	xaxis.limits.pad(10)

	// we need to know where the data is as a fraction of the plot size
	xmin := xaxis.limits.min
	xmax := xaxis.limits.max
	ymin := yaxis.limits.min
	ymax := yaxis.limits.max

	xmin -= 0.5
	ymin -= 0.5
	xmax += 0.5
	ymax += 0.5
	xaxes[0].setLimits(xmin, xmax)
	yaxes[0].setLimits(ymin, ymax)

}

func (c *Chart) Save() {

	im := fyne.CurrentApp().Driver().CanvasForObject(c).Capture()

	fx := c.canvas.Position().X / c.Size().Width
	fy := c.canvas.Position().Y / c.Size().Height
	lx := (c.canvas.Position().X + c.canvas.Size().Width) / c.Size().Width
	ly := (c.canvas.Position().Y + c.canvas.Size().Height) / c.Size().Height

	W := im.Bounds().Dx()
	H := im.Bounds().Dy()

	oW := (lx - fx) * float32(W)
	oH := (ly - fy) * float32(H)

	i0 := int(fx * float32(W))
	j0 := int(fy * float32(H))

	IM := image.NewRGBA(image.Rect(0, 0, int(oW), int(oH)))

	for i := 0; i < int(oW); i++ {
		for j := 0; j < int(oH); j++ {
			C := im.At(i0+i, j0+j)
			IM.Set(i, j, C)
		}
	}

	thisdir, _ := os.Getwd()
	// log.Println("Saving to folder", thisdir)
	uri := storage.NewFileURI(thisdir)
	// log.Println(uri.Name())
	dir, _ := storage.ListerForURI(uri)

	win := fyne.CurrentApp().Driver().AllWindows()[0]

	dlg := dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			log.Println(err)
			return
		}
		if writer == nil {
			log.Println("file save cancelled")
			return
		}
		defer writer.Close()
		errval := png.Encode(writer, IM)

		if errval != nil {
			log.Println(err)
			return
		}
	}, win)

	dlg.SetLocation(dir)
	dlg.Show()

}

type chartRenderer struct {
	c *Chart
}

func (p chartRenderer) Destroy() {
	for i := range p.c.plottables {
		p.c.plottables[i] = nil
	}
}

func (p chartRenderer) MinSize() fyne.Size {
	return fyne.NewSize(200, 200)
}

func (p chartRenderer) Refresh() {
}

func (p chartRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{p.c.canvas, p.c.axes[0], p.c.axes[1], p.c.axes[2], p.c.axes[3]}
}

func (p chartRenderer) Layout(size fyne.Size) {

	w := size.Width
	h := size.Height

	p.c.canvas.Move(fyne.NewPos(w*axismargin, h*axismargin))

	// log.Println(w*(1-2*axismargin), h*(1-2*axismargin))

	p.c.canvas.Resize(fyne.NewSize(w*(1-2*axismargin), h*(1-2*axismargin)))
	for _, axis := range p.c.axes {
		if !axis.Visible() {
			continue
		}
		switch axis.orientation {
		case Bottom:
			axis.Move(fyne.NewPos(w*axismargin, h*(1-axismargin)))
			axis.Resize(fyne.NewSize(w*(1-2*axismargin), axismargin*h))
		case Top:
			axis.Move(fyne.NewPos(w*axismargin, 0))
			axis.Resize(fyne.NewSize(w*(1-2*axismargin), axismargin*h))
		case Left:
			axis.Move(fyne.NewPos(0, h*axismargin))
			axis.Resize(fyne.NewSize(w*axismargin, h*(1-2*axismargin)))
		case Right:
			axis.Move(fyne.NewPos(w*(1-axismargin), h*axismargin))
			axis.Resize(fyne.NewSize(w*axismargin, h*(1-2*axismargin)))
		}
	}
}
