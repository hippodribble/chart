package chart

import (
	"fyne.io/fyne/v2"
)

type Plottable interface {
	positionPlotObjects(*plot) error // creates and positions the shapes that realise the plot
	allShapes() []fyne.CanvasObject  // returns the list of shapes created (because the underlying Plotter can often not be accessed)
	allAxes() []axisOrientation      // returns the orientation of axes that are used (because the underlying Plotter can often not be accessed)
	dataRange() [2]axisLimits
	legendEntry() fyne.CanvasObject // gets the legend entry that should accompany the series in a plot - could be an empty rectangle if a chart type does not implement it or need it.
	name() string                   // typically the name of the series
}

type markertype int

const (
	None markertype = iota
	Circle
	Square
)

type Plotter struct {
	X, Y       []float64           // data in world coordinates to be plotted
	ox, oy     axisOrientation     // for the x and y axes respectively
	shapes     []fyne.CanvasObject // primitive shapes that are drawn to realise the plot
	seriesname string              // name of the series
	Plottable
}
