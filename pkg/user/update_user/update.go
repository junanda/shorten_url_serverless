package main

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"time"

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
		updateUserReq model.UpdateUserRequest
		user          model.User
	)

	err := json.Unmarshal([]byte(request.Body), &updateUserReq)
	if err != nil {
		return utils.ApiResponse(http.StatusBadRequest, "Invalid request body")
	}

	_, err = mail.ParseAddress(updateUserReq.Email)
	if err != nil {
		return utils.ApiResponse(http.StatusBadRequest, "Invalid email format")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Get the current user data from DynamoDB
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(updateUserReq.ID),
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
		return utils.ApiResponse(http.StatusNotFound, "User not found")
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		utils.PrintError("Error unmarshalMap", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal user data")
	}

	// Update user data
	user.Username = updateUserReq.Username
	user.Email = updateUserReq.Email
	user.UpdateDate = time.Now()

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		utils.PrintError("Error marshalMap", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Could not marshal user data")
	}

	updateInput := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("member"),
	}

	_, err = svc.PutItem(updateInput)
	if err != nil {
		utils.PrintError("Error update data", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Could not update user in DynamoDB")
	}

	return utils.ApiResponse(http.StatusOK, "User updated successfully")
}

func main() {
	lambda.Start(Handler)
}
