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
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/junanda/shortenerUrl/utils"
)

var (
	secretKey = []byte(os.Getenv("JWT_SECRET"))
)

func Handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var (
		dataRequest  model.RequestLogin
		user         model.User
		responseBody model.LoginResponse
	)

	err := json.Unmarshal([]byte(request.Body), &dataRequest)
	if err != nil {
		utils.PrintError("Error parsing body data", err)
		return utils.ApiResponse(http.StatusBadRequest, "Invalid request body")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	getInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(dataRequest.Username),
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
		utils.PrintError("Error unmarshalMap DynamoDB", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal user data")
	}

	if !utils.CompareHashPassword(dataRequest.Password, user.Password) {
		return utils.ApiResponse(http.StatusUnauthorized, "Invalid username or password")
	}

	// create token JWT
	token, err := utils.GenerateToken(user, secretKey)
	if err != nil {
		utils.PrintError("Error Generate Token", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to generate token")
	}

	responseBody.Message = "Login Success"
	responseBody.Token = token

	response, err := json.Marshal(responseBody)
	if err != nil {
		utils.PrintError("Error Marshal", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to marshal user data")
	}

	return utils.ApiResponse(http.StatusOK, string(response))
}

func main() {
	lambda.Start(Handler)
}
