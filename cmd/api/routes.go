package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/hotels", app.listHotelsHandler)
	router.HandlerFunc(http.MethodGet, "/v1/hotels/:hotel_id", app.getHotelPriceHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.createUserHandler)
	router.HandlerFunc(http.MethodGet, "/v1/users", app.listUsersHandler)

	router.HandlerFunc(http.MethodPost, "/v1/favorites/:user_id", app.createFavoriteHandler)

	return app.recoverPanic(app.rateLimit(router))
}
