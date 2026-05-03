package errors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	stdErrors "errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		var appErr *AppError
		if stdErrors.As(err, &appErr) {
			c.JSON(appErr.HTTPStatus, gin.H{
				"error":   appErr.Code,
				"message": appErr.Message,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "INTERNAL_ERROR",
			"message": "Something went wrong",
		})
	}
}
