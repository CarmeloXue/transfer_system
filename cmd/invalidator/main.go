package main

import (
	"context"
	"main/common/config"
	"main/common/db"
	"main/common/log"
	"main/common/recovery"
	"main/internal/account"
	"main/internal/transaction"
	"main/model"
	"time"

	"github.com/spf13/viper"
)

func main() {
	log.Init()
	defer log.Cleanup()
	config.Init()

	ticker := time.NewTicker(time.Minute * time.Duration(viper.GetInt(config.ConfigKeyInvalidateInterval)))
	defer ticker.Stop()
	txnDB, err := db.GetTransactionDB()
	if err != nil {
		panic("Could not initialize transaction database")
	}
	accDB, err := db.GetAccountDBClient()
	if err != nil {
		panic("Could not initialize account database")
	}
	transactionRepo := transaction.NewRepository(txnDB)
	accTCC := account.NewTCCService(accDB)
	ctx := context.Background()

	go func() {
		recovery.GoRecovery()

		for {
			select {
			case <-ticker.C:
				log.GetSugger().Info("start to invalidate expired transaction")
				// Run scan jog
				transactions, err := transactionRepo.QueryExpiredTransactions(ctx)
				if err != nil {
					log.GetSugger().Error("query expired transaction error", "err", err)
				}

				log.GetSugger().Info("get expored transactions", "transactions", transactions)

				for _, txn := range transactions {
					go func() {
						if err := accTCC.Cancel(ctx, txn.TransactionID); err != nil && err != account.ErrEmptyRollback {
							log.GetSugger().Error("failed to cancel", "txn", txn.TransactionID)

						}

						if err := transactionRepo.UpdateTransactionStatus(ctx, txn.TransactionID, model.Failed); err != nil {
							log.GetSugger().Error("failed to invalidate transaction", "txn", txn.TransactionID)
						}

						log.GetSugger().Info("auto invalicated expired transaction", "txn", txn.TransactionID)
					}()
				}
			}
		}
	}()

	log.GetSugger().Info("start invalidator")

	select {}
}
