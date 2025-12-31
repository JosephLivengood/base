package ping

import "time"

// SeenIP represents an IP record in PostgreSQL
type SeenIP struct {
	IP        string    `json:"ip" db:"ip"`
	NumVisits int       `json:"num_visits" db:"num_visits"`
	LastSeen  time.Time `json:"last_seen" db:"last_seen"`
}

// Ping represents a single ping record in DynamoDB
type Ping struct {
	ID        string    `json:"id"`
	IP        string    `json:"ip"`
	Timestamp time.Time `json:"timestamp"`
}

// DynamoPing is used for DynamoDB marshaling/unmarshaling
type DynamoPing struct {
	ID        string `dynamodbav:"id"`
	PK        string `dynamodbav:"pk"`
	IP        string `dynamodbav:"ip"`
	Timestamp string `dynamodbav:"timestamp"`
}

func (d DynamoPing) ToPing() Ping {
	ts, _ := time.Parse(time.RFC3339Nano, d.Timestamp)
	return Ping{
		ID:        d.ID,
		IP:        d.IP,
		Timestamp: ts,
	}
}

type PingResponse struct {
	YourIP      string   `json:"your_ip"`
	PostgresIPs []SeenIP `json:"postgres_ips"`
	DynamoPings []Ping   `json:"dynamo_pings"`
}
