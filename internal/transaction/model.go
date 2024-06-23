package transaction

type CreateTransactionRequest struct {
	SourceAccountID      int    `json:"source_account_id" binding:"required"`
	DestinationAccountID int    `json:"destination_account_id" binding:"required"`
	Amount               string `json:"amount" binding:"required"`
}

type ConfirmTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

type (
	QueryTransactionRequest struct {
		TransactionID string `uri:"transaction_id" json:"transaction_id" binding:"required"`
	}
)
