package response

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ok(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
		"data":    data,
	})
}

func Err(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": err.Error(),
		"data":    "",
	})
}

func ErrorParam(c *gin.Context, errMsg string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": fmt.Sprintf("param error: %v", errMsg),
		"data":    "",
	})
}

func ErrorServer(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": "server error",
		"data":    "",
	})
}

func ErrorDuplicated(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, gin.H{
		"message": message,
		"data":    "",
	})
}

func ErrorServerWithErrorMessage(c *gin.Context, err error) {
	c.JSON(http.StatusInternalServerError, gin.H{
		"message": fmt.Sprintf("server error: %v", err.Error()),
		"data":    "",
	})
}
func ErrorNotFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"message": "resource not found",
		"data":    "",
	})
}

func WithExternalResponse(c *gin.Context, resp ExternalResponse) {
	c.JSON(resp.Code, gin.H{
		"message": resp.Message,
	})
}

func MapExternalErrors(c *gin.Context, err error, errMap map[error]*ExternalResponse) {
	var (
		response *ExternalResponse
		errExist bool
	)
	defer func() {
		if response == nil {
			response = &ExternalResponse{
				Code:    500,
				Message: "Internal Server error",
			}
		}
		c.JSON(response.Code, gin.H{
			"message": response.Message,
		})
	}()

	if errMap == nil {
		return
	}

	response, errExist = errMap[err]
	if !errExist {
		return
	}
}

type ExternalResponse struct {
	Code    int
	Message string
}
