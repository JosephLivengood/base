package subscription

import "time"

type Subscription struct {
	ID             string     `json:"id" db:"id"`
	OrganizationID string     `json:"organization_id" db:"organization_id"`
	Tier           string     `json:"tier" db:"tier"`
	ActiveFrom     time.Time  `json:"active_from" db:"active_from"`
	ActiveUntil    *time.Time `json:"active_until,omitempty" db:"active_until"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}
