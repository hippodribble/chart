package chart

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

type plot struct {
	widget.BaseWidget
	ch         *Chart
	plottables []Plottable
	lastSize   fyne.Size
}

func newPlot(ch *Chart, pl []Plottable) *plot {
	p := &plot{ch: ch, plottables: pl}
	p.ExtendBaseWidget(p)
	return p
}

func (p *plot) CreateRenderer() fyne.WidgetRenderer {
	return plotRenderer{canvas: p}
}

func (p *plot) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

func (p *plot) updateAxis(axis *axis) {
	// log.Println(axis.orientation, "Axis updated")
}

func (p *plot) MouseIn(e *desktop.MouseEvent) {
	// log.Println("In Plot")
}
func (p *plot) MouseMoved(e *desktop.MouseEvent) {
	// log.Println("Moved in plot",e.Position.X, e.Position.Y)
}

func (p *plot) MouseOut() {
	// log.Println("Mouse out of plot")
}

func (p *plot) MouseDown(e *desktop.MouseEvent) {
	// log.Println("Down in plot",e.Position.X)

}
func (p *plot) MouseUp(e *desktop.MouseEvent) {

}

type XYs [][]float64

type plotRenderer struct {
	canvas *plot
	fyne.WidgetRenderer
}

func (p plotRenderer) Destroy() {
	p.canvas = nil
}

func (p plotRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

func (p plotRenderer) Refresh() {
	log.Println("canvas plot refresh called")
}

func (p plotRenderer) Objects() []fyne.CanvasObject {

	objects := []fyne.CanvasObject{}
	// log.Printf("PlotRenderer - Objects()")

	for i, series := range p.canvas.ch.plottables {

		e := series.legendEntry()
		x := LegendRectangleSize
		y := LegendRectangleSize * float32(i) * 1.1
		e.Move(fyne.NewPos(x, y))
		e.Resize(fyne.NewSize(100, LegendRectangleSize))
		// log.Printf("    legend moved to %v",e.Position())
		objects = append(objects, e)
		// nn:=series.name()
		// log.Printf("  Legend %s",nn)
		for _, sh := range series.allShapes() {

			if sh == nil {
				continue
			}
			objects = append(objects, sh)
		}
	}
	return objects

}

func (p plotRenderer) Layout(size fyne.Size) {

	if p.canvas == nil {
		return
	}

	if len(p.canvas.plottables) == 0 {
		return
	}

	// log.Printf("PlotRenderer - Layout()")

	for i, pl := range p.canvas.plottables {

		if p.canvas.plottables[i] == nil {
			log.Fatalln("empty plotter")
		}

		if p.canvas == nil {

			log.Fatalln("no canvas to draw to!")
		}
		pl.positionPlotObjects(p.canvas)

	}
}
