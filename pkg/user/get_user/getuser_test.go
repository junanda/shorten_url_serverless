package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestGetUserHandler(t *testing.T) {
	tests := []struct {
		name               string
		request            events.APIGatewayProxyRequest
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Valid user request",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"id": "123"},
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: map[string]interface{}{
						"claim": model.Claims{Role: "user"},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "User data retrieved successfully",
		},
		{
			name: "Unauthorized user request",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"id": "123"},
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: map[string]interface{}{
						"claim": model.Claims{Role: "guest"},
					},
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Unauthorized: must be authorized by user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := Handler(tt.request)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			assert.Equal(t, tt.expectedBody, response.Body)
		})
	}
}
