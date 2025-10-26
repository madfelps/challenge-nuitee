package models

import "database/sql"

type Models struct {
	Users         UserModel
	Favorites     FavoriteModel
	Notifications NotificationModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:         UserModel{DB: db},
		Favorites:     FavoriteModel{DB: db},
		Notifications: NotificationModel{DB: db},
	}
}
