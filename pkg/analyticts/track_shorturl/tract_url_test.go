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

func (m *MockDynamoDB) GetItem(input *dynamodb.GetItemInput) (*dynamodb.GetItemOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDB) Scan(input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*dynamodb.ScanOutput), args.Error(1)
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name               string
		request            events.APIGatewayProxyRequest
		mockGetItemOutput  *dynamodb.GetItemOutput
		mockGetItemErr     error
		mockScanOutput     *dynamodb.ScanOutput
		mockScanErr        error
		expectedStatusCode int
		expectedBody       string
	}{
		{
			name: "Valid short URL with counts",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"url": "abc123"},
			},
			mockGetItemOutput: &dynamodb.GetItemOutput{
				Item: map[string]*dynamodb.AttributeValue{
					"short_url": {S: aws.String("abc123")},
					"idshort":   {S: aws.String("123")},
				},
			},
			mockGetItemErr: nil,
			mockScanOutput: &dynamodb.ScanOutput{
				Items: []map[string]*dynamodb.AttributeValue{
					{"idshort": {S: aws.String("123")}},
					{"idshort": {S: aws.String("123")}},
				},
			},
			mockScanErr:        nil,
			expectedStatusCode: http.StatusOK,
			expectedBody:       `{"short_url":"abc123","access_count":2}`,
		},
		{
			name: "Short URL not found",
			request: events.APIGatewayProxyRequest{
				PathParameters: map[string]string{"url": "nonexistent"},
			},
			mockGetItemOutput:  &dynamodb.GetItemOutput{},
			mockGetItemErr:     nil,
			expectedStatusCode: http.StatusNotFound,
			expectedBody:       "Short URL not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockDynamoDB)
			mockSvc.On("GetItem", mock.Anything).Return(tt.mockGetItemOutput, tt.mockGetItemErr)
			mockSvc.On("Scan", mock.Anything).Return(tt.mockScanOutput, tt.mockScanErr)

			// Replace svc with mockSvc in your Handler when testing
			response, err := Handler(tt.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatusCode, response.StatusCode)

			if response.StatusCode == http.StatusOK {
				var respBody map[string]interface{}
				err = json.Unmarshal([]byte(response.Body), &respBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response.Body)
			} else {
				assert.Contains(t, response.Body, tt.expectedBody)
			}
		})
	}
}
