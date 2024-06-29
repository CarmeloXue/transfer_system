package account

import (
	"bytes"
	"encoding/json"
	"main/common/db/testutils"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func Test_CreateAccount_InvalidParameters(t *testing.T) {
	db, err := testutils.SetupTestDB()
	if err != nil {
		panic("setup db failed")
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	handler := NewHandler(NewAccountService(db))
	handler.CreateAccount(c)
	assert.Equal(t, 400, recorder.Code)
	validateResponseErrorMessage(t, "Invalid Request", recorder.Body)
}

func Test_QueryAccount_InvalidParameters(t *testing.T) {
	db, err := testutils.SetupTestDB()
	if err != nil {
		panic("setup db failed")
	}
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	handler := NewHandler(NewAccountService(db))
	handler.QueryAccount(c)
	assert.Equal(t, 400, recorder.Code)
	validateResponseErrorMessage(t, "Invalid Request", recorder.Body)
}

func validateResponseErrorMessage(t *testing.T, expectMsg string, body *bytes.Buffer) {
	type message struct {
		Message string `json:"message"`
	}
	msg := message{}
	json.NewDecoder(bytes.NewReader(body.Bytes())).Decode(&msg)
	assert.Equal(t, expectMsg, msg.Message)
}
