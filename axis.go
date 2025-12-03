package chart

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/chart/data"
)

type axisLimits struct {
	min, max float64
}

func (al axisLimits) pad(percent float64) *axisLimits {
	w := al.max - al.min
	a := al.min - w*percent/100
	b := al.max + w*percent/100
	return &axisLimits{a, b}
}

type axisOrientation int

const (
	Bottom axisOrientation = iota
	Left
	Right
	Top
)

type axis struct {
	widget.BaseWidget
	orientation                axisOrientation
	boundaryColour, fillColour color.Color
	limits                     axisLimits
	majorTicks                 []labelledTick
	rect                       *canvas.Rectangle
	line                       *canvas.Line
	ticklines                  []*canvas.Line
	ticklabels                 []*canvas.Text
	observers                  []axisChangeObserver
}

func newAxis(o axisOrientation) *axis {
	a := &axis{
		orientation:    o,
		boundaryColour: theme.Color(theme.ColorNameForeground),
		fillColour:     theme.Color(theme.ColorNameBackground),
	}
	a.setLimits(0, 10)

	a.rect = canvas.NewRectangle(color.Transparent)
	a.rect.StrokeColor = a.boundaryColour
	a.rect.StrokeWidth = 1
	a.line = canvas.NewLine(a.boundaryColour)
	a.line.StrokeWidth = 1
	a.ExtendBaseWidget(a)
	return a
}

func (a *axis) CreateRenderer() fyne.WidgetRenderer {
	r := axisRenderer{axis: a}
	return &r
}

func (a *axis) setLimits(min, max float64) {
	a.limits.min = min
	a.limits.max = max
	a.majorTicks = a.limits.makeLabelledTicks()
	a.ticklines = make([]*canvas.Line, len(a.majorTicks))
	for i := range a.ticklines {
		a.ticklines[i] = canvas.NewLine(a.boundaryColour)
	}
	a.ticklabels = make([]*canvas.Text, len(a.majorTicks))
	for i := range a.ticklabels {
		t := canvas.NewText(a.majorTicks[i].label, theme.Color(theme.ColorNameForeground))
		t.TextSize = 10
		if a.orientation == Left {
			t.Alignment = fyne.TextAlignTrailing
		}
		if a.orientation == Bottom || a.orientation == Top {
			t.Alignment = fyne.TextAlignCenter
		}
		a.ticklabels[i] = t
	}
	a.notifyObservers(a)
}

type labelledTick struct {
	location float64
	label    string
}

// tick locations are expressed in fractional values along the axis
func (a *axisLimits) makeLabelledTicks() []labelledTick {

	if a.max == a.min {
		return nil
	}

	if a.max < a.min {
		return nil
	}

	tickvals := autoTick2(float64(a.min), float64(a.max), 5, 15)
	labelledTicks := make([]labelledTick, len(tickvals))

	for i, v := range tickvals {
		labelledTicks[i] = labelledTick{location: (v - a.min) / (a.max - a.min), label: fmt.Sprintf("%.5g", v)}
	}

	return labelledTicks
}

type axisRenderer struct {
	axis *axis
}

func (a *axisRenderer) Layout(size fyne.Size) {
	w := size.Width
	h := size.Height
	a.axis.rect.Resize(size)
	a.axis.rect.StrokeColor = a.axis.boundaryColour
	switch a.axis.orientation {
	case Bottom:
		a.axis.line.Position1 = fyne.NewPos(0, 0)
		a.axis.line.Position2 = fyne.NewPos(size.Width, 0)
		for i, tickline := range a.axis.ticklines {
			x := w * float32(a.axis.majorTicks[i].location)
			tickline.Position1 = fyne.NewPos(x, 0)
			tickline.Position2 = fyne.NewPos(x, 5)
		}
		for i, ticklabel := range a.axis.ticklabels {
			x := w * float32(a.axis.majorTicks[i].location)
			ticklabel.Move(fyne.NewPos(x, 10))
		}

	case Top:
		a.axis.line.Position1 = fyne.NewPos(0, size.Height)
		a.axis.line.Position2 = fyne.NewPos(size.Width, size.Height)
		for i, tickline := range a.axis.ticklines {
			x := w * float32(a.axis.majorTicks[i].location)
			tickline.Position1 = fyne.NewPos(x, size.Height)
			tickline.Position2 = fyne.NewPos(x, size.Height-5)
		}
		for i, ticklabel := range a.axis.ticklabels {
			x := w * float32(a.axis.majorTicks[i].location)
			ticklabel.Move(fyne.NewPos(x, size.Height-30))
		}

	case Left:
		a.axis.line.Position1 = fyne.NewPos(size.Width, 0)
		a.axis.line.Position2 = fyne.NewPos(size.Width, size.Height)
		for i, tickline := range a.axis.ticklines {
			y := h * float32(1-a.axis.majorTicks[i].location)
			tickline.Position1 = fyne.NewPos(size.Width, y)
			tickline.Position2 = fyne.NewPos(size.Width-5, y)
		}
		for i, ticklabel := range a.axis.ticklabels {
			y := h * float32(1-a.axis.majorTicks[i].location)
			ticklabel.Move(fyne.NewPos(size.Width-10, y-11))
		}

	case Right:
		a.axis.line.Position1 = fyne.NewPos(0, 0)
		a.axis.line.Position2 = fyne.NewPos(0, size.Height)
		for i, tickline := range a.axis.ticklines {
			y := h * float32(1-a.axis.majorTicks[i].location)
			tickline.Position1 = fyne.NewPos(0, y)
			tickline.Position2 = fyne.NewPos(5, y)
		}
		for i, ticklabel := range a.axis.ticklabels {
			y := h * float32(1-a.axis.majorTicks[i].location)
			ticklabel.Move(fyne.NewPos(10, y-6))
		}
	}
}

func (a *axisRenderer) Refresh() {
}

func (a *axisRenderer) Objects() []fyne.CanvasObject {
	objs := []fyne.CanvasObject{}
	objs = append(objs, a.axis.line)
	for _, o := range a.axis.ticklines {
		objs = append(objs, o)
	}
	for _, o := range a.axis.ticklabels {
		objs = append(objs, o)
	}
	return objs
}

func (a *axisRenderer) Destroy() {

}

func (a *axisRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

// assuming some y-values, get the range of values and fit them to a part of the available y-axis range
//
//	for example, set the data such that the minimum will be displayed at 20%, and the maximum at 95% to leave room at the bottom of a plot, etc.
//	A linear transform that returns scale and offset to be applied to data in a plot
func AutoRangeMean(data data.Rows, umin, umax, H float64) []float32 {
	vmin := 1.0e+20
	vmax := -1.0e+20
	for _, d := range data {
		v := d.Statistics.Mean
		if v < vmin {
			vmin = v
		}
		if v > vmax {
			vmax = v
		}
	}
	// we have the range, now set min to minpercent and max to maxpercent

	out := make([]float32, len(data))
	for i, d := range data {
		v := d.Statistics.Mean
		out[i] = float32(((v-vmin)/(vmax-vmin)*(umax-umin) + umin - 1) * -H)
	}
	return out
}

// autorange the sequence numbers to floats
func AutoRangeSequences(data data.Rows, umin, umax, W float64) []float32 {
	vmin := 1.0e+20
	vmax := -1.0e+20
	for _, d := range data {
		v := float64(d.Sequence)
		if v < vmin {
			vmin = v
		}
		if v > vmax {
			vmax = v
		}
	}
	// we have the range, now set min to minpercent and max to maxpercent

	out := make([]float32, len(data))
	for i, d := range data {
		v := float64(d.Sequence)
		out[i] = float32(((v-vmin)/(vmax-vmin)*(umax-umin) + umin) * W)
	}
	return out
}

type axisChangeObserver interface {
	updateAxis(axis *axis)
}

type axisChangeNotifier interface {
	addObserver(observer axisChangeObserver)
	removeObserver(observer axisChangeObserver)
	notifyObservers(axis *axis)
}

func (a *axis) addObserver(observer axisChangeObserver) {
	a.observers = append(a.observers, observer)
}

func (a *axis) removeObserver(observer axisChangeObserver) {
	for i, o := range a.observers {
		if o == observer {
			a.observers = append(a.observers[:i], a.observers[i+1:]...)
		}
	}
}

func (a *axis) notifyObservers(axis *axis) {
	for _, o := range a.observers {
		o.updateAxis(axis)
	}
}
