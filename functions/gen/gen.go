package main

import (
	"bytes"
	"errors"

	"io/ioutil"
	"math"
	"net/http"
	"os"
	"text/template"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func init() {
	if os.Getenv("CONTEXT") == "dev" {
		output := zerolog.ConsoleWriter{}
		output.Out = os.Stderr
		log.Logger = log.Output(output)
	}
}

func homeUrl() string {
	if os.Getenv("CONTEXT") == "dev" {
		return "http://localhost:3000"
	} else {
		return os.Getenv("URL")
	}
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

	BOX_TEMPLATE_SRC = "<rect x=\"{{.X}}\" y=\"{{.Y}}\" " +
		"width=\"{{.Width}}\" height=\"{{.Height}}\" fill=\"transparent\" " +
		"style=\"stroke-width:2;stroke:rgb(0,0,0)\"/>"
	BOX_TEMPLATE = template.Must(template.New("rect").Parse(BOX_TEMPLATE_SRC))

	PAGE_WIDTH  = 210.
	PAGE_HEIGHT = 297.
)

type TemplateConstructor struct {
	PerHeight int
	PerWidth  int
	TextBoxes []TextBox
}

type TextBox struct {
	X         float64
	Y         float64
	Width     float64
	Height    float64
	ParamName string
}

func createSvg(
	constructor TemplateConstructor,
	data map[string]string) (string, error) {

	width := PAGE_WIDTH / float64(constructor.PerWidth)
	height := PAGE_HEIGHT / float64(constructor.PerHeight)

	contents := ""
	for _, box := range constructor.TextBoxes {
		text := data[box.ParamName]
		tsize, err := textSize(text)
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
		contentsdata.Text = data[box.ParamName]
		contentsdata.FontSize = math.Min(
			16*box.Width/float64(tsize[0]), box.Height)

		var textbox bytes.Buffer
		err = TEXTBOX_TEMPLATE.Execute(&textbox, contentsdata)
		if err != nil {
			log.Error().Err(err).Msg("Error in executing template")
			return "", err
		}

		contents += textbox.String()
	}

	var svgdata struct {
		Width    float64
		Height   float64
		Contents string
	}

	svgdata.Width = width
	svgdata.Height = height
	svgdata.Contents = contents

	var output bytes.Buffer
	err := SVG_TEMPLATE.Execute(&output, svgdata)

	if err != nil {
		log.Error().Err(err).Msg("Error in executing template")
		return "", err
	}

	return output.String(), nil
}

type Params struct {
	Name   string
	Width  float64
	Height float64
	Font   float64
}

func GenResponse() (string, error) {
	data := map[string]string{"fname": "Ksenya", "lname": "Kosterova"}
	constructor := TemplateConstructor{
		PerWidth:  2,
		PerHeight: 4,
		TextBoxes: []TextBox{{
			X:         PAGE_WIDTH / 4,
			Y:         29.813,
			Width:     92.,
			Height:    26.5,
			ParamName: "fname",
		}, {
			X:         PAGE_WIDTH / 4,
			Y:         60.257,
			Width:     90.476,
			Height:    36.364,
			ParamName: "lname",
		},
		},
	}

	output, err := createSvg(constructor, data)
	if err != nil {
		log.Error().Err(err).Msg("Error in creating svg")
		return "", errors.New("error in creating svg")
	}

	return output, nil
}
