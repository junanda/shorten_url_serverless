package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        events.APIGatewayProxyRequest
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Valid Request",
			request: events.APIGatewayProxyRequest{
				Body: `{"email":"test@example.com","password":"password123", "username":"anditest"}`,
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "User registered successfully",
		},
		{
			name: "Invalid Email Format",
			request: events.APIGatewayProxyRequest{
				Body: `{"email":"test","password":"password123","username":"test"}`,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid email format",
		},
		{
			name: "Invalid Request Body",
			request: events.APIGatewayProxyRequest{
				Body: `{"email":"test@example.com","password":"","username":"tidy"}`,
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response, err := Handler(test.request)
			assert.NoError(t, err)
			assert.Equal(t, test.expectedStatus, response.StatusCode)
			assert.Contains(t, response.Body, test.expectedBody)
		})
	}
}
