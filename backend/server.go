package backend

import (
	"encoding/json"
	"io"
	"net/http"
)

// NewServer wires POST /evaluate and GET /history. store may be nil (no DB).
// Each calc app defines Input, Result, Evaluate(Input) Result, and
// Result.Headline()/Result.Label() for persistence.
func NewServer(store Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/evaluate", evaluateHandler(store))
	mux.HandleFunc("/history", historyHandler(store))
	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func evaluateHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			writeErr(w, http.StatusBadRequest, "could not read body")
			return
		}
		var in Input
		if err := json.Unmarshal(raw, &in); err != nil {
			writeErr(w, http.StatusBadRequest, "invalid json")
			return
		}
		if err := in.Validate(); err != nil {
			writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		res := Evaluate(in)
		if store != nil {
			_, _ = store.Save(Record{Input: raw, Headline: res.Headline(), Label: res.Label()})
		}
		writeJSON(w, http.StatusOK, res)
	}
}

func historyHandler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeErr(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if store == nil {
			writeErr(w, http.StatusServiceUnavailable, "history unavailable: no database configured")
			return
		}
		items, err := store.List(50)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, "could not read history")
			return
		}
		writeJSON(w, http.StatusOK, items)
	}
}
