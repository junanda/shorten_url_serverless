package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/junanda/shortenerUrl/utils"
)

func Handler(request events.APIGatewayProxyRequest) (*events.APIGatewayProxyResponse, error) {
	var (
		claim model.Claims
	)

	pathId := request.PathParameters["id"]

	if _, ok := request.RequestContext.Authorizer["claim"]; ok {
		claim = request.RequestContext.Authorizer["claim"].(model.Claims)
	}

	if claim.Role != "admin" {
		return utils.ApiResponse(http.StatusUnauthorized, "Unauthorized: must be authorized by admin")
	}

	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)

	// Delete user data from DynamoDB
	deleteInput := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(pathId),
			},
		},
		TableName: aws.String("member"),
	}

	_, err := svc.DeleteItem(deleteInput)
	if err != nil {
		utils.PrintError("Error Delete Data DynamoDB", err)
		return utils.ApiResponse(http.StatusInternalServerError, "Failed to delete user from DynamoDB")
	}

	return utils.ApiResponse(http.StatusOK, "User deleted successfully")
}

func main() {
	lambda.Start(Handler)
}
