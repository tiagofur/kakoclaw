package storage

import "time"

type Contact struct {
	ID             int64      `json:"id"`
	Email          string     `json:"email"`
	FirstName      string     `json:"first_name"`
	LastName       string     `json:"last_name"`
	Phone          string     `json:"phone"`
	Company        string     `json:"company"`
	Title          string     `json:"title"`
	Tags           string     `json:"tags"`
	CustomFields   string     `json:"custom_fields"`
	Status         string     `json:"status"`
	Source         string     `json:"source"`
	SubscribedAt   *time.Time `json:"subscribed_at,omitempty"`
	UnsubscribedAt *time.Time `json:"unsubscribed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ContactList struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Account     string    `json:"account"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ContactListMember struct {
	ContactID int64     `json:"contact_id"`
	ListID    int64     `json:"list_id"`
	AddedAt   time.Time `json:"added_at"`
}

type Segment struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	Rules        string    `json:"rules"`
	Account      string    `json:"account"`
	ContactCount int       `json:"contact_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type SegmentRule struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

type ContactFilter struct {
	Q      string
	Status string
	Tags   string
	Page   int
	Limit  int
}

type EmailDelivery struct {
	ID                int64      `json:"id"`
	CampaignAccount   string     `json:"campaign_account"`
	CampaignName      string     `json:"campaign_name"`
	ContactID         *int64     `json:"contact_id,omitempty"`
	Subject           string     `json:"subject"`
	Status            string     `json:"status"`
	ProviderMessageID string     `json:"provider_message_id"`
	SentAt            *time.Time `json:"sent_at,omitempty"`
	DeliveredAt       *time.Time `json:"delivered_at,omitempty"`
	OpenedAt          *time.Time `json:"opened_at,omitempty"`
	ClickedAt         *time.Time `json:"clicked_at,omitempty"`
	BouncedAt         *time.Time `json:"bounced_at,omitempty"`
	Error             string     `json:"error"`
	CreatedAt         time.Time  `json:"created_at"`
}
