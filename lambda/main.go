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
		return false
	}
	hash := hmac.New(sha256.New, []byte(channelSecret))
	hash.Write(body)
	return hmac.Equal(decoded, hash.Sum(nil))
}

func HandleLambdaRequest(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println(req)
	bot, err := linebot.New(os.Getenv("CHANNEL_SECRET"), os.Getenv("ACCESS_TOKEN"))
	if err != nil {
		log.Println("Line bot sdk Initialization failed: ", err)
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	// API Gatewayによりリクエストヘッダー名が変更される可能性があるので、大文字小文字を区別しない。
	// https://developers.line.biz/ja/reference/messaging-api/#request-headers
	var signature string
	if req.Headers["X-Line-Signature"] != "" {
		signature = req.Headers["X-Line-Signature"]
	} else {
		signature = req.Headers["x-line-signature"]
	}

	if !ValidateSignature(os.Getenv("CHANNEL_SECRET"), signature, []byte(req.Body)) {
		log.Println("Signature validation failed: ")
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
		log.Println("Deserializing json failed: ", err)
		return events.APIGatewayProxyResponse{
			Body:       err.Error(),
			StatusCode: 500,
		}, err
	}

	for _, puzzleResult := range puzzleResults.Events {
		switch m := puzzleResult.Message.(type) {
		case *linebot.TextMessage:
			if _, err = bot.ReplyMessage(puzzleResult.ReplyToken, linebot.NewTextMessage(m.Text)).Do(); err != nil {
				log.Println("Replying message failed: ", err)
				return events.APIGatewayProxyResponse{
					StatusCode: http.StatusBadRequest,
					Body:       fmt.Sprintf(`{"message":"%s"}`+"\n", http.StatusText(http.StatusBadRequest)),
				}, nil
			}
		default:
			log.Println("Message type is not text: ", fmt.Sprintf("%T", puzzleResult.Message))
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
