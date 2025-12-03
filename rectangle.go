package chart

import (
	"image/color"

	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
)

type rectangle struct {
	canvas.Rectangle
	highlightColour color.Color
	value           float64

	desktop.Hoverable
	desktop.Mouseable
	desktop.Cursorable
}

// simple floating rectangle for plots
//
//	default has gray fill, black outline and turns yellow on mouseover.
//	These can be changed.
func NewRectangle() *rectangle {

	r := &rectangle{}

	// defaults
	r.highlightColour = theme.Color(theme.ColorNameError)
	r.FillColor = theme.Color(theme.ColorNameBackground)
	r.StrokeWidth = 0
	r.StrokeColor = theme.Color(theme.ColorNameForeground)
	return r
}

func (r *rectangle) MouseIn(e *desktop.MouseEvent) {
	r.highlightColour, r.FillColor = r.FillColor, r.highlightColour
	r.Refresh()
}
func (r *rectangle) MouseMoved(e *desktop.MouseEvent) {
}

func (r *rectangle) MouseOut() {
	r.highlightColour, r.FillColor = r.FillColor, r.highlightColour
	r.Refresh()
}

func (r *rectangle) MouseDown(e *desktop.MouseEvent) {
}

func (r *rectangle) MouseUp() {
}

func (r *rectangle) Cursor() desktop.Cursor {
	return desktop.CrosshairCursor
}

func (r *rectangle) Refresh() {
	r.Rectangle.Refresh()
}
