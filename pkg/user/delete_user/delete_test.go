package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name               string
		request            events.APIGatewayProxyRequest
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Unauthorized user",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"id": "123"},
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: map[string]interface{}{
						"claim": model.Claims{Role: "user"},
					},
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedBody:       "Unauthorized: must be authorized by admin",
		},
		{
			name: "Authorized admin",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"id": "123"},
				RequestContext: events.APIGatewayProxyRequestContext{
					Authorizer: map[string]interface{}{
						"claim": model.Claims{Role: "admin"},
					},
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "User deleted successfully",
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
