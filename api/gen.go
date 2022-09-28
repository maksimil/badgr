package api

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func init() {
	if os.Getenv("VERCEL_ENV") == "development" {
		output := zerolog.ConsoleWriter{}
		output.Out = os.Stderr
		log.Logger = log.Output(output)
	}
}

func homeUrl() string {
	if os.Getenv("VERCEL_ENV") == "development" {
		return "http://localhost:3000"
	} else {
		return os.Getenv("VERCEL_URL")
	}
}

func textLength(text string) (int, error) {
	resp, err := http.Get(homeUrl() + "/assets/arial.ttf")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	font, err := opentype.ParseReaderAt(bytes.NewReader(body))
	if err != nil {
		return 0, err
	}

	var b sfnt.Buffer

	totaladv := fixed.Int26_6(0)

	for _, c := range text {
		idx, err := font.GlyphIndex(&b, c)
		if err != nil {
			continue
		}
		_, advance, err := font.GlyphBounds(&b, idx, fixed.Int26_6(1024), 0)
		if err != nil {
			continue
		}
		totaladv += advance
	}

	log.Info().Msgf("length: %s", totaladv)

	return totaladv.Round(), nil
}

var (
	svgtemplate = "<svg width=\"{{.Width}}mm\" height=\"{{.Height}}mm\" " +
		"xmlns=\"http://www.w3.org/2000/svg\" " +
		"style=\"font-family:Arial;\"> " +
		"<text x=\"50%\" y=\"50%\" text-anchor=\"middle\" " +
		"style=\"font-size:{{.Font}}px\">{{.Name}}</text></svg>"
	width  = 210
	height = 297
)

type Params struct {
	Name   string
	Width  float32
	Height float32
	Font   float32
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
	tmpl, err := template.New("svg").Parse(svgtemplate)
	if err != nil {
		log.Error().Err(err).Msg("Err on parsing template")
		http.Error(w, "Err on parsing template", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	data := Params{"aa;sdfKsykaa", float32(width) / 2, float32(height) / 4, 0}

	tlength, err := textLength(data.Name)
	if err != nil {
		log.Error().Err(err).Msg("Error in measuring string length")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
	data.Font = 60.4724409456 * float32(data.Width) / float32(tlength)

	err = tmpl.Execute(&buf, data)
	if err != nil {
		log.Error().Err(err).Msg("Err on executing template")
		http.Error(w, "Err on executing template", http.StatusInternalServerError)
		return
	}

	log.Info().Str("svg", buf.String()).Msg("Created and sent svg")

	fmt.Fprint(w, buf.String())
}
