package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
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

	if req.Name == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "name is required")
		return
	}

	if req.Email == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "email is required")
		return
	}

	if req.Password == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "password is required")
		return
	}

	if len(req.Password) < 6 {
		app.errorResponse(w, r, http.StatusBadRequest, "password must be at least 6 characters")
		return
	}

	if !isValidEmail(req.Email) {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid email format")
		return
	}

	db, err := sql.Open("postgres", app.config.db.dsn)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "database connection failed")
		return
	}
	defer db.Close()

	var existingID int
	err = db.QueryRow("SELECT id FROM users WHERE email = $1", req.Email).Scan(&existingID)
	if err != sql.ErrNoRows {
		if err != nil {
			app.logError(r, err)
			app.errorResponse(w, r, http.StatusInternalServerError, "database error")
			return
		}
		app.errorResponse(w, r, http.StatusConflict, "email already exists")
		return
	}

	passwordHash := hashPassword(req.Password)

	var userID int
	var createdAt time.Time
	err = db.QueryRow(
		"INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at",
		req.Name, req.Email, passwordHash,
	).Scan(&userID, &createdAt)

	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create user")
		return
	}

	user := User{
		ID:        userID,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: createdAt,
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

	db, err := sql.Open("postgres", app.config.db.dsn)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "database connection failed")
		return
	}
	defer db.Close()

	rows, err := db.Query(
		"SELECT id, name, email, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "database error")
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt)
		if err != nil {
			app.logError(r, err)
			app.errorResponse(w, r, http.StatusInternalServerError, "database error")
			return
		}
		users = append(users, user)
	}

	var total int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
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

func isValidEmail(email string) bool {

	if len(email) < 5 {
		return false
	}

	hasAt := false
	hasDot := false

	for i, char := range email {
		if char == '@' {
			if hasAt || i == 0 || i == len(email)-1 {
				return false
			}
			hasAt = true
		}
		if char == '.' && hasAt {
			hasDot = true
		}
	}

	return hasAt && hasDot
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
