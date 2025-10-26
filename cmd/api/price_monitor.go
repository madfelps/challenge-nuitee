package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

func (app *application) StartPriceMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("price monitor started")

	for range ticker.C {
		app.checkPrices()
	}
}

func (app *application) checkPrices() {
	favorites, err := app.models.Favorites.ListAllFavorites()
	if err != nil {
		log.Printf("error getting favorites: %v", err)
		return
	}

	for _, favorite := range favorites {
		log.Printf("checking price for hotel %s (User: %d, Target: $%.2f)",
			favorite.HotelID, favorite.UserID, favorite.TargetPrice)

		currentPrice, hotelName, err := app.getCurrentHotelPrice(favorite.HotelID)
		if err != nil {
			log.Printf("error getting price for hotel %s: %v", favorite.HotelID, err)
			continue
		}

		log.Printf("found price for %s: $%.2f", hotelName, currentPrice)

		if currentPrice <= favorite.TargetPrice {
			fmt.Printf("ALERT: User %d - Hotel %s - Current price $%.2f is lower than target $%.2f\n",
				favorite.UserID, hotelName, currentPrice, favorite.TargetPrice)
		}
	}
}

func (app *application) getCurrentHotelPrice(hotelID string) (float64, string, error) {
	ctx := context.Background()

	hotelDetails, res, err := app.apiClient.StaticDataApi.GetHotelDetails(ctx).HotelId(hotelID).Execute()
	if err != nil {
		return 0, "", fmt.Errorf("failed to get hotel details: %v", err)
	}

	if res.StatusCode != 200 {
		return 0, "", fmt.Errorf("API returned status %d", res.StatusCode)
	}

	hotelName := "not identified hotel"
	if data, ok := hotelDetails["data"].(map[string]interface{}); ok {
		if name, ok := data["name"].(string); ok {
			hotelName = name
		}
	}

	checkIn := time.Now().AddDate(0, 0, 30).Format("2006-01-02")
	checkOut := time.Now().AddDate(0, 0, 31).Format("2006-01-02")

	minPrice, err := app.getMinPriceFromAPI(hotelID, checkIn, checkOut, app.config.apiKey)
	if err != nil {
		log.Printf("error getting min rates for hotel %s: %v", hotelID, err)
		return 0, hotelName, fmt.Errorf("failed to get rates: %v", err)
	}
	if minPrice > 0 {
		log.Printf("found real price for %s: $%.2f", hotelName, minPrice)
		return minPrice, hotelName, nil
	}

	log.Printf("no price data found for hotel %s", hotelID)
	return 0, hotelName, fmt.Errorf("no price data found")
}
