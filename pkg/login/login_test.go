package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/junanda/shortenerUrl/pkg/model"
	"github.com/junanda/shortenerUrl/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDynamoDB is a mock of DynamoDB service
type MockDynamoDB struct {
	mock.Mock
}

func (m *MockDynamoDB) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func TestHandler(t *testing.T) {
	pass, _ := utils.EncryptPassword("testpass")
	tests := []struct {
		name               string
		request            events.APIGatewayProxyRequest
		mockGetItemOutput  *dynamodb.GetItemOutput
		mockGetItemErr     error
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Valid Login",
			request: events.APIGatewayProxyRequest{
				Body: `{"username":"testuser","password":"testpass"}`,
			},
			mockGetItemOutput: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"username": {S: aws.String("testuser")},
					"password": {S: aws.String(pass)},
				},
			},
			mockGetItemErr:     nil,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "Login Success",
		},
		{
			name: "Invalid Login - User not found",
			request: events.APIGatewayProxyRequest{
				Body: `{"username":"unknown","password":"testpass"}`,
			},
			mockGetItemOutput:  &dynamodb.GetItemOutput{},
			mockGetItemErr:     nil,
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "User not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockDynamoDB)
			mockSvc.On("GetItem", mock.Anything).Return(tt.mockGetItemOutput, tt.mockGetItemErr)

			// Replace svc with mockSvc in your Handler when testing
			response, err := Handler(tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)

			var respBody model.LoginResponse
			err = json.Unmarshal([]byte(response.Body), &respBody)
			assert.NoError(t, err)
			assert.Contains(t, respBody.Message, tt.expectedBody)
		})
	}
}
