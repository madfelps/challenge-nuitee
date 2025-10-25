package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

type HotelFavorite struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	HotelID     string    `json:"hotel_id"`
	TargetPrice float64   `json:"target_price"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateFavoriteRequest struct {
	HotelID     string  `json:"hotel_id"`
	TargetPrice float64 `json:"target_price"`
}

type CreateFavoriteResponse struct {
	Favorite HotelFavorite `json:"favorite"`
}

func (app *application) createFavoriteHandler(w http.ResponseWriter, r *http.Request) {

	params := httprouter.ParamsFromContext(r.Context())
	userIDStr := params.ByName("user_id")

	if userIDStr == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "user_id parameter is required")
		return
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil || userID <= 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid user_id parameter")
		return
	}

	var req CreateFavoriteRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid JSON")
		return
	}

	if req.HotelID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "hotel_id is required")
		return
	}

	if req.TargetPrice <= 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "target_price is required and must be greater than 0")
		return
	}

	db, err := sql.Open("postgres", app.config.db.dsn)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "database connection failed")
		return
	}
	defer db.Close()

	var userExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)", userID).Scan(&userExists)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "database error")
		return
	}

	if !userExists {
		app.errorResponse(w, r, http.StatusNotFound, "user not found")
		return
	}

	var favoriteExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users_favorites WHERE user_id = $1 AND hotel_id = $2)", userID, req.HotelID).Scan(&favoriteExists)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "database error")
		return
	}

	if favoriteExists {
		app.errorResponse(w, r, http.StatusConflict, "hotel already in favorites")
		return
	}

	var favoriteID int
	var createdAt time.Time
	err = db.QueryRow(
		"INSERT INTO users_favorites (user_id, hotel_id, target_price) VALUES ($1, $2, $3) RETURNING id, created_at",
		userID, req.HotelID, req.TargetPrice,
	).Scan(&favoriteID, &createdAt)

	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to create favorite")
		return
	}

	favorite := HotelFavorite{
		ID:          favoriteID,
		UserID:      userID,
		HotelID:     req.HotelID,
		TargetPrice: req.TargetPrice,
		CreatedAt:   createdAt,
	}

	response := CreateFavoriteResponse{
		Favorite: favorite,
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{"data": response}, nil)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to encode response")
		return
	}
}
