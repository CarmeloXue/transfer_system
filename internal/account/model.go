package account

type (
	CreateAccountRequest struct {
		AccountID      uint64 `json:"account_id" binding:"required"`
		InitialBalance string `json:"initial_balance"`
	}

	// CreateAccountResponse represents the JSON response body structure
	CreateAccountResponse struct {
		AccountID uint64 `json:"account_id"`
		Balance   string `json:"balance"`
	}

	QueryAccountRequest struct {
		AccountID uint64 `uri:"account_id" json:"account_id" binding:"required"`
	}

	// CreateAccountResponse represents the JSON response body structure
	QueryResponse struct {
		AccountID uint64 `json:"account_id"`
		Balance   string `json:"balance"`
	}
)
