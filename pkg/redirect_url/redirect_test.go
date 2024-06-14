package main

import (
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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

func (m *MockDynamoDB) PutItem(input *dynamodb.PutItemInput) (*dynamodb.PutItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name               string
		request            events.APIGatewayProxyRequest
		mockGetItemOutput  *dynamodb.GetItemOutput
		mockGetItemErr     error
		mockPutItemOutput  *dynamodb.PutItemOutput
		mockPutItemErr     error
		expectedStatusCode int
		expectedLocation   string
	}{
		{
			name: "Valid short URL",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"url": "abc123"},
			},
			mockGetItemOutput: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"short_url": {S: aws.String("abc123")},
					"url":       {S: aws.String("https://example.com")},
				},
			},
			mockGetItemErr:     nil,
			mockPutItemOutput:  &dynamodb.PutItemOutput{},
			mockPutItemErr:     nil,
			expectedStatusCode: http.StatusFound,
			expectedLocation:   "https://example.com",
		},
		{
			name: "Short URL not found",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"url": "nonexistent"},
			},
			mockGetItemOutput:  &dynamodb.GetItemOutput{},
			mockGetItemErr:     nil,
			expectedStatusCode: http.StatusNotFound,
			expectedLocation:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockDynamoDB)
			mockSvc.On("GetItem", mock.Anything).Return(tt.mockGetItemOutput, tt.mockGetItemErr)
			mockSvc.On("PutItem", mock.Anything).Return(tt.mockPutItemOutput, tt.mockPutItemErr)

			// Replace svc with mockSvc in your Handler when testing
			response, err := Handler(tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)
			if tt.expectedLocation != "" {
				assert.Equal(t, tt.expectedLocation, response.Headers["Location"])
			}
		})
	}
}
