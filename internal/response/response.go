package response

import "github.com/gin-gonic/gin"

type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
type Envelope struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

func OK(c *gin.Context, data any)      { c.JSON(200, Envelope{Success: true, Data: data}) }
func Created(c *gin.Context, data any) { c.JSON(201, Envelope{Success: true, Data: data}) }
func Fail(c *gin.Context, status int, code, msg string) {
	c.JSON(status, Envelope{Success: false, Error: &ErrorBody{Code: code, Message: msg}})
	c.Abort()
}
