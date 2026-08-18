package backend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEvaluate_SingleLengthShortCircuit(t *testing.T) {
	// 10 pieces of 2.0m from 6m stock: 3 per stock -> ceil(10/3)=4 units.
	r := Evaluate(Input{StockLength: 6, Pieces: []Piece{{Length: 2, Count: 10}}, CutLabour: 5})
	if !r.SingleLength {
		t.Fatal("expected single-length short-circuit")
	}
	if r.StockUnits != 4 {
		t.Fatalf("stockUnits=%d want 4", r.StockUnits)
	}
}

func TestEvaluate_PacksMultipleLengths(t *testing.T) {
	// 3x2.3, 5x1.7, 2x0.9 from 6m stock. FFD should beat naive one-per-stock (10).
	r := Evaluate(Input{StockLength: 6, CutLabour: 5, Pieces: []Piece{
		{Length: 2.3, Count: 3}, {Length: 1.7, Count: 5}, {Length: 0.9, Count: 2},
	}})
	if r.SingleLength {
		t.Fatal("should not be single-length")
	}
	if r.StockUnits >= 10 || r.StockUnits < 3 {
		t.Fatalf("stockUnits=%d not a sensible packing", r.StockUnits)
	}
	// Sanity: no bin exceeds stock length.
	for _, b := range r.Bins {
		sum := 0.0
		for _, L := range b {
			sum += L
		}
		if sum > 6+1e-9 {
			t.Fatalf("bin overflows stock: %.2f", sum)
		}
	}
}

func TestValidate(t *testing.T) {
	if err := (Input{StockLength: 6, Pieces: []Piece{{Length: 2, Count: 3}}}).Validate(); err != nil {
		t.Fatalf("valid rejected: %v", err)
	}
	for i, bad := range []Input{
		{StockLength: 0, Pieces: []Piece{{Length: 2, Count: 1}}},
		{StockLength: 6, Pieces: nil},
		{StockLength: 6, Pieces: []Piece{{Length: 8, Count: 1}}}, // longer than stock
	} {
		if err := bad.Validate(); err == nil {
			t.Fatalf("bad %d accepted", i)
		}
	}
}

func TestEvaluateEndpoint(t *testing.T) {
	srv := NewServer(nil)
	body := `{"stockLength":6,"cutLabour":5,"pieces":[{"length":2.3,"count":3},{"length":1.7,"count":5}]}`
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/evaluate", strings.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rec.Code, rec.Body.String())
	}
	var r Result
	json.Unmarshal(rec.Body.Bytes(), &r)
	if r.StockUnits < 1 {
		t.Fatalf("stockUnits=%d", r.StockUnits)
	}
}
