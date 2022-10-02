package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

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
