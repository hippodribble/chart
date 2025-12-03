package chart

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const LegendRectangleSize float32 = 20

type LegendEntry interface {
	GetLegendEntry() fyne.CanvasObject
}

type LineLegendEntry struct {
	widget.BaseWidget
	Shapes []fyne.CanvasObject
}

func (e *LineLegendEntry) GetLegendEntry() fyne.CanvasObject {
	return e
}

func NewLineLegendEntry(label string, strokewidth float32, strokecolour color.Color) *LineLegendEntry {
	l := &LineLegendEntry{}

	lbl := canvas.NewText(label, theme.Color(theme.ColorNameForeground))
	lbl.TextSize = 12
	lbl.TextStyle.Italic = true
	lbl.TextStyle.Bold = true

	box := canvas.NewRectangle(color.RGBA{220, 220, 220, 255})
	// box.StrokeColor = color.RGBA{192,192,192,255}
	// box.StrokeWidth = 1
	box.SetMinSize(fyne.NewSize(LegendRectangleSize+10, LegendRectangleSize+10))
	box.CornerRadius = 2

	line := canvas.NewLine(strokecolour)
	line.StrokeWidth = strokewidth

	l.Shapes = append(l.Shapes, box)
	l.Shapes = append(l.Shapes, lbl)
	l.Shapes = append(l.Shapes, line)

	l.ExtendBaseWidget(l)
	return l
}

func (l *LineLegendEntry) CreateRenderer() fyne.WidgetRenderer {

	b := container.New(&lineLegendItemLayout{}, l.Shapes...)
	return widget.NewSimpleRenderer(b)
}

type lineLegendItemLayout struct {
}

func (l *lineLegendItemLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	// log.Println("    - Layout() in LineLegendItemLayout")

	for _, o := range objects {

		if lbl, ok := o.(*canvas.Text); ok {
			lbl.Move(fyne.NewPos(LegendRectangleSize+theme.Padding(), LegendRectangleSize/2-10))
		}

		if rect, ok := o.(*canvas.Rectangle); ok {
			rect.Move(fyne.NewPos(rect.StrokeWidth/2, rect.StrokeWidth/2))
			rect.Resize(fyne.NewSize(LegendRectangleSize+rect.StrokeWidth, LegendRectangleSize+rect.StrokeWidth))
			// log.Printf("     Moved legend rectangle to relative position %v with size %v", rect.Position(), rect.Size())
		}

		if line, ok := o.(*canvas.Line); ok {
			line.Position1 = fyne.NewPos(.15*LegendRectangleSize, .85*LegendRectangleSize)
			line.Position2 = fyne.NewPos(.85*LegendRectangleSize, .15*LegendRectangleSize)
		}
	}
}

func (l *lineLegendItemLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fyne.NewSize(LegendRectangleSize+100, LegendRectangleSize)
}
