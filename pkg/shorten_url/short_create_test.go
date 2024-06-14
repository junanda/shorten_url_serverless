package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestCreateShortURLHandler(t *testing.T) {
	tests := []struct {
		name               string
		request            events.APIGatewayProxyRequest
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Valid URL creation",
			request: events.APIGatewayProxyRequest{
				Body: `{"url": "https://example.com"}`,
			},
			expectedStatusCode: http.StatusOK,
			expectedBody:       "Short URL created successfully",
		},
		{
			name: "Invalid URL creation",
			request: events.APIGatewayProxyRequest{
				Body: `{"url": "not_a_valid_url"}`,
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedBody:       "Invalid URL provided",
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
