package transaction

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"gorm.io/gorm"

	trxModel "main/internal/model/transaction"
	"main/tools/log"
)

type TransactionMessageHandler struct {
	TrxRepo trxModel.Repository
	Topic   string
}

type TransactionCreationMessage struct {
	SourceAccountId      int
	DestinationAccountID int
	Amount               int64
	TransactionId        string
}

// ExecuteLocalTransaction is called before sending the message to broker
// In this case is create a
func (l *TransactionMessageHandler) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	var creationMessage TransactionCreationMessage
	if err := json.NewDecoder(bytes.NewReader(msg.Body)).Decode(&creationMessage); err != nil {
		return primitive.RollbackMessageState
	}
	trx := trxModel.Transaction{
		TransactionID:        creationMessage.TransactionId,
		SourceAccountID:      creationMessage.SourceAccountId,
		DestinationAccountID: creationMessage.DestinationAccountID,
		Amount:               creationMessage.Amount,
		TransactionStatus:    trxModel.Processing,
	}
	if err := l.TrxRepo.CreateTransaction(context.TODO(), trx); err != nil {
		return primitive.RollbackMessageState
	}
	return primitive.CommitMessageState
}

// CheckLocalTransaction is called to check the status of the transaction
func (l *TransactionMessageHandler) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	var creationMessage TransactionCreationMessage
	if err := json.NewDecoder(bytes.NewReader(msg.Body)).Decode(&creationMessage); err != nil {
		return primitive.RollbackMessageState
	}

	trx, err := l.TrxRepo.GetTransactionByID(context.TODO(), creationMessage.TransactionId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return primitive.RollbackMessageState
		}
		return primitive.UnknowState
	}
	switch trx.TransactionStatus {
	case trxModel.Failed, trxModel.Fulfiled:
		return primitive.RollbackMessageState
	case trxModel.Pending, trxModel.Processing:
		return primitive.CommitMessageState
	default:
		log.GetSugger().Error("unknow status", "status", trx.TransactionStatus)
		return primitive.UnknowState
	}
}
