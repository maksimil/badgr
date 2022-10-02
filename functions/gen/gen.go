package main

import (
	"errors"

	"os"

	"github.com/rs/zerolog/log"
)

func homeUrl() string {
	if os.Getenv("CONTEXT") == "dev" {
		return "http://localhost:3000"
	} else {
		return os.Getenv("URL")
	}
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

	output, err := CreateSvg(constructor, data)
	if err != nil {
		log.Error().Err(err).Msg("Error in creating svg")
		return "", errors.New("error in creating svg")
	}

	return output, nil
}
