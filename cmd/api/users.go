package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	models "github.com/madfelps/challenge-nuitee/internal/data"
	"github.com/madfelps/challenge-nuitee/internal/validator"
)

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	User User `json:"user"`
}

func (app *application) createUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}

	v := validator.New()
	validateUser(v, &req)
	ValidateEmail(v, req.Email)
	ValidatePasswordPlaintext(v, req.Password)

	if !v.Valid() {
		app.errorResponse(w, r, http.StatusUnprocessableEntity, v.Errors)
		return
	}

	passwordHash := hashPassword(req.Password)

	userID, err := app.models.Users.Insert(req.Name, req.Email, passwordHash)
	if err != nil {
		switch {
		case errors.Is(err, models.ErrDuplicateEmail):
			app.errorResponse(w, r, http.StatusConflict, "email already exists")
		default:
			app.logError(r, err)
			app.errorResponse(w, r, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	user := User{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	response := CreateUserResponse{
		User: user,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"data": response}, nil)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to encode response")
		return
	}
}

func (app *application) listUsersHandler(w http.ResponseWriter, r *http.Request) {
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	offset := 0
	limit := 20

	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid offset parameter")
			return
		}
	}

	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 || limit > 100 {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid limit parameter (must be between 1 and 100)")
			return
		}
	}

	users, total, err := app.models.Users.List(limit, offset)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "database error")
		return
	}

	response := map[string]interface{}{
		"users":  users,
		"total":  total,
		"offset": offset,
		"limit":  limit,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": response}, nil)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to encode response")
		return
	}
}

func hashPassword(password string) string {

	salt := make([]byte, 16)
	rand.Read(salt)

	passwordWithSalt := append([]byte(password), salt...)

	hash := sha256.Sum256(passwordWithSalt)

	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(hash[:])
}

func verifyPassword(password, storedHash string) bool {

	parts := strings.Split(storedHash, ":")
	if len(parts) != 2 {
		return false
	}

	salt, err := hex.DecodeString(parts[0])
	if err != nil {
		return false
	}

	passwordWithSalt := append([]byte(password), salt...)

	hash := sha256.Sum256(passwordWithSalt)
	hashStr := hex.EncodeToString(hash[:])

	return hashStr == parts[1]
}

func validateUser(v *validator.Validator, user *CreateUserRequest) {

	v.Check(user.Name != "", "name", "must be provided")
	v.Check(len(user.Name) <= 100, "name", "must not be more than 100 characters long")

}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}
