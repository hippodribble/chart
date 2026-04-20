package main

import (
	"fmt"
	"log"
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/hippodribble/chart"
)

func TestSomething(tt *testing.T) {
	fmt.Println("Test complete")
}

var centre *fyne.Container

var ch *chart.Chart

func main() {

	ap := app.New()
	w := ap.NewWindow("chart tester")

	w.SetContent(gui())
	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
	// fmt.Println("Done")

}

func gui() fyne.CanvasObject {

	centre = container.NewStack()
	return container.NewBorder(
		widget.NewToolbar(
			widget.NewToolbarAction(theme.ContentAddIcon(), makeChart),
			widget.NewToolbarAction(theme.DocumentSaveIcon(), saveChart),
		), nil,
		nil, nil,
		centre,
	)
}

func makeChart() {
	N := 2880
	X := make([]float64, N)
	// Y := make([]float64, N)
	labels := make([]string, 24)

	for i := range N {
		X[i] = float64(math.Sin(float64(i) / 4))
		// X[i] = float64(i)
	}
	plotter := chart.NewHeatMapPlotter(X, chart.Bottom, chart.Left, labels, 50, 50, 1, nil)

	ch = chart.NewChart(plotter)
	if ch == nil {
		log.Fatalln("chart created, but nil")
	}

	ch.PadAxes()

	centre.RemoveAll()
	centre.Add(ch)
	centre.Refresh()

	fmt.Println("made chart")

}

func saveChart() {
	if ch == nil {
		return
	}
	ch.Save()
	fmt.Println("chart saved")
}
