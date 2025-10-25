package main

import (
	"context"
	"net/http"
	"strconv"

	liteapi "github.com/liteapi-travel/go-sdk/v3"
)

type Hotel struct {
	HotelID     string  `json:"hotel_id"`
	Name        string  `json:"name"`
	Address     string  `json:"address"`
	City        string  `json:"city"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Stars       float64 `json:"stars"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
}

type HotelsResponse struct {
	Hotels []Hotel `json:"hotels"`
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
}

func (app *application) listHotelsHandler(w http.ResponseWriter, r *http.Request) {

	countryCode := r.URL.Query().Get("countryCode")
	cityName := r.URL.Query().Get("cityName")

	if countryCode == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "countryCode parameter is required")
		return
	}

	if cityName == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "cityName parameter is required")
		return
	}

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

	configuration := liteapi.NewConfiguration()
	apiKey := r.Header.Get("X-API-KEY")
	if apiKey == "" {

		apiKey = r.Header.Get("Authorization")
		if apiKey == "" {
			app.errorResponse(w, r, http.StatusUnauthorized, "API key is required")
			return
		}
	}

	configuration.AddDefaultHeader("X-API-KEY", apiKey)
	apiClient := liteapi.NewAPIClient(configuration)

	ctx := context.Background()

	result, res, err := apiClient.StaticDataApi.GetHotels(ctx).
		CountryCode(countryCode).
		CityName(cityName).
		Offset(int32(offset)).
		Limit(int32(limit)).
		Execute()

	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to fetch hotels from LiteAPI")
		return
	}

	if res.StatusCode != http.StatusOK {
		app.errorResponse(w, r, res.StatusCode, "LiteAPI returned an error")
		return
	}

	hotels := make([]Hotel, 0)

	if data, ok := result["data"].([]interface{}); ok {
		for _, item := range data {
			if hotelData, ok := item.(map[string]interface{}); ok {
				h := Hotel{}

				if id, ok := hotelData["id"].(string); ok {
					h.HotelID = id
				}
				if name, ok := hotelData["name"].(string); ok {
					h.Name = name
				}
				if address, ok := hotelData["address"].(string); ok {
					h.Address = address
				}
				if city, ok := hotelData["city"].(string); ok {
					h.City = city
				}
				if country, ok := hotelData["country"].(string); ok {
					h.Country = country
				}

				h.CountryCode = countryCode
				if stars, ok := hotelData["stars"].(float64); ok {
					h.Stars = stars
				}
				if lat, ok := hotelData["latitude"].(float64); ok {
					h.Latitude = lat
				}
				if lng, ok := hotelData["longitude"].(float64); ok {
					h.Longitude = lng
				}

				hotels = append(hotels, h)
			}
		}
	}

	response := HotelsResponse{
		Hotels: hotels,
		Total:  len(hotels),
		Offset: offset,
		Limit:  limit,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": response}, nil)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to encode response")
		return
	}
}
