package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/junanda/shortenerUrl/utils"
)

func Handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var (
		recoveryReq model.RecoveryRequest
		user        model.User
	)

	err := json.Unmarshal([]byte(request.Body), &recoveryReq)
	if err != nil {
		return utils.ApiResponse(http.StatusBadRequest, "Invalid request body")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Check if the email exists in the member table
	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"email": {
				S: aws.String(recoveryReq.Email),
			},
		},
		TableName: aws.String("member"),
	}

	result, err := svc.GetItem(getInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to get user from DynamoDB")
	}

	if result.Item == nil {
		return utils.ApiResponse(http.StatusNotFound, "Email not found")
	}

	err = dynamodbattribute.UnmarshalMap(result.Item, &user)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal user data")
	}

	// Generate a new random password
	newPassword := utils.GenerateRandomPassword()
	user.Password = newPassword

	// Update the user's password in the member table
	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Could not marshal user data")
	}

	updateInput := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String("member"),
	}

	_, err = svc.PutItem(updateInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Could not update user in DynamoDB")
	}

	// Send the new password to SQS
	sqsSvc := sqs.New(sess)
	messageBody, err := json.Marshal(map[string]string{
		"email":    user.Email,
		"password": newPassword,
	})
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to marshal SQS message")
	}

	sendMessageInput := &sqs.SendMessageInput{
		MessageBody: aws.String(string(messageBody)),
		QueueUrl:    aws.String(os.Getenv("SQS_QUEUE_URL")),
	}

	_, err = sqsSvc.SendMessage(sendMessageInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to send message to SQS")
	}

	return utils.ApiResponse(http.StatusOK, "Password recovery successful")
}

func main() {
	lambda.Start(Handler)
}
