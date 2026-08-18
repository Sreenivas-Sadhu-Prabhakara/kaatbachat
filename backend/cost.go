package backend

import (
	"fmt"
	"sort"
)

// Piece is a required cut length and how many of them the job needs.
type Piece struct {
	Length float64 `json:"length"`
	Count  int     `json:"count"`
}

// Input is a cut-to-length job: stock length, the required pieces, and per-cut labour.
type Input struct {
	StockLength float64 `json:"stockLength"`
	Pieces      []Piece `json:"pieces"`
	CutLabour   float64 `json:"cutLabour"`
}

// Result is the cutting plan.
type Result struct {
	StockUnits   int         `json:"stockUnits"`
	TotalOffcut  float64     `json:"totalOffcut"`
	TotalCuts    int         `json:"totalCuts"`
	LabourCost   float64     `json:"labourCost"`
	Bins         [][]float64 `json:"bins"`         // lengths cut from each stock unit
	SingleLength bool        `json:"singleLength"` // trivial job: one distinct length
}

// Headline is the number of stock units required.
func (r Result) Headline() float64 { return float64(r.StockUnits) }

// Label distinguishes the trivial single-length job from a packed one.
func (r Result) Label() string {
	if r.SingleLength {
		return "single-length"
	}
	return "packed"
}

// Validate reports whether the Input is well formed.
func (in Input) Validate() error {
	if in.StockLength <= 0 {
		return fmt.Errorf("stock length must be positive")
	}
	if len(in.Pieces) == 0 {
		return fmt.Errorf("at least one piece is required")
	}
	for _, p := range in.Pieces {
		if p.Length <= 0 || p.Count <= 0 {
			return fmt.Errorf("piece length and count must be positive")
		}
		if p.Length > in.StockLength {
			return fmt.Errorf("piece length %.2f exceeds stock length", p.Length)
		}
	}
	if in.CutLabour < 0 {
		return fmt.Errorf("cut labour cannot be negative")
	}
	return nil
}

// distinctLengths counts how many different lengths the job asks for.
func distinctLengths(pieces []Piece) int {
	set := map[float64]bool{}
	for _, p := range pieces {
		set[p.Length] = true
	}
	return len(set)
}

// Evaluate plans the cut. For a single distinct length it short-circuits to
// plain ceiling division (no "optimization" framing); otherwise it runs a
// first-fit-decreasing bin-packing heuristic to minimise stock units.
func Evaluate(in Input) Result {
	total := 0
	for _, p := range in.Pieces {
		total += p.Count
	}

	if distinctLengths(in.Pieces) == 1 {
		length := in.Pieces[0].Length
		perStock := int(in.StockLength / length)
		if perStock < 1 {
			perStock = 1
		}
		units := (total + perStock - 1) / perStock
		bins := make([][]float64, units)
		remaining := total
		for i := 0; i < units; i++ {
			n := perStock
			if remaining < n {
				n = remaining
			}
			for j := 0; j < n; j++ {
				bins[i] = append(bins[i], length)
			}
			remaining -= n
		}
		return finalize(in, bins, true)
	}

	// Expand pieces into individual lengths, largest first (FFD).
	var lengths []float64
	for _, p := range in.Pieces {
		for i := 0; i < p.Count; i++ {
			lengths = append(lengths, p.Length)
		}
	}
	sort.Sort(sort.Reverse(sort.Float64Slice(lengths)))

	var bins [][]float64
	var remain []float64
	for _, L := range lengths {
		placed := false
		for i := range bins {
			if remain[i] >= L {
				bins[i] = append(bins[i], L)
				remain[i] -= L
				placed = true
				break
			}
		}
		if !placed {
			bins = append(bins, []float64{L})
			remain = append(remain, in.StockLength-L)
		}
	}
	return finalize(in, bins, false)
}

func finalize(in Input, bins [][]float64, single bool) Result {
	offcut := 0.0
	cuts := 0
	for _, b := range bins {
		used := 0.0
		for _, L := range b {
			used += L
			cuts++
		}
		offcut += in.StockLength - used
	}
	return Result{
		StockUnits:   len(bins),
		TotalOffcut:  offcut,
		TotalCuts:    cuts,
		LabourCost:   float64(cuts) * in.CutLabour,
		Bins:         bins,
		SingleLength: single,
	}
}
