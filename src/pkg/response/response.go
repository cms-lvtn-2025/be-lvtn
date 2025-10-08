package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, status int, err string) {
	c.JSON(status, Response{
		Success: false,
		Error:   err,
	})
}

func BadRequest(c *gin.Context, err string) {
	Error(c, http.StatusBadRequest, err)
}

func Unauthorized(c *gin.Context, err string) {
	Error(c, http.StatusUnauthorized, err)
}

func InternalError(c *gin.Context, err string) {
	Error(c, http.StatusInternalServerError, err)
}

func NotFound(c *gin.Context, err string) {
	Error(c, http.StatusNotFound, err)
}
