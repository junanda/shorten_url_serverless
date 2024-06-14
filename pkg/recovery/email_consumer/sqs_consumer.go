package main

import (
	"encoding/json"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/junanda/shortenerUrl/utils"
)

func Handler() {
	sess := session.Must(session.NewSession())
	sqsSvc := sqs.New(sess)
	queueUrl := os.Getenv("SQS_QUEUE_URL")

	// Receive message from SQS
	receiveMsgInput := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(queueUrl),
		MaxNumberOfMessages: aws.Int64(1),
		WaitTimeSeconds:     aws.Int64(20),
	}

	msgOutput, err := sqsSvc.ReceiveMessage(receiveMsgInput)
	if err != nil {
		utils.PrintError("Failed to receive message from SQS", err)
		return
	}

	if len(msgOutput.Messages) == 0 {
		utils.PrintError("No messages received", nil)
		return
	}

	// Process each message
	for _, message := range msgOutput.Messages {
		var data map[string]string
		err := json.Unmarshal([]byte(*message.Body), &data)
		if err != nil {
			utils.PrintError("Failed to unmarshal message", err)
			continue
		}

		email := data["email"]
		password := data["password"]

		// Send email with the new password
		err = utils.SendEmail(email, "Password Recovery", "Your new password is: "+password)
		if err != nil {
			utils.PrintError("Failed to send email", err)
			continue
		}

		// Delete the message from the queue after processing
		deleteMsgInput := &sqs.DeleteMessageInput{
			QueueUrl:      aws.String(queueUrl),
			ReceiptHandle: message.ReceiptHandle,
		}
		_, err = sqsSvc.DeleteMessage(deleteMsgInput)
		if err != nil {
			utils.PrintError("Failed to delete message from SQS", err)
			continue
		}
	}
}

func main() {
	lambda.Start(Handler)
}
