package ping

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"base/api/internal/database"
)

type Repository struct {
	postgres *database.PostgresDB
	dynamo   *database.DynamoDB
}

func NewRepository(postgres *database.PostgresDB, dynamo *database.DynamoDB) *Repository {
	return &Repository{
		postgres: postgres,
		dynamo:   dynamo,
	}
}

// PostgreSQL operations

func (r *Repository) UpsertIPPostgres(ctx context.Context, ip string) error {
	query := `
		INSERT INTO seen_ips (ip, num_visits, last_seen)
		VALUES ($1, 1, NOW())
		ON CONFLICT (ip) DO UPDATE SET
			num_visits = seen_ips.num_visits + 1,
			last_seen = NOW()
	`
	_, err := r.postgres.ExecContext(ctx, query, ip)
	return err
}

func (r *Repository) GetLastIPsPostgres(ctx context.Context, limit int) ([]SeenIP, error) {
	var ips []SeenIP
	query := `SELECT ip, num_visits, last_seen FROM seen_ips ORDER BY last_seen DESC LIMIT $1`
	err := r.postgres.SelectContext(ctx, &ips, query, limit)
	return ips, err
}

// DynamoDB operations

const pingPartitionKey = "PING" // Fixed partition key for GSI

func (r *Repository) InsertPingDynamo(ctx context.Context, ip string) error {
	now := time.Now().UTC()
	id := fmt.Sprintf("%d-%s", now.UnixNano(), ip)

	item := DynamoPing{
		ID:        id,
		PK:        pingPartitionKey,
		IP:        ip,
		Timestamp: now.Format(time.RFC3339Nano), // Nano for better sort precision
	}

	av, err := attributevalue.MarshalMap(item)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("pings"),
		Item:      av,
	}

	_, err = r.dynamo.Client.PutItem(ctx, input)
	return err
}

func (r *Repository) GetLastPingsDynamo(ctx context.Context, limit int) ([]Ping, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String("pings"),
		IndexName:              aws.String("pk-timestamp-index"),
		KeyConditionExpression: aws.String("pk = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: pingPartitionKey},
		},
		ScanIndexForward: aws.Bool(false), // Descending order
		Limit:            aws.Int32(int32(limit)),
	}

	result, err := r.dynamo.Client.Query(ctx, input)
	if err != nil {
		return nil, err
	}

	var pings []Ping
	for _, item := range result.Items {
		var dp DynamoPing
		if err := attributevalue.UnmarshalMap(item, &dp); err != nil {
			continue
		}
		pings = append(pings, dp.ToPing())
	}

	return pings, nil
}
