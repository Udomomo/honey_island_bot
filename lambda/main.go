package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type V1Request struct {
	PuzzleId int  `json:"puzzle_id,string"` // requres `,string`, since input is json-encoded string
	IsSolved bool `json:"is_solved,string"`
}

type V1Response struct {
	Message string `json:"message:"`
}

func ExtractRequestBody(input string) (*V1Request, error) {
	var req V1Request
	err := json.Unmarshal([]byte(input), &req)
	if err != nil {
		return nil, err
	}
	return &req, nil
}

func HandleLambdaEvent(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	body, err := ExtractRequestBody(req.Body)

	if err != nil {
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	if body.IsSolved {
		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("%dに正解しました", body.PuzzleId),
			StatusCode: 200,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       fmt.Sprintf("%dをスキップしました", body.PuzzleId),
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
