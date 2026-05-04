package clients

import (
	"context"
	"time"

	cfg "marketing-revenue-analytics/config"

	"marketing-revenue-analytics/utils"

	"github.com/avast/retry-go/v4"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	sqslib "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/goccy/go-json"
	"go.uber.org/zap"
)

type SqsClient struct {
	sqsclient  *sqslib.Client
	logger     *zap.Logger
	maxRetries int
	baseDelay  int
}

// convert types take an int and return a string value.
type consumerFunction func(string) error

func NewSqsClient(logger *zap.Logger) *SqsClient {
	// Load IAM role credentials and specify the AWS region.
	sqsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cfg.GetString("aws.region")))
	if err != nil {
		logger.Fatal("Error loading AWS SDK config: %v", zap.Error(err))
	}

	// Create an SQS client using the IAM role credentials and the specified AWS region.
	return &SqsClient{
		sqsclient:  sqslib.NewFromConfig(sqsCfg),
		logger:     logger,
		maxRetries: cfg.GetInt("retry.maxRetries"),
		baseDelay:  cfg.GetInt("retry.baseDelay"),
	}
}

func (sqs *SqsClient) SendMessage(ctx context.Context, data, queueURL string) (string, error) {
	var formatInput = &sqslib.SendMessageInput{
		MessageBody: aws.String(data),
		QueueUrl:    &queueURL,
	}
	var messageID string
	err := retry.Do(
		func() error {
			message, err := sqs.sqsclient.SendMessage(ctx, formatInput)
			if err != nil {
				return err
			}
			messageID = *message.MessageId
			return nil
		},
		retry.Attempts(uint(sqs.maxRetries)),                  // Customize retry attempts
		retry.DelayType(retry.FixedDelay),                     // You can choose between FixedDelay and BackOffDelay
		retry.Delay(time.Duration(sqs.baseDelay)*time.Second), // Customize delay between retries
	)

	if err != nil {
		return "", err
	}
	return messageID, nil
}

func (sqs *SqsClient) SendSqsMessageWithBody(ctx context.Context, sqsDataPayload interface{}, queueUrl string) error {
	logger := utils.GetCtxLogger(ctx)
	stringifyPayload, err := json.Marshal(sqsDataPayload)
	if err != nil {
		logger.Error("Error marshalling sqs Payload", zap.Error(err))
		return err
	}
	_, err = sqs.SendMessage(ctx, string(stringifyPayload), queueUrl)
	if err != nil {
		logger.Error("Error sending sqs event", zap.Error(err))
		return err
	}

	return nil
}

func (sqs *SqsClient) StartConsuming(ctx context.Context, queueURL string, fn consumerFunction) {
	sqs.logger.Info("SQS consumer started", zap.String("queue", queueURL))
	for {
		select {
		case <-ctx.Done():
			sqs.logger.Info("SQS consumer stopped")
			return
		default:
			sqs.processMessage(ctx, fn, queueURL)
		}
	}
}

func (sqs *SqsClient) processMessage(ctx context.Context, consumerFunction consumerFunction, queueURL string) {
	// Continuously poll for messages
	input := &sqslib.ReceiveMessageInput{
		QueueUrl:            &queueURL,
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     int32(cfg.GetInt("aws.waitTime")),
	}
	data, err := sqs.sqsclient.ReceiveMessage(ctx, input)
	if err != nil {
		return
	}
	for _, dataBody := range data.Messages {
		// Consume the message
		err := consumerFunction(*dataBody.Body)
		if err != nil {
			// If error, don't delete the sqs message and after visibility timeout, message will be retried
			continue
		}

		err = sqs.deleteMessage(ctx, dataBody.ReceiptHandle, &queueURL)
		if err != nil {
			sqs.logger.Error("Error deleting the message", zap.String("messageId", *(dataBody.MessageId)))
		}
	}
}

func (sqs *SqsClient) deleteMessage(ctx context.Context, messageHandle, queueURL *string) error {
	deleteMessageParams := &sqslib.DeleteMessageInput{
		QueueUrl:      queueURL,
		ReceiptHandle: messageHandle,
	}

	_, err := sqs.sqsclient.DeleteMessage(ctx, deleteMessageParams)
	if err != nil {
		return err
	}
	return nil
}
