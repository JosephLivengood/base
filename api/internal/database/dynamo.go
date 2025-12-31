package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/smithy-go/middleware"
)

type DynamoDB struct {
	Client *dynamodb.Client
}

type DynamoConfig struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
	Logger    *slog.Logger
	Debug     bool
}

func NewDynamo(cfg DynamoConfig) (*DynamoDB, error) {
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKey,
			cfg.SecretKey,
			"",
		)),
	)
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(awsCfg, func(o *dynamodb.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		if cfg.Debug && cfg.Logger != nil {
			o.APIOptions = append(o.APIOptions, func(stack *middleware.Stack) error {
				return stack.Initialize.Add(&loggingMiddleware{
					logger: cfg.Logger.With("db", "dynamodb"),
				}, middleware.After)
			})
		}
	})

	return &DynamoDB{Client: client}, nil
}

type loggingMiddleware struct {
	logger *slog.Logger
}

func (m *loggingMiddleware) ID() string {
	return "DynamoDBLogging"
}

func (m *loggingMiddleware) HandleInitialize(
	ctx context.Context,
	in middleware.InitializeInput,
	next middleware.InitializeHandler,
) (middleware.InitializeOutput, middleware.Metadata, error) {
	start := time.Now()
	out, md, err := next.HandleInitialize(ctx, in)
	duration := time.Since(start)

	op := middleware.GetOperationName(ctx)

	// Try to get table name from request
	var table string
	switch req := in.Parameters.(type) {
	case *dynamodb.PutItemInput:
		if req.TableName != nil {
			table = *req.TableName
		}
	case *dynamodb.GetItemInput:
		if req.TableName != nil {
			table = *req.TableName
		}
	case *dynamodb.QueryInput:
		if req.TableName != nil {
			table = *req.TableName
		}
	case *dynamodb.ScanInput:
		if req.TableName != nil {
			table = *req.TableName
		}
	case *dynamodb.DeleteItemInput:
		if req.TableName != nil {
			table = *req.TableName
		}
	case *dynamodb.UpdateItemInput:
		if req.TableName != nil {
			table = *req.TableName
		}
	}

	attrs := []slog.Attr{
		slog.String("op", op),
		slog.Duration("duration", duration),
	}
	if table != "" {
		attrs = append(attrs, slog.String("table", table))
	}
	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	m.logger.LogAttrs(ctx, slog.LevelDebug, "query", attrs...)
	return out, md, err
}

func (db *DynamoDB) Health(ctx context.Context) error {
	_, err := db.Client.ListTables(ctx, &dynamodb.ListTablesInput{
		Limit: aws.Int32(1),
	})
	return err
}
