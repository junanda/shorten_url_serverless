package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/junanda/shortenerUrl/utils"
)

func Handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var (
		user model.DataUser
	)

	pathId := request.PathParameters["id"]
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Get the user data from DynamoDB
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(pathId),
			},
		},
		TableName: aws.String("member"),
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		utils.PrintError("Error get data user", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to get user from DynamoDB")
	}

	if result.Item == nil {
		return utils.ApiResponse(404, "User not found")
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		utils.PrintError("Error unmarshalMap DynamoDB", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal user data")
	}

	responseBody, err := json.Marshal(user)
	if err != nil {
		utils.PrintError("Error Marshal", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to marshal user data")
	}

	return utils.ApiResponse(http.StatusOK, string(responseBody))
}

func main() {
	lambda.Start(Handler)
}
