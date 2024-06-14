package main

import (
	"encoding/json"
	"net/http"
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
		req   model.ShortenURLRequest
		claim model.Claims
	)

	claim = request.RequestContext.Authorizer["claim"].(model.Claims)

	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return utils.ApiResponse(http.StatusBadRequest, "Invalid request body")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Check the maximum shortener URL count for the current month from the member table
	memberTable := "member"
	memberInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(claim.Uid),
			},
		},
		TableName: aws.String(memberTable),
	}

	memberResult, err := svc.GetItem(memberInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to get member data from DynamoDB")
	}

	if memberResult.Item == nil {
		return utils.ApiResponse(http.StatusNotFound, "Member not found")
	}

	var member struct {
		MaxShortMonth int `json:"max_short_month"`
	}
	err = dynamodbattribute.UnmarshalMap(memberResult.Item, &member)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal member data")
	}

	// Count the number of URLs shortened by this user in the current month
	shortURLTable := "shorturl"
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(shortURLTable),
		IndexName: aws.String("UserIDIndex"),
		KeyConditions: map[string]*dynamodb.Condition{
			"userid": {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(req.UserID),
					},
				},
			},
			"createdate": {
				ComparisonOperator: aws.String("BEGINS_WITH"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(time.Now().Format("2006-01")),
					},
				},
			},
		},
	}

	queryResult, err := svc.Query(queryInput)
	if err != nil {
		utils.PrintError("Error DynamoDB Query", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to query short URLs")
	}

	if len(queryResult.Items) >= member.MaxShortMonth {
		return utils.ApiResponse(http.StatusBadRequest, "Monthly shortening limit exceeded")
	}

	// Generate short URL and save to DynamoDB
	shortURL := model.ShortURL{
		IDShort:    utils.GenerateUUID(), // This should be replaced with actual unique ID generation logic
		UserID:     req.UserID,
		URL:        req.URL,
		ShortURL:   utils.ShortenURL(req.URL), // This should be replaced with actual short URL generation logic
		CreateDate: time.Now(),
	}

	av, err := dynamodbattribute.MarshalMap(shortURL)
	if err != nil {
		utils.PrintError("Error MarshalMap", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Could not marshal short URL data")
	}

	putInput := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(shortURLTable),
	}

	_, err = svc.PutItem(putInput)
	if err != nil {
		utils.PrintError("Error Input data DynamoDB", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Could not put short URL item into DynamoDB")
	}

	return utils.ApiResponse(http.StatusOK, "URL shortened successfully")
}

func main() {
	lambda.Start(Handler)
}
