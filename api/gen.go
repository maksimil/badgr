package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	if os.Getenv("VERCEL_ENV") == "development" {
		output := zerolog.ConsoleWriter{}
		output.Out = os.Stderr
		log.Logger = log.Output(output)
	}
}

func Handle(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("method", r.Method).Msg("Handling request")
	switch r.Method {
	case http.MethodGet:
		handle(w, r)
	default:
		log.Error().Msg("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hi")
}
