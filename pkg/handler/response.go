package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type errorResponse struct {
	Message string `json:"message"`
}
type statusResponse struct {
	Status string `json:"status"`
}

func newErrorResponse(c *gin.Context, statusCode int, message string) {
	logrus.Error(message)
	c.AbortWithStatusJSON(statusCode, errorResponse{normalizeErrorMessage(statusCode, message)})
}

func bindingErrorMessage(err error) string {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		messages := make([]string, 0, len(validationErrs))
		for _, fieldErr := range validationErrs {
			messages = append(messages, validationFieldMessage(fieldErr))
		}
		return strings.Join(messages, "; ")
	}

	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return "invalid JSON body"
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := typeErr.Field
		if field == "" {
			return fmt.Sprintf("invalid value type for %s", typeErr.Type.String())
		}
		return fmt.Sprintf("field %s has invalid type", field)
	}

	return "invalid request body"
}

func validationFieldMessage(err validator.FieldError) string {
	field := fieldName(err.Field())

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("field %s is required", field)
	case "email":
		return fmt.Sprintf("field %s must be a valid email address", field)
	case "len":
		return fmt.Sprintf("field %s must be %s characters long", field, err.Param())
	case "numeric":
		return fmt.Sprintf("field %s must contain only digits", field)
	default:
		return fmt.Sprintf("field %s is invalid", field)
	}
}

func fieldName(name string) string {
	var result []rune
	for i, ch := range name {
		if i > 0 && ch >= 'A' && ch <= 'Z' {
			result = append(result, '_')
		}
		if ch >= 'A' && ch <= 'Z' {
			ch = ch - 'A' + 'a'
		}
		result = append(result, ch)
	}
	return string(result)
}

func normalizeErrorMessage(statusCode int, message string) string {
	switch message {
	case "No authorization header":
		return "Authorization header is required."
	case "Invalid authorization header":
		return "Authorization header must use the format: Bearer <token>."
	case "User Id not found", "User Id is of invalid type", "user id not found", "user id is of invalid type":
		return "Could not identify the current user."
	case "invalid company id":
		return "Company ID must be a valid number."
	case "invalid invitation id":
		return "Invitation ID must be a valid number."
	case "invalid user id":
		return "User ID must be a valid number."
	case "invalid event id":
		return "Event ID must be a valid number."
	case "invalid idea id":
		return "Idea ID must be a valid number."
	case "invalid availability id":
		return "Availability ID must be a valid number."
	case "invalid start_time":
		return "Field start_time must be a valid RFC3339 date-time."
	case "invalid end_time":
		return "Field end_time must be a valid RFC3339 date-time."
	case "company_id mismatch":
		return "Field company_id must match the company ID in the URL."
	case "invalid time range":
		return "End time must be later than start time."
	case "invalid status":
		return "Field status must be one of: unknown, going, not_going."
	case "no fields to update":
		return "Provide at least one field to update."
	case "company has no members":
		return "The company has no members yet."
	case "user is not a member of the company":
		return "You are not a member of this company."
	case "user not found":
		if statusCode == http.StatusNotFound {
			return "User not found."
		}
		return "No user with those details was found."
	case "event not found":
		return "Event not found."
	case "company not found":
		return "Company not found."
	case "auth challenge storage unavailable":
		return "Temporary error while processing the verification code. Please try again."
	case "invalid email or password":
		return "Incorrect email or password."
	case "pending registration not found":
		return "Verification code was not requested or has already expired."
	case "verification code expired":
		return "Verification code has expired. Request a new one."
	case "incorrect verification code":
		return "Verification code is incorrect."
	case "user with this email already exists":
		return "An account with this email already exists."
	case "user with this username already exists":
		return "This username is already taken."
	case "invitation already handled":
		return "This invitation has already been processed."
	case "cannot invite yourself":
		return "You cannot invite yourself."
	case "user already in company":
		return "This user is already a member of the company."
	case "invitation already sent":
		return "An invitation has already been sent to this user."
	case "only company owner can remove members":
		return "Only the company owner can remove members."
	case "cannot remove company owner":
		return "The company owner cannot be removed."
	case "title is required":
		return "Field title is required."
	case "name is required":
		return "Field name is required."
	case "username is required":
		return "Field username is required."
	case "start_time is required":
		return "Field start_time is required."
	case "title cannot be empty":
		return "Field title cannot be empty."
	case "description cannot be empty":
		return "Field description cannot be empty."
	case "password updated":
		return "Password updated."
	case "user deleted":
		return "User deleted."
	case "requested item was not found":
		return "Requested item was not found."
	case "avatar file is required":
		return "Field avatar is required."
	case "avatar file is too large":
		return "Avatar file must be 5 MB or smaller."
	case "avatar must be a png, jpeg, webp or gif image":
		return "Avatar must be a PNG, JPEG, WEBP or GIF image."
	case "failed to open avatar file", "failed to read avatar file":
		return "Could not process the uploaded avatar."
	}

	if strings.HasPrefix(message, "invalid password: ") {
		return capitalizeMessage(strings.TrimPrefix(message, "invalid password: ")) + "."
	}

	if strings.Contains(message, "token contains an invalid number of segments") ||
		strings.Contains(message, "token is malformed") ||
		strings.Contains(message, "token signature is invalid") ||
		strings.Contains(message, "token has invalid claims") ||
		strings.Contains(message, "token is expired") ||
		strings.Contains(message, "Invalid signing method") ||
		strings.Contains(message, "Invalid token claims") {
		return "Access token is invalid or has expired."
	}

	if strings.Contains(message, "JWT_SECRET not set") {
		return "Authentication service is temporarily unavailable."
	}

	if strings.Contains(message, "PASSWORD_SALT not set") {
		return "Password service is temporarily unavailable."
	}

	if strings.Contains(message, "SMTP credentials are not configured") ||
		strings.Contains(message, "SMTP connection failed") ||
		strings.Contains(message, "error sending email:") {
		return "Could not send the email right now. Please try again later."
	}

	if strings.Contains(message, "no rows in result set") {
		return "Requested item was not found."
	}

	return message
}

func capitalizeMessage(message string) string {
	if message == "" {
		return message
	}

	runes := []rune(message)
	first := runes[0]
	if first >= 'a' && first <= 'z' {
		runes[0] = first - 'a' + 'A'
	}

	result := string(runes)
	if strings.HasSuffix(result, ".") {
		return strings.TrimSuffix(result, ".")
	}
	return result
}
