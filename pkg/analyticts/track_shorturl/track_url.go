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
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/junanda/shortenerUrl/utils"
)

func Handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var (
		shortURL  string
		count     int
		shortData model.ShortURL
	)

	shortURL = request.PathParameters["url"]
	if shortURL == "" {
		return utils.ApiResponse(http.StatusBadRequest, "short_url parameter is required")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Retrieve the shortened URL data from the DynamoDB table based on the short_url provided
	getItemInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"short_url": {
				S: aws.String(shortURL),
			},
		},
		TableName: aws.String("shorturl"),
	}

	getResult, err := svc.GetItem(getItemInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to get short URL data from DynamoDB")
	}

	if getResult.Item == nil {
		return utils.ApiResponse(http.StatusNotFound, "Short URL not found")
	}

	err = dynamodbattribute.UnmarshalMap(getResult.Item, &shortData)
	if err != nil {
		utils.PrintError("Error unmarshalMap", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal user data")
	}

	filt := expression.Name("idshort").Equal(expression.Value(shortData.IDShort))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to build DynamoDB expression")
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String("shorturl"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	result, err := svc.Scan(scanInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to scan DynamoDB")
	}

	count = len(result.Items)

	responseBody := map[string]interface{}{
		"short_url":    shortURL,
		"access_count": count,
	}

	response, err := json.Marshal(responseBody)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to marshal response")
	}

	return utils.ApiResponse(http.StatusOK, string(response))
}

func main() {
	lambda.Start(Handler)
}
