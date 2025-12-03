package chart

import (
	"errors"
	"log"
	"math"
	"sort"

	"fyne.io/fyne/v2"
)

type CrossplotPair struct {
	V1, V2 float64
}

func autoTick2(a, b float64, minticks, maxticks int) []float64 {
	var f0, f1 float64
	f0 = math.Pow(10, math.Floor(math.Log10(b-a)))
	f1 = f0
	q := a / f1
	r := math.Floor(q)*f1 - f1
	list := []float64{}
	for r <= b {
		if r >= a && r <= b {
			if math.Abs(r) < 1.0e-10 {
				r = 0
			}
			list = append(list, r)
		}
		r += f1
	}
	if len(list) < minticks {
		f1 = f0 / 2
		q := a / f1
		r := math.Floor(q)*f1 - f1
		list = []float64{}
		for r <= b {
			if r >= a && r <= b {
				if math.Abs(r) < 1.0e-10 {
					r = 0
				}
				list = append(list, r)
			}
			r += f1
		}
	} else {

		return list
	}

	if len(list) < minticks {
		f1 = f0 / 5
		q := a / f1
		r := math.Floor(q)*f1 - f1
		list = []float64{}
		for r <= b {
			if r >= a && r <= b {
				if math.Abs(r) < 1.0e-10 {
					r = 0
				}
				list = append(list, r)
			}
			r += f1
		}
	} else {
		return list
	}
	if len(list) < minticks {
		f1 = f0 / 10
		q := a / f1
		r := math.Floor(q)*f1 - f1
		list = []float64{}
		for r <= b {
			if r >= a && r <= b {
				if math.Abs(r) < 1.0e-10 {
					r = 0
				}
				list = append(list, r)
			}
			r += f1
		}
	} else {
		return list
	}
	return list
}

func Percentile(h []float64, percentile float64) float64 {
	if len(h) == 0 || percentile < 0 || percentile > 100 {
		return 0
	}
	sort.Float64s(h)
	n := int(float64(len(h)-1) * percentile / 100)
	return h[n]
}

func QuartilesAndRange(h []float64) ([]float64, error) {
	if len(h) == 0 {
		return nil, errors.New("no data provided to calculate quartiles")
	}
	if h[0] == h[len(h)-1] {
		return nil, errors.New("zero data range")
	}
	sort.Float64s(h)
	q1 := int(float64(len(h)-1) * .25)
	q2 := int(float64(len(h)-1) * .5)
	q3 := int(float64(len(h)-1) * .75)
	log.Println(q1,q3,len(h))

	return []float64{h[0], h[q1], h[q2], h[q3], h[len(h)-1]}, nil
}

// makes Freedman-Diaconis bin limits for a set of numbers
//   - the number of bins increases with the cube root of the size of the data
//   - the number of bins increases with the inter-quartile range of the data
func MakeHistogramBinLimits(h []float64) ([]float64, error) {
	q, err := QuartilesAndRange(h)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	a := (q[3] - q[1]) * 2 / math.Pow(float64(len(h)), .333)
	if a <= 0 || math.IsNaN(a) {
		log.Println("HISTOGRAM: Bad IQR", q[1], q[3])
		return nil, errors.New("histogram bin spacing error")
	}
	D := math.Floor(a/math.Pow(10, math.Floor(math.Log10(a)))+0.5) * math.Pow(10, math.Floor(math.Log10(a)))
	// log.Println(D, a)
	base := math.Floor(q[0]/D) * D
	// log.Printf("Spacing for %d values %.4g as %.4g base %.4g min %.4g max %.4g Q1=%.4g Q3=%.4g\n",len(h),a,D,base,q[0],q[len(q)-1],q[1],q[3])
	bins := []float64{}
	for base < D+q[len(q)-1] {
		bins = append(bins, base)
		base += D
	}
	return bins, nil
}

func PearsonCorrelation(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, errors.New("for correlation, arrays must be of the same length")
	}
	if len(x) == 0 {
		return 0, errors.New("no data to correlate - empty arrays")
	}
	var sxx, syy, sxy, sx, sy float64
	for i := 0; i < len(x); i++ {
		sx += x[i]
		sy += y[i]
		sxx += x[i] * x[i]
		sxy += x[i] * y[i]
		syy += y[i] * y[i]
	}
	n := float64(len(x))
	return (n*sxy - sx*sy) / math.Sqrt((n*sxx-sx*sx)*(n*syy-sy*sy)), nil
}

// converts an angle in degrees to a compass rose direction (one of 8 directions)
func RoseDirectionLabel(f float64) string {
	k := f + 22.5
	k = math.Floor(k / 45)
	K := int(k)
	for K < 0 {
		K += 8
	}
	for K > 7 {
		K -= 8
	}

	return RoseDirections[K]
}

func OrthogonalCircleArc(cx, cy, r float64, a, b float64, smoothness int, from, to string) ([][2]fyne.Position, error) {

	arclocs := make([][2]fyne.Position, 0)

	// arc := make([]fyne.CanvasObject, 0)

	log.Printf("\nProcessing link %s -> %s", from, to)

	xa, ya := cx+r*math.Cos(a), cx+r*math.Sin(a)
	xb, yb := cx+r*math.Cos(b), cx+r*math.Sin(b)

	diff := b - a
	diff = math.Mod(diff, 2*math.Pi)
	diff = math.Abs(diff)
	log.Println(diff)
	if math.Abs(diff-math.Pi) < 1.0e-2 {
		arclocs = append(arclocs, [2]fyne.Position{
			fyne.NewPos(float32(xa), float32(ya)),
			fyne.NewPos(float32(xb), float32(yb)),
		})
		log.Println(arclocs[0])
		return arclocs, nil
	}

	// var ma, mb float64

	var R, CX, CY float64

	m1 := (xa - cx) / (cy - ya) // gradients are defined as being normal to
	m2 := (xb - cx) / (cy - yb) // the radii of the 2 angles under consideration

	if ya == cy { // m1 would be infinite, which is a problem
		CX = xa
	} else if yb == cy { // m2 would be infinite, another problem
		CX = xb
	} else { // neither m1 nor m2 is infinite
		CX = (ya - yb - m1*xa + m2*xb) / (m2 - m1) // fails if m1==m2, but this seems very unlikely
	}

	if ya == cy { // m1 infinite, so use m2 to find CY
		CY = yb + m2*(CX-xb)
	} else {
		CY = ya + m1*(CX-xa)
	}

	R = math.Sqrt((CX-xa)*(CX-xa) + (CY-ya)*(CY-ya))
	// R2 := math.Sqrt((CX-xb)*(CX-xb) + (CY-yb)*(CY-yb))
	// log.Printf("Circle at %5.2f,%5.2f, radius %5.2f", CX, CY, R)

	pa := a + math.Pi/2 // angles from the entry and exit to the centre, rotated
	pb := b - math.Pi/2 // these have gradients m1 and m2 respectively
	dp := pb - pa       // angle from pa to pb - same as b-a + PI

	// log.Printf("a=%5.2f b=%5.2f pa=%5.2f pb=%5.2f b-a=%5.2f dp=%5.2f", a, b, pa, pb, b-a, dp)

	if dp < -math.Pi { // larger than a semicircle and negative (ie clockwise)
		pb += math.Pi * 2 // make it anticlockwise from pa to pb again
		dp = pb - pa
	}
	if dp > math.Pi { // larger than a semicircle but anticlockwise
		pb -= math.Pi * 2 // make it smaller than a semicircle, possibly clockwise
		dp = pb - pa
	}

	if dp < 0 { // if its clockwise, rotate pa and pb 180 degrees
		pa += math.Pi
		pb += math.Pi
		dp = pb - pa
	}

	dp /= float64(smoothness)

	// dp*=.9

	// log.Printf("pa %5.2f pb %5.2f dp %5.3f", pa, pb, dp)
	phi := 0.0
	for i := 0; i < smoothness; i++ {
		phi = dp*float64(i) + pa
		x := CX + R*math.Cos(phi)
		y := CY + R*math.Sin(phi)
		x1 := CX + R*math.Cos(phi+dp)
		y1 := CY + R*math.Sin(phi+dp)
		arclocs = append(arclocs, [2]fyne.Position{fyne.NewPos(float32(x), float32(y)), fyne.NewPos(float32(x1), float32(y1))})
	}
	return arclocs, nil
}
