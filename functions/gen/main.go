package main

import (
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	if os.Getenv("CONTEXT") == "dev" {
		output := zerolog.ConsoleWriter{}
		output.Out = os.Stderr
		log.Logger = log.Output(output)
	}
}

func handler(req events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	resp, err := GenResponse()

	if err != nil {
		return &events.APIGatewayProxyResponse{
			StatusCode: 504,
			Body:       err.Error(),
		}, nil
	}

	return &events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       resp,
	}, nil
}

func main() {
	lambda.Start(handler)
}
