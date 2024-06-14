package main

import (
	"encoding/json"
	"net/http"
	"strconv"

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
		req     model.ListMembersRequest
		members []model.DataUser
	)

	err := json.Unmarshal([]byte(request.Body), &req)
	if err != nil {
		return utils.ApiResponse(http.StatusBadRequest, "Invalid request body")
	}

	page, err := strconv.Atoi(request.QueryStringParameters["page"])
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(request.QueryStringParameters["limit"])
	if err != nil {
		limit = 10
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	filt := expression.Name("role").Equal(expression.Value(req.Role))
	if req.Subscribe != "" {
		filt = filt.And(expression.Name("subscribe").Equal(expression.Value(req.Subscribe)))
	}

	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to build DynamoDB expression")
	}

	scanInput := &dynamodb.ScanInput{
		TableName:                 aws.String("member"),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		Limit:                     aws.Int64(int64(limit)),
	}

	if page > 1 {
		scanInput.ExclusiveStartKey = map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(request.QueryStringParameters["lastEvaluatedKey"]),
			},
		}
	}

	result, err := svc.Scan(scanInput)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to scan DynamoDB")
	}

	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &members)
	if err != nil {
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to unmarshal DynamoDB scan items")
	}

	totalCount := len(members)
	totalPages := totalCount / limit
	if totalCount%limit != 0 {
		totalPages++
	}

	response := model.ListMembersResponse{
		Data:       members,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	return utils.ApiResponse(http.StatusOK, response)
}

func main() {
	lambda.Start(Handler)
}
