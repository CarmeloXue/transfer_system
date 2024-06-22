package account

import "time"

type Account struct {
	ID        int       `json:"id"`
	AccountID int       `json:"account_id"`
	Balance   float64   `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (Account) TableName() string {
	return "account_tab"
}
