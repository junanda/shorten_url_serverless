package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestListUserHandler(t *testing.T) {
	tests := []struct {
		name               string
		request            events.APIGatewayProxyRequest
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Authorized admin",
			request: events.APIGatewayProxyRequest{
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: map[string]interface{}{
						"claim": model.Claims{Role: "admin"},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "List of users retrieved successfully",
		},
		{
			name: "Unauthorized user",
			request: events.APIGatewayProxyRequest{
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: map[string]interface{}{
						"claim": model.Claims{Role: "user"},
					},
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Unauthorized: must be authorized by admin",
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
