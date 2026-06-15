package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AppError struct {
	HTTPStatus int
	Code       string
	Message    string
}

var (
	ErrBadRequest       = AppError{http.StatusBadRequest, "BAD_REQUEST", "Invalid input data provided"}
	ErrUnauthorized     = AppError{http.StatusUnauthorized, "UNAUTHORIZED", "Access denied. Please authenticate"}
	ErrForbidden        = AppError{http.StatusForbidden, "FORBIDDEN", "You don't have permission to access this resource"}
	ErrNotFound         = AppError{http.StatusNotFound, "NOT_FOUND", "Requested resource not found"}
	ErrConflict         = AppError{http.StatusConflict, "CONFLICT", "Resource already exists or data conflict"}
	ErrInternal         = AppError{http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error occurred"}
	ErrAlreadyVoted     = AppError{http.StatusConflict, "ALREADY_VOTED", "You have already cast your vote"}
	ErrVotingClosed     = AppError{http.StatusForbidden, "VOTING_CLOSED", "Voting period has ended"}
	ErrInvalidCandidate = AppError{http.StatusBadRequest, "INVALID_CANDIDATE", "Candidate does not exist"}
)

func NewError(c *gin.Context, appErr AppError, details any) {
	reqID, _ := c.Get("request_id")
	reqIDStr, _ := reqID.(string)
	if reqIDStr == "" {
		reqIDStr = "req_" + uuid.New().String()
	}

	c.JSON(appErr.HTTPStatus, WebResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: details,
		},
		RequestID: reqIDStr,
	})
}
