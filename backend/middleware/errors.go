// Package middleware provides HTTP middleware and shared JSON error helpers.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// errorBody defines the standard JSON error structure for all API error responses.
//
//	{
//	  "error": {
//	    "code": "SOME_ERROR_CODE",
//	    "message": "Human readable error message"
//	  }
//	}
type errorBody struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondError writes a JSON error response with the given status, code, and message.
func RespondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, errorBody{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

// RespondInternalError writes a generic 500 error response.
func RespondInternalError(c *gin.Context) {
	RespondError(c,
		http.StatusInternalServerError,
		"INTERNAL_SERVER_ERROR",
		"An unexpected error occurred. Please try again later.",
	)
}

// RespondNotFound writes a structured 404 error response.
func RespondNotFound(c *gin.Context) {
	RespondError(c,
		http.StatusNotFound,
		"NOT_FOUND",
		"The requested resource was not found.",
	)
}
