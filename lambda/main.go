package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

type V1Request struct {
	PuzzleId int  `json:"puzzle_id,string"` // requres `,string`, since input is json-encoded string
	IsSolved bool `json:"is_solved,string"`
}

type V1Response struct {
	Message string `json:"message:"`
}

func ValidateSignature(channelSecret, signature string, body []byte) bool {
	decoded, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		log.Println(err)
		return false
	}
	hash := hmac.New(sha256.New, []byte(channelSecret))
	hash.Write(body)
	return hmac.Equal(decoded, hash.Sum(nil))
}

func HandleLambdaRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("ACCESS_TOKEN"))
	if err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	if !ValidateSignature(os.Getenv("CHANNEL_SECRET"), req.Headers["X-Line-Signature"], []byte(req.Body)) {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf(`{"message":"%s"}`+"\n", linebot.ErrInvalidSignature.Error()),
		}, nil
	}

	puzzleResults := &struct {
		Destination string           `json:"destination"`
		Events      []*linebot.Event `json:"events"`
	}{}
	if err := json.Unmarshal([]byte(req.Body), puzzleResults); err != nil {
		log.Print(err)
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	for _, puzzleResult := range puzzleResults.Events {
		switch m := puzzleResult.Message.(type) {
		case *linebot.TextMessage:
			if _, err = bot.ReplyMessage(puzzleResult.ReplyToken, linebot.NewTextMessage(m.Text)).Do(); err != nil {
				log.Print(err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       fmt.Sprintf(`{"message":"%s"}`+"\n", http.StatusText(http.StatusBadRequest)),
				}, nil
			}
		default:
			log.Print(err)
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       fmt.Sprintf(`{"message":"%s"}`+"\n", http.StatusText(http.StatusBadRequest)),
			}, nil
		}
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(HandleLambdaRequest)
}
