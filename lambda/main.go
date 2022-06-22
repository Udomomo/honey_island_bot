package main

import (
	"fmt"

	"github.com/aws/aws-lambda-go/lambda"
)

type V1Request struct {
	PuzzleId int  `json:"puzzle_id"`
	IsSolved bool `json:"is_solved"`
}

type V1Response struct {
	Message string `json:"message:"`
}

func HandleLambdaEvent(request V1Request) (V1Response, error) {
	if request.IsSolved {
		return V1Response{Message: fmt.Sprintf("%dに正解しました", request.PuzzleId)}, nil
	}
	return V1Response{Message: fmt.Sprintf("%dをスキップしました", request.PuzzleId)}, nil
}

func main() {
	lambda.Start(HandleLambdaEvent)
}
