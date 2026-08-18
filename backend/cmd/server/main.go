// Command server runs the kaatbachat HTTP middleware.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Sreenivas-Sadhu-Prabhakara/kaatbachat/backend"
)

func main() {
	var store backend.Store
	if url := os.Getenv("DATABASE_URL"); url != "" {
		pg, err := backend.NewPostgresStore(context.Background(), url)
		if err != nil {
			log.Printf("history disabled (db connect failed): %v", err)
		} else {
			defer pg.Close()
			store = pg
		}
	} else {
		log.Print("history disabled (DATABASE_URL not set)")
	}
	addr := ":" + envOr("PORT", "8080")
	log.Printf("kaatbachat listening on %s", addr)
	if err := http.ListenAndServe(addr, backend.NewServer(store)); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
