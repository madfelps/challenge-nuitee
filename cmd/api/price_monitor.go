package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/wneessen/go-mail"
)

func (app *application) StartPriceMonitor() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	app.logger.PrintInfo("price monitor started", nil)

	for range ticker.C {
		app.reconcilePriceMonitor()
	}
}

func (app *application) reconcilePriceMonitor() {
	favorites, err := app.models.Favorites.ListAllFavorites()
	if err != nil {
		app.logger.PrintError(err, nil)
		return
	}

	for _, favorite := range favorites {

		if favorite.WasNotified {
			continue
		}

		app.logger.PrintInfo("checking price for hotel", map[string]string{
			"hotel_id":     favorite.HotelID,
			"user_id":      strconv.Itoa(favorite.UserID),
			"target_price": fmt.Sprintf("%.2f", favorite.TargetPrice),
		})

		currentPrice, hotelName, err := app.getCurrentHotelPrice(favorite.HotelID)
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"hotel_id": favorite.HotelID,
			})
			continue
		}

		app.logger.PrintInfo("found price for hotel", map[string]string{
			"hotel_id":      favorite.HotelID,
			"hotel_name":    hotelName,
			"current_price": fmt.Sprintf("%.2f", currentPrice),
		})

		if currentPrice > favorite.TargetPrice {
			continue
		}

		app.logger.PrintInfo("alerting user", map[string]string{
			"user_id":       strconv.Itoa(favorite.UserID),
			"hotel_id":      favorite.HotelID,
			"hotel_name":    hotelName,
			"current_price": fmt.Sprintf("%.2f", currentPrice),
			"target_price":  fmt.Sprintf("%.2f", favorite.TargetPrice),
		})
		msg := app.createNotificationMessage("from@example.com", "to@example.com", fmt.Sprintf("$%.2f", favorite.TargetPrice), fmt.Sprintf("$%.2f", currentPrice))
		app.logger.PrintInfo("sending email for user", map[string]string{
			"user_id": strconv.Itoa(favorite.UserID),
		})
		app.sendEmail(msg)
		err = app.models.Favorites.MarkAsNotified(favorite.ID)
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"favorite_id": strconv.Itoa(favorite.ID),
			})
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
		app.logger.PrintError(err, map[string]string{
			"hotel_id": hotelID,
		})
		return 0, hotelName, fmt.Errorf("failed to get rates: %v", err)
	}
	if minPrice > 0 {
		return minPrice, hotelName, nil
	}

	app.logger.PrintInfo("no price data found for hotel", map[string]string{
		"hotel_id": hotelID,
	})
	return 0, hotelName, fmt.Errorf("no price data found")
}

func (app *application) createNotificationMessage(from, to, targetPrice, currentPrice string) *mail.Msg {
	msg := mail.NewMsg()
	if err := msg.From(from); err != nil {
		app.logger.PrintError(err, map[string]string{
			"from": from,
		})
	}
	if err := msg.To(to); err != nil {
		app.logger.PrintError(err, map[string]string{
			"to": to,
		})
	}
	msg.Subject("ALERT OF PRICE DROP!")
	body := fmt.Sprintf("Hey!\nThe price for one of your favorite hotels has dropped below your target price!\nYour target price is %s and the current price is %s.\nBook now!", targetPrice, currentPrice)
	msg.SetBodyString(mail.TypeTextPlain, body)
	return msg
}

func (app *application) sendEmail(msg *mail.Msg) {
	if err := app.email.DialAndSend(msg); err != nil {
		app.logger.PrintError(err, nil)
		return
	}
	app.logger.PrintInfo("notification email sent successfully", nil)
}
