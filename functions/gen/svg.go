package main

import (
	"bytes"
	"io/ioutil"
	"math"
	"net/http"
	"text/template"

	"github.com/rs/zerolog/log"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

var (
	SVG_TEMPLATE_SRC = "<svg width=\"{{.Width}}mm\" height=\"{{.Height}}mm\" " +
		"viewBox=\"0 0 {{.Width}} {{.Height}}\" " +
		"xmlns=\"http://www.w3.org/2000/svg\" " +
		"style=\"font-family:Arial;\">{{.Contents}}</svg>"
	SVG_TEMPLATE = template.Must(template.New("main svg").Parse(SVG_TEMPLATE_SRC))

	TEXTBOX_TEMPLATE_SRC = "<text x=\"{{.X}}\" y=\"{{.Y}}\" " +
		"text-anchor=\"middle\" style=\"font-size:{{.FontSize}}\">" +
		"{{.Text}}</text>"
	TEXTBOX_TEMPLATE = template.Must(template.New("textbox").Parse(TEXTBOX_TEMPLATE_SRC))

	PAGE_WIDTH  = 210.
	PAGE_HEIGHT = 297.
)

type TemplateConstructor struct {
	PerHeight int
	PerWidth  int
	TextBoxes []TextBox
}

type TextBox struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func CreateSvg(constructor TemplateConstructor, data []string) (string, error) {
	var svgdata struct {
		Width    float64
		Height   float64
		Contents string
	}
	svgdata.Width = PAGE_WIDTH / float64(constructor.PerWidth)
	svgdata.Height = PAGE_HEIGHT / float64(constructor.PerHeight)
	svgdata.Contents = ""

	for idx, box := range constructor.TextBoxes {
		tsize, err := textSize(data[idx])
		if err != nil {
			log.Error().Err(err).Msg("Error in measuring text")
			return "", err
		}

		var contentsdata struct {
			X        float64
			Y        float64
			FontSize float64
			Text     string
		}
		contentsdata.X = box.X
		contentsdata.Y = box.Y
		contentsdata.Text = data[idx]
		contentsdata.FontSize = math.Min(
			16*box.Width/float64(tsize[0]), box.Height)

		var textbox bytes.Buffer
		err = TEXTBOX_TEMPLATE.Execute(&textbox, contentsdata)
		if err != nil {
			log.Error().Err(err).Msg("Error in executing template")
			return "", err
		}

		svgdata.Contents += textbox.String()
	}

	var output bytes.Buffer
	err := SVG_TEMPLATE.Execute(&output, svgdata)
	if err != nil {
		log.Error().Err(err).Msg("Error in executing template")
		return "", err
	}

	return output.String(), nil
}

func textSize(text string) ([2]int, error) {
	resp, err := http.Get(homeUrl() + "/assets/arial.ttf")
	if err != nil {
		return [2]int{0, 0}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return [2]int{0, 0}, err
	}

	font, err := opentype.ParseReaderAt(bytes.NewReader(body))
	if err != nil {
		return [2]int{0, 0}, err
	}

	var b sfnt.Buffer

	totaladv := fixed.Int26_6(0)
	height := 0

	for _, c := range text {
		idx, err := font.GlyphIndex(&b, c)
		if err != nil {
			continue
		}
		bounds, advance, err := font.GlyphBounds(&b, idx, fixed.Int26_6(1024), 0)
		if err != nil {
			continue
		}
		totaladv += advance
		height = int(math.Max(float64(height),
			float64((bounds.Max.Y - bounds.Min.Y).Round())))
	}

	log.Info().Msgf("length: %v", totaladv.Round())
	log.Info().Msgf("height: %v", height)

	return [2]int{totaladv.Round(), height}, nil
}
