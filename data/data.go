package data

import (
	"math"
	"sort"
)

type Row struct {
	Line                string
	Sequence            int
	Category, Attribute string
	Statistics          Statistic
}

type Rows []Row

type Statistic struct {
	Name                   string
	Min, Max, Mean, SD, Q2 float64
	Q1, Q3                 float64
	NObs, NRej             int
}

func StatsFromList(vals []float64) Statistic {
	sort.Float64s(vals)
	var mean, meansquare float64
	minval := 1.0e+40
	maxval := -1.0e+40
	for _, v := range vals {
		mean += v
		meansquare += v * v
		if v > maxval {
			maxval = v
		}
		if v < minval {
			minval = v
		}
	}
	mean /= float64(len(vals))
	meansquare /= float64(len(vals))

	iQ1 := len(vals) / 4
	iQ2 := len(vals) / 2
	iQ3 := len(vals) / 4 * 3

	stats := Statistic{
		Name: "",
		Min:  minval,
		Max:  maxval,
		Mean: mean,
		SD:   math.Sqrt(meansquare - mean*mean),
		Q1:   vals[iQ1],
		Q2:   vals[iQ2],
		Q3:   vals[iQ3],
	}
	return stats
}
