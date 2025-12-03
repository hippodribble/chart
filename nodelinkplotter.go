package chart

import (
	"errors"
	"image/color"
	"log"
	"math"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
)

type NodeLink struct {
	From, To LinkNode
	Count    int
}

type LinkNode struct {
	name  string
	angle float64
}

// creates a plottable for a node link diagram
//
//	constructor generates outer circle, labels and arcs, as well as unique nodes and links
type NodeLinkPlotter struct {
	Plotter
	outerCircle        []fyne.CanvasObject   // outer perimeter circle
	linkNodes          []LinkNode            // Nodes on the perimeter to be drawn as labels
	uniqueLinks        []NodeLink            // links between nodes to be drawn as arcs
	arcs               [][]fyne.CanvasObject // the arcs
	arcNormalPositions [][][2]fyne.Position  // to make the arcs from segments, the start and end points of each segment are recorded for a unit circle at the origin
	arcList            []fyne.CanvasObject   // the arcs expressed as a list, to speed up the redraws.
	labels             []fyne.CanvasObject   // the labels
	lastsize           fyne.Size             // the last size of the plot, which can be checked to reduce redrawing (fyne redraws enthusiastically the way I use it!)
}

func NewNodeLinkPlotter(links *[][2]string) *NodeLinkPlotter {
	p := &NodeLinkPlotter{}
	p.makeNodesAndLinks(links)
	p.makeShapes()
	return p
}

// create the nodes and links
func (p *NodeLinkPlotter) makeNodesAndLinks(links *[][2]string) {

	// make unique nodes
	p.linkNodes = make([]LinkNode, 0)
	mapper := make(map[string]int)
	for _, pair := range *links {
		mapper[pair[0]]++
		mapper[pair[1]]++
	}
	N := len(mapper)
	nodenamelist := []string{}
	for k := range mapper {
		nodenamelist = append(nodenamelist, k)
	}
	sort.Strings(nodenamelist)
	log.Println("Found", N, "unique nodes")
	dphi := 2 * math.Pi / float64(N)
	count := 0
	for _, k := range nodenamelist {
		phi := dphi * float64(count)
		count++
		p.linkNodes = append(p.linkNodes, LinkNode{k, phi})
	}

	// make unique links
	tempLinks := make([]NodeLink, 0)
	for _, pair := range *links {
		for i := range p.linkNodes {
			if p.linkNodes[i].name == pair[0] {
				for j := range p.linkNodes {
					if p.linkNodes[j].name == pair[1] {
						log.Println("Found link between", p.linkNodes[i].name, "and", p.linkNodes[j].name)
						tempLinks = append(tempLinks, NodeLink{p.linkNodes[i], p.linkNodes[j], 0})
					}
				}
			}
		}
	}

	// these are not unique links - so reduce them
	uniqueLinks := make(map[NodeLink]int)

	for i := range tempLinks {
		uniqueLinks[tempLinks[i]]++
	}

	p.uniqueLinks = []NodeLink{}
	for k, v := range uniqueLinks {
		log.Println("Unique Link", k.From.name, k.To.name)
		k.Count = v
		p.uniqueLinks = append(p.uniqueLinks, k)
	}
}

func (p *NodeLinkPlotter) makeShapes() {

	log.Println("\nNode Link plot - make shapes")
	// outer circle
	N := 180
	p.outerCircle = make([]fyne.CanvasObject, N)
	for i := range 180 {
		l := canvas.NewLine(theme.Color(theme.ColorNameForeground))
		x, y := math.Cos(float64(i)*math.Pi/90), math.Sin(float64(i)*math.Pi/90)
		l.Position1 = fyne.NewPos(float32(x), float32(y))
		x, y = math.Cos(float64(i+1)*math.Pi/90), math.Sin(float64(i+1)*math.Pi/90)
		l.Position2 = fyne.NewPos(float32(x), float32(y))
		l.StrokeWidth = 2
		p.outerCircle[i] = l
	}

	// Node labels
	for i := range p.linkNodes {
		label := canvas.NewText(p.linkNodes[i].name, theme.Color(theme.ColorNameForeground))
		label.Move(fyne.NewPos(
			float32(math.Cos(p.linkNodes[i].angle)),
			float32(math.Sin(p.linkNodes[i].angle)),
		))
		p.labels = append(p.labels, label)
		log.Println(p.linkNodes[i].name, p.linkNodes[i].angle)
	}

	// Arcs - create the coordinates for a unit circle at the origin - reposition later by scaling
	p.arcs = make([][]fyne.CanvasObject, len(p.uniqueLinks))
	N = 50

	p.arcNormalPositions = [][][2]fyne.Position{}

	for i, link := range p.uniqueLinks {

		arcSegmentPositions, _ := OrthogonalCircleArc(0, 0, 1, link.From.angle, link.To.angle, N, link.From.name, link.To.name)
		p.arcNormalPositions = append(p.arcNormalPositions, arcSegmentPositions)
		arc := make([]fyne.CanvasObject, len(arcSegmentPositions))
		for k, segmentPositions := range arcSegmentPositions {
			segment := canvas.NewLine(color.RGBA{255, 0, 0, 64})
			segment.Position1 = segmentPositions[0]
			segment.Position2 = segmentPositions[1]
			segment.StrokeWidth = float32(link.Count)
			arc[k] = segment // add to single arc
		}
		p.arcs[i] = arc // add single arc to the list of arcs
	}

	for _, arc := range p.arcs {
		p.arcList = append(p.arcList, arc...)
	}

	log.Printf("%3d arcs %3d segments %3d sets of normal positions", len(p.arcs), len(p.arcList), len(p.arcNormalPositions))

}

func (p *NodeLinkPlotter) positionPlotObjects(cv *plot) error {

	if cv.Size() == p.lastsize {
		return nil
	}

	log.Println("Node Link Position")

	p.lastsize = cv.Size()

	var w float32 = float32(cv.Size().Width)
	var h float32 = float32(cv.Size().Height)

	cx := w / 2
	cy := h / 2
	// centre := fyne.NewPos(cx, cy)

	rmax := cx
	if cy < rmax {
		rmax = cy
	}

	dphi := math.Pi / 90

	// position the outer circle
	for i, seg := range p.outerCircle {
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

	// position the perimeter labels
	// N := len(p.labels)
	// dphi = 2 * math.Pi / float64(N)
	for i, label := range p.labels {
		hw, hh := label.MinSize().Components()
		hh /= 2
		hw /= 2
		// log.Println(hw,hh)
		// phi := math.Pi/2 - dphi*float64(i)
		phi := p.linkNodes[i].angle
		label.Move(fyne.NewPos(
			float32(math.Cos(phi)*float64(rmax*1.05))+cx-hw,
			float32(-math.Sin(phi)*float64(rmax*1.05))+cy-hh,
		))
	}

	// position each arc drawn, as 90 segments independent of length (otherwise each one will need to be recreated on every redraw)
	if len(p.arcNormalPositions) != len(p.arcs) {
		log.Println(len(p.arcNormalPositions), len(p.arcs), "are not the same!")
		return errors.New("number of normal acr coordinate pairs needs to match number of arcs in NodeLink chart")
	}

	// log.Println(len(p.arcNormalPositions), len(p.arcs), "should be the same!")

	for i := range p.arcs {
		pi := p.arcNormalPositions[i]
		for j, segment := range p.arcs[i] {
			segment.(*canvas.Line).Position1 = scaleFlipped(pi[j][0], rmax*.9).AddXY(cx, cy)
			segment.(*canvas.Line).Position2 = scaleFlipped(pi[j][1], rmax*.9).AddXY(cx, cy)
		}
	}
	return nil
}

func (p *NodeLinkPlotter) allShapes() []fyne.CanvasObject {
	// log.Println("Node Link Plot - all shapes")
	return append(append(p.outerCircle, p.labels...), p.arcList...)
}

func (s *NodeLinkPlotter) allAxes() []axisOrientation {
	return []axisOrientation{}
}

func (s *NodeLinkPlotter) dataRange() [2]axisLimits {
	return [2]axisLimits{{-1, 1}, {-1, 1}}
}

func (h *NodeLinkPlotter) legendEntry() fyne.CanvasObject {
	return canvas.NewRectangle(color.Transparent)
}

func (p *NodeLinkPlotter) name() string {
	return "Node Link Chart"
}
