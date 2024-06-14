package main

import (
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
	shortID := request.PathParameters["url"]
	if shortID == "" {
		return utils.ApiResponse(http.StatusBadRequest, "Missing shortid parameter")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Retrieve the original URL from the shorturl table
	shortURLTable := "shorturl"
	getItemInput := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"short_url": {
				S: aws.String(shortID),
			},
		},
		TableName: aws.String(shortURLTable),
	}

	result, err := svc.GetItem(getItemInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to get short URL data from DynamoDB")
	}

	if result.Item == nil {
		return utils.ApiResponse(http.StatusNotFound, "Short URL not found")
	}

	var shortURL model.ShortURL
	err = dynamodbattribute.UnmarshalMap(result.Item, &shortURL)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal short URL data")
	}

	// Save analytics data
	analyticsTable := "analytics"
	analytics := model.Analytics{
		AnalyticId: utils.GenerateUUID(),
		ShortUrlId: shortURL.IDShort,
		Browser:    request.Headers["User-Agent"],
		IpRequest:  request.RequestContext.Identity.SourceIP,
		AccessDate: time.Now(),
	}

	av, err := dynamodbattribute.MarshalMap(analytics)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Could not marshal analytics data")
	}

	putItemInput := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(analyticsTable),
	}

	_, err = svc.PutItem(putItemInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Could not put analytics item into DynamoDB")
	}

	// Redirect to the original URL
	return &events.APIGatewayProxyResponse{
		StatusCode: http.StatusFound,
		Headers: map[string]string{
			"Location": shortURL.URL,
		},
	}, nil
}

func main() {
	lambda.Start(Handler)
}
