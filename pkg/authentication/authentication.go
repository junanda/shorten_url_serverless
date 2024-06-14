package main

import (
	"errors"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/golang-jwt/jwt"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/junanda/shortenerUrl/utils"
)

func handler(request events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	var (
		bearerToken string
		claim       *model.Claims
	)

	token := request.AuthorizationToken
	tokenSlice := strings.Split(token, " ")

	if len(tokenSlice) > 1 {
		bearerToken = tokenSlice[len(tokenSlice)-1]
	}
	if bearerToken == "" {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
	}

	claim, err := utils.ParseToken(bearerToken)
	if err != nil {
		v, _ := err.(*jwt.ValidationError)
		switch v.Errors {
		case jwt.ValidationErrorSignatureInvalid:
			return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
		case jwt.ValidationErrorExpired:
			return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized, token expired or user has logout")
		default:
			return events.APIGatewayCustomAuthorizerResponse{}, errors.New("Unauthorized")
		}
	}

	return generatePolicy("user", "Allow", request.MethodArn, map[string]interface{}{"claim": claim}), nil
}

func generatePolicy(principalID, effect, resource string, context map[string]interface{}) events.APIGatewayCustomAuthorizerResponse {
	authResponse := events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalID}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}
	authResponse.Context = context
	return authResponse
}

func main() {
	lambda.Start(handler)
}
