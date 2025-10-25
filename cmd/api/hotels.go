package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
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

type Occupancy struct {
	Adults   int   `json:"adults"`
	Children []int `json:"children"`
}

type MinRateSearchRequest struct {
	HotelIds         []string    `json:"hotelIds"`
	Checkin          string      `json:"checkin"`
	Checkout         string      `json:"checkout"`
	Occupancies      []Occupancy `json:"occupancies"`
	Currency         string      `json:"currency"`
	GuestNationality string      `json:"guestNationality"`
	Timeout          int         `json:"timeout,omitempty"`
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

func (app *application) getHotelPriceHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	hotelID := params.ByName("hotel_id")

	if hotelID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "hotel_id parameter is required")
		return
	}

	apiKey := r.Header.Get("X-API-KEY")
	if apiKey == "" {
		apiKey = r.Header.Get("Authorization")
		if apiKey == "" {
			app.errorResponse(w, r, http.StatusUnauthorized, "API key is required")
			return
		}
	}

	hotelName := fmt.Sprintf("Hotel %s", hotelID)

	checkIn := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	checkOut := time.Now().AddDate(0, 0, 31).Format("2006-01-02")

	minPrice, err := app.getMinPriceFromAPI(hotelID, checkIn, checkOut, apiKey)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to get hotel rates")
		return
	}
	if minPrice <= 0 {
		app.errorResponse(w, r, http.StatusNotFound, "no price data found for this hotel")
		return
	}

	response := map[string]interface{}{
		"hotel_id":   hotelID,
		"hotel_name": hotelName,
		"price":      minPrice,
		"currency":   "USD",
		"check_in":   checkIn,
		"check_out":  checkOut,
		"adults":     1,
		"updated_at": time.Now().Format("2006-01-02T15:04:05Z"),
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"data": response}, nil)
	if err != nil {
		app.logError(r, err)
		app.errorResponse(w, r, http.StatusInternalServerError, "failed to encode response")
		return
	}
}

func (app *application) getMinPriceFromAPI(hotelID, checkIn, checkOut, apiKey string) (float64, error) {

	requestData := MinRateSearchRequest{
		HotelIds:         []string{hotelID},
		Checkin:          checkIn,
		Checkout:         checkOut,
		Occupancies:      []Occupancy{{Adults: 1, Children: []int{}}},
		Currency:         "USD",
		GuestNationality: "US",
		Timeout:          30,
	}

	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %v", err)
	}

	url := LITE_API_URL + "/hotels/min-rates"
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("X-API-Key", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("API returned status %d", res.StatusCode)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return 0, fmt.Errorf("failed to parse response: %v", err)
	}

	minPrice := app.extractMinPriceFromAPIResponse(response)
	return minPrice, nil
}

func (app *application) extractMinPriceFromAPIResponse(response map[string]interface{}) float64 {
	minPrice := 0.0

	if data, ok := response["data"].([]interface{}); ok {
		for _, item := range data {
			if itemData, ok := item.(map[string]interface{}); ok {
				if price, ok := itemData["price"].(float64); ok {
					if minPrice == 0 || price < minPrice {
						minPrice = price
					}
				}
			}
		}
	}

	return minPrice
}
