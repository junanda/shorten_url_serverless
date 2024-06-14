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
	var user model.User
	err := json.Unmarshal([]byte(request.Body), &user)
	if err != nil {
		utils.PrintError("Error Unmarshal request body", err)
		return utils.ApiResponse(http.StatusBadRequest, "Invalid request body")
	}

	_, err = mail.ParseAddress(user.Email)
	if err != nil {
		utils.PrintError("Error Check Email", err)
		return utils.ApiResponse(http.StatusBadRequest, "Invalid email format")
	}

	passEncrypt, _ := utils.EncryptPassword(user.Password)

	user.RegisterDate = time.Now()
	user.UpdateDate = time.Now()
	user.UserSubsscribe = "free"
	user.MaxShortUrl = 8
	user.IdUser = utils.GenerateUUID()
	user.Role = "member"
	user.Password = passEncrypt

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	av, err := dynamodbattribute.MarshalMap(user)
	if err != nil {
		utils.PrintError("Error MarshalMap DynamoDB", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Could not marshal user data")
	}

	tableName := "member"
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		utils.PrintError("Error Input Data", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Could not put item into Database")
	}

	return utils.ApiResponse(http.StatusOK, "User registered successfully")
}

func main() {
	lambda.Start(Handler)
}
