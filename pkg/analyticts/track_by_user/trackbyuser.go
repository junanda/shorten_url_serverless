package main

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/junanda/shortenerUrl/utils"
)

func Handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var (
		userID string
	)

	userID = request.PathParameters["id"]
	if userID == "" {
		return utils.ApiResponse(http.StatusBadRequest, "user_id parameter is required")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Query short URLs owned by the user
	shortURLTable := "shorturl"
	filt := expression.Name("iduser").Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to build DynamoDB expression")
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(shortURLTable),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}

	shortURLResult, err := svc.Scan(scanInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to scan DynamoDB for short URLs")
	}

	if len(shortURLResult.Items) == 0 {
		return utils.ApiResponse(http.StatusNotFound, "No short URLs found for the user")
	}

	// Collect all short URL IDs
	var shortURLIDs []string
	for _, item := range shortURLResult.Items {
		shortURLIDs = append(shortURLIDs, *item["id"].S)
	}

	// Query analytics data for the user's short URLs
	analyticsTable := "analytics"
	var operandBuilders []expression.OperandBuilder
	for _, id := range shortURLIDs {
		operandBuilders = append(operandBuilders, expression.Value(id))
	}
	if len(operandBuilders) == 0 {
		return utils.ApiResponse(http.StatusNotFound, "No analytics data found for the given short URLs")
	}
	analyticsFilt := expression.Name("idshort").In(operandBuilders[0], operandBuilders[1:]...)
	analyticsExpr, err := expression.NewBuilder().WithFilter(analyticsFilt).Build()
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to build DynamoDB expression for analytics")
	}

	analyticsScanInput := &dynamodb.ScanInput{
		TableName:                 aws.String(analyticsTable),
		ExpressionAttributeNames:  analyticsExpr.Names(),
		ExpressionAttributeValues: analyticsExpr.Values(),
		FilterExpression:          analyticsExpr.Filter(),
	}

	analyticsResult, err := svc.Scan(analyticsScanInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to scan DynamoDB for analytics")
	}

	// Menghitung jumlah data analitik berdasarkan idshort
	countsByIdShort := make(map[string]int)
	for _, item := range analyticsResult.Items {
		idShort := *item["idshort"].S // asumsikan bahwa 'idshort' adalah key untuk ID dalam item
		countsByIdShort[idShort]++
	}

	responseBody := map[string]interface{}{
		"user_id":    userID,
		"short_urls": shortURLResult.Items,
		"analytics":  analyticsResult.Items,
		"counts":     countsByIdShort, // jumlah data analitik per idshort ke dalam response
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
