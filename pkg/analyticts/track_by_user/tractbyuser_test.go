package main

import (
	"encoding/json"
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

func (m *MockDynamoDB) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.ScanOutput), args.Error(1)
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name                    string
		request                 events.APIGatewayProxyRequest
		mockScanOutputShortURLs *dynamodb.ScanOutput
		mockScanOutputAnalytics *dynamodb.ScanOutput
		mockScanErr             error
		expectedStatusCode      int
		expectedBody            string
	}{
		{
			name: "Valid user with analytics",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"id": "user123"},
			},
			mockScanOutputShortURLs: &dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{"id": {S: aws.String("short1")}},
					{"id": {S: aws.String("short2")}},
				},
			},
			mockScanOutputAnalytics: &dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{"idshort": {S: aws.String("short1")}},
					{"idshort": {S: aws.String("short2")}},
				},
			},
			mockScanErr:        nil,
			expectedStatusCode: http.StatusOK,
			expectedBody:       "user123",
		},
		{
			name: "User not found",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"id": "user404"},
			},
			mockScanOutputShortURLs: &dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{},
			},
			mockScanErr:        nil,
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "No short URLs found for the user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockDynamoDB)
			mockSvc.On("Scan", mock.Anything).Return(tt.mockScanOutputShortURLs, tt.mockScanErr).Once() // For short URLs
			mockSvc.On("Scan", mock.Anything).Return(tt.mockScanOutputAnalytics, tt.mockScanErr).Once() // For analytics

			// Replace svc with mockSvc in your Handler when testing
			response, err := Handler(tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)

			if response.StatusCode == http.StatusOK {
				var respBody map[string]interface{}
				err = json.Unmarshal([]byte(response.Body), &respBody)
				assert.NoError(t, err)
				assert.Contains(t, response.Body, tt.expectedBody)
			} else {
				assert.Contains(t, response.Body, tt.expectedBody)
			}
		})
	}
}
