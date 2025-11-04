package models

import (
	"context"
	"database/sql"
	"time"
)

type Favorite struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	HotelID     string    `json:"hotel_id"`
	TargetPrice float64   `json:"target_price"`
	CreatedAt   time.Time `json:"created_at"`
}

type FavoriteModel struct {
	DB *sql.DB
}

func (m FavoriteModel) ListAllFavorites() ([]Favorite, error) {
	query := `
		SELECT id, user_id, hotel_id, target_price, created_at
		FROM users_favorites
		ORDER BY created_at DESC
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var favorites []Favorite
	for rows.Next() {
		var f Favorite
		err := rows.Scan(&f.ID, &f.UserID, &f.HotelID, &f.TargetPrice, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		favorites = append(favorites, f)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return favorites, nil
}
