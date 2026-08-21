package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func OKMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": message,
	})
}

func Fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
	})
}

func ValidationFail(c *gin.Context, errors []FieldError) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{
		"code":    422,
		"message": "validation failed",
		"errors":  errors,
	})
}

func Unauthorized(c *gin.Context, message string) {
	Fail(c, http.StatusUnauthorized, 401, message)
}

func Forbidden(c *gin.Context, message string) {
	Fail(c, http.StatusForbidden, 403, message)
}

func NotFound(c *gin.Context, message string) {
	Fail(c, http.StatusNotFound, 404, message)
}
