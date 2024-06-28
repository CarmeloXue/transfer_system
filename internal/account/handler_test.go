package account

import (
	"main/common/db/testutils"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestInvalidParameters(t *testing.T) {

	db, err := testutils.SetupTestDB()
	if err != nil {
		panic("setup db failed")
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	handler := NewHandler(NewAccountService(db))
	handler.CreateAccount(c)
	assert.Equal(t, 400, recorder.Code)
}
