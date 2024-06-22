package account

import "time"

type Account struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	AccountID int       `gorm:"unique;not null" json:"account_id"`
	Balance   float64   `gorm:"type:decimal(20,8);not null;default:0" json:"balance"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName sets the insert table name for this struct type.
func (Account) TableName() string {
	return "account_tab"
}

func (r *repository) CreateAccount(account *Account) error {
	if err := r.db.Create(account).Error; err != nil {
		return err
	}
	return nil
}

func (r *repository) GetAccountByID(accountID int) (Account, error) {
	var acc Account
	if err := r.db.First(&acc, Account{
		AccountID: accountID,
	}).Error; err != nil {
		return Account{}, err
	}
	return acc, nil
}

func (r *repository) countAccount() (int64, error) {
	var count int64
	if err := r.db.Model(Account{}).Count(&count).Error; err != nil {
		return count, err
	}
	return count, nil
}
