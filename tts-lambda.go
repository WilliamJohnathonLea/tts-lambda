package main

import (
	"context"
	"io"
	"os"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/polly"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/satori/go.uuid"
)

// InputEvent The request model for the lambda function
type InputEvent struct {
	Text string `json:"text"`
	Voice string `json:"voice"`
}

// Response The response model for the lambda function
type Response struct {
	Success bool `json:"success"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}

func synthesizeVoice(client *polly.Polly, text, voice *string) (io.ReadCloser, error) {
	input := polly.SynthesizeSpeechInput{
		OutputFormat: aws.String("mp3"),
		SampleRate: aws.String("16000"),
		Text: text,
		VoiceId: voice,
	}

	output, err := client.SynthesizeSpeech(&input)

	return output.AudioStream, err
}

func uploadToS3(uploder *s3manager.Uploader, bucket, key *string, file *io.ReadCloser) (*s3manager.UploadOutput, error) {
	return uploder.Upload(&s3manager.UploadInput{
		Bucket: bucket,
		Key: key,
		Body: *file,
	})
}

// HandleRequest The request handler
func HandleRequest(ctx context.Context, event InputEvent) (Response, error) {
	queueURL := aws.String(os.Getenv("SQS_URL"))
	bucketName := aws.String(os.Getenv("BUCKET_NAME"))
	awsSession := session.New()
	pollyClient := polly.New(awsSession)
	s3Uploader := s3manager.NewUploader(awsSession)
	sqsClinet := sqs.New(awsSession)
	id, err := uuid.NewV1()
	failureResponse := Response{}

	if err != nil {
		// Error when creating audio file name
		return failureResponse, err
	}

	output, err := synthesizeVoice(pollyClient, &event.Text, &event.Voice)

	if err != nil {
		// Failed to synthesize audio file
		return failureResponse, err
	}

	_, err = uploadToS3(s3Uploader, bucketName, aws.String(id.String() + ".mp3"), &output)

	if err != nil {
		// Failed to upload audio to S3
		return failureResponse, err
	}

	_, err = sqsClinet.SendMessage(&sqs.SendMessageInput{
		MessageBody: aws.String(id.String() + ".mp3"),
		QueueUrl: queueURL,
	})

	if err != nil {
		// Failed to push message to audio file name to SQS
		return failureResponse, err
	}

	return Response{Success: true, FileName: id.String(), FileType: "mp3"}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
