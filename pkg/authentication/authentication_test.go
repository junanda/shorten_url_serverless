package main

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/stretchr/testify/assert"
)

// Mock utils function for parsing token
func mockParseToken(token string) (*model.Claims, error) {
	if token == "validToken" {
		return &model.Claims{Role: "member"}, nil
	} else if token == "expiredToken" {
		return nil, &jwt.ValidationError{Errors: jwt.ValidationErrorExpired}
	} else {
		return nil, &jwt.ValidationError{Errors: jwt.ValidationErrorSignatureInvalid}
	}
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name               string
		token              string
		expectedStatusCode int
		expectedPolicy     string
		expectedError      string
	}{
		{
			name:               "Valid Token",
			token:              "Bearer validToken",
			expectedStatusCode: 200,
			expectedPolicy:     "Allow",
			expectedError:      "",
		},
		{
			name:               "Expired Token",
			token:              "Bearer expiredToken",
			expectedStatusCode: 401,
			expectedError:      "Unauthorized, token expired or user has logout",
		},
		{
			name:               "Invalid Token",
			token:              "Bearer invalidToken",
			expectedStatusCode: 401,
			expectedError:      "Unauthorized",
		},
		{
			name:               "No Token Provided",
			token:              "",
			expectedStatusCode: 401,
			expectedError:      "Unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := events.APIGatewayCustomAuthorizerRequest{
				AuthorizationToken: tt.token,
				MethodArn:          "arn:aws:execute-api:us-west-2:*:example/prod/POST/request",
			}

			// Replace utils.ParseToken with mockParseToken in your handler when testing
			// utils.ParseToken = mockParseToken

			response, err := handler(request)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPolicy, response.PolicyDocument.Statement[0].Effect)
			}
		})
	}
}
