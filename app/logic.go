package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	gomail "gopkg.in/mail.v2"
	"gorm.io/gorm"
)

func GetWeatherData(cityName string) (*ResponseOutput, error) {
	var city City
	cityResult := DB.Where("name = ?", cityName).First(&city)
	if cityResult.Error != nil {
		if cityResult.Error == gorm.ErrRecordNotFound {
			return nil, ErrCityNotFound
		}
		return nil, fmt.Errorf("%w: error finding city '%s': %v", ErrDatabaseOperation, cityName, cityResult.Error)
	}

	var weather Weather
	now := time.Now()

	weatherResult := DB.Preload("City").Preload("Description").Where("city_id = ?", city.ID).Order(gorm.Expr("ABS(EXTRACT(EPOCH FROM (observation_time - ?)))", now)).First(&weather)

	if weatherResult.Error != nil {
		if weatherResult.Error == gorm.ErrRecordNotFound {
			return nil, ErrCityNotFound
		}
		return nil, fmt.Errorf("%w: error getting weather data for city ID %d: %v", ErrDatabaseOperation, city.ID, weatherResult.Error)
	}

	return &ResponseOutput{
		Temperature: weather.Temperature,
		Humidity:    weather.Humidity,
		Description: weather.Description.Description,
	}, nil
}

func SubscribeUser(email, frequency, cityName string) error {
	if frequency != "hourly" && frequency != "daily" {
		return ErrInvalidFrequency
	}

	var city City
	cityCheckResult := DB.Where("name = ?", cityName).First(&city)
	if cityCheckResult.Error != nil {
		if cityCheckResult.Error == gorm.ErrRecordNotFound {
			return ErrCityNotFound
		}
		return fmt.Errorf("%w: error checking city existence '%s': %v", ErrDatabaseOperation, cityName, cityCheckResult.Error)
	}

	var existingUser User
	checkResult := DB.Where("email = ?", email).First(&existingUser)
	if checkResult.Error == nil {
		return ErrEmailAlreadyExists
	}
	if checkResult.Error != gorm.ErrRecordNotFound {
		return fmt.Errorf("%w: error checking existing email: %v", ErrDatabaseOperation, checkResult.Error)
	}

	token, err := GenerateToken()
	if err != nil {
		return fmt.Errorf("error generating token: %w", err)
	}

	newUser := User{
		ID:          uuid.New(),
		Email:       email,
		IsActivated: false,
		Token:       token,
		Frequency:   frequency,
		City:        cityName,
	}
	createResult := DB.Create(&newUser)
	if createResult.Error != nil {
		return fmt.Errorf("%w: error inserting user into database: %v", ErrDatabaseOperation, createResult.Error)
	}

	confirmAdres := settings.LocalURL + "/confirm/" + token
	unSubscribeAdres := settings.LocalURL + "/unsubscribe/" + token

	htmlTemplate := settings.EmailSubscribeText

	emailHTMLBody := strings.ReplaceAll(htmlTemplate, "{1}", confirmAdres)
	emailHTMLBody = strings.ReplaceAll(emailHTMLBody, "{2}", unSubscribeAdres)

	message := gomail.NewMessage()
	message.SetHeader("From", settings.SMTP.Username)
	message.SetHeader("To", email)
	message.SetHeader("Subject", "Subscription to weather report - Confirmation")
	message.SetBody("text/html", emailHTMLBody)
	dialer := gomail.NewDialer(settings.SMTP.Adres, settings.SMTP.Port, settings.SMTP.Username, settings.SMTP.Password)

	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("Error sending confirmation email to %s: %v", email, err)
		return fmt.Errorf("error sending confirmation email: %w", err)
	}
	log.Printf("Successfully sent confirmation email to %s", email)
	return nil
}

func ConfirmUser(token string) error {
	var user User
	result := DB.Where("token = ? AND is_activated = ?", token, false).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return ErrInvalidOrExpiredToken
		}
		return fmt.Errorf("%w: error finding user for confirmation: %v", ErrDatabaseOperation, result.Error)
	}

	updateResult := DB.Model(&user).Update("is_activated", true)
	if updateResult.Error != nil {
		return fmt.Errorf("%w: error updating user activation status: %v", ErrDatabaseOperation, updateResult.Error)
	}
	log.Printf("User with token %s confirmed successfully", token)
	return nil
}

func UnsubscribeUser(token string) error {
	var user User
	result := DB.Where("token = ?", token).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return ErrInvalidOrExpiredToken
		}
		return fmt.Errorf("%w: error finding user for unsubscribe: %v", ErrDatabaseOperation, result.Error)
	}

	deleteResult := DB.Delete(&user)
	if deleteResult.Error != nil {
		return fmt.Errorf("%w: error deleting user during unsubscribe: %v", ErrDatabaseOperation, deleteResult.Error)
	}

	if deleteResult.RowsAffected == 0 {
		return fmt.Errorf("%w: no rows affected during unsubscribe deletion", ErrDatabaseOperation)
	}
	log.Printf("User with token %s unsubscribed successfully", token)
	return nil
}

func sendWeatherUpdateEmail(recipientEmail, city string, weather *ResponseOutput) error {
	if settings.SMTP.Adres == "" || settings.SMTP.Username == "" || settings.SMTP.Password == "" {
		log.Println("Warning: SMTP settings are incomplete. Cannot send email.")
		return fmt.Errorf("SMTP settings are not configured")
	}

	htmlTemplate := settings.EmailUpdateText

	// Приведений нижче вміст email повідомлень виконує функцію fallback на випадок якщо потрібний контент не буде знайдено в базі даних

	emailHTMLBody := ""
	if weather == nil {
		emailHTMLBody = fmt.Sprintf(`
            <!DOCTYPE html>
            <html>
            <head><title>Weather Update</title></head>
            <body>
                <h2>Weather Update for %s</h2>
                <p>Could not retrieve weather information at this time.</p>
                <div class="footer">
                    <p>You can unsubscribe at any time by visiting: [Unsubscribe Link Missing]</p>
                </div>
            </body>
            </html>
            `,
			city,
		)
	} else {
		if htmlTemplate == "" {
			log.Println("Warning: EmailUpdateText template is empty. Using fallback format.")
			htmlTemplate = `
                <!DOCTYPE html>
                <html>
                <head><title>Weather Update for %s</title></head>
                <body>
                    <h2>Weather Update for %s</h2>
                    <div class="weather-details">
                        <p><strong>Temperature:</strong> %d°C</p>
                        <p><strong>Humidity:</strong> %d%%</p>
                        <p><strong>Description:</strong> %s</p>
                    </div>
                    <div class="footer">
                        <p>You can unsubscribe at any time by visiting: [Unsubscribe Link Missing]</p>
                    </div>
                </body>
                </html>
                `
		}

		emailHTMLBody = fmt.Sprintf(htmlTemplate,
			city,
			weather.Temperature,
			weather.Humidity,
			weather.Description,
		)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", settings.SMTP.Username)
	message.SetHeader("To", recipientEmail)
	message.SetHeader("Subject", fmt.Sprintf("Weather Update for %s", city))
	message.SetBody("text/html", emailHTMLBody)

	dialer := gomail.NewDialer(settings.SMTP.Adres, settings.SMTP.Port, settings.SMTP.Username, settings.SMTP.Password)

	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("Email send failed to %s for city %s: %v", recipientEmail, city, err)
		return fmt.Errorf("failed to send email to %s: %w", recipientEmail, err)
	}

	log.Printf("Email sent successfully to %s for city %s", recipientEmail, city)
	return nil
}

func SendWeatherUpdates(frequency string) {
	log.Printf("Starting weather update process for frequency: %s", frequency)

	users, err := GetSubscribedUsers(frequency)
	if err != nil {
		log.Printf("Error fetching subscribed users for frequency '%s': %v", frequency, err)
		return
	}

	if len(users) == 0 {
		log.Printf("No activated users subscribed to %s updates found.", frequency)
		return
	}

	log.Printf("Found %d activated users subscribed to %s updates.", len(users), frequency)

	for _, user := range users {
		emailRetryMutex.Lock()
		_, inRetryQueue := emailRetryQueue[user.Email]
		emailRetryMutex.Unlock()

		if inRetryQueue {
			log.Printf("Email %s for city %s is already in the retry queue, skipping initial send.", user.Email, user.City)
			continue
		}

		weather, err := GetWeatherData(user.City)
		weatherInfo := ""
		if err != nil {
			log.Printf("Could not get weather data for city '%s' (User: %s) during update run: %v", user.City, user.Email, err)
			weatherInfo = formatWeatherInfo(nil, user.City)
		} else {
			weatherInfo = formatWeatherInfo(weather, user.City)
		}

		sendErr := sendWeatherUpdateEmail(user.Email, user.City, weather)

		if sendErr != nil {
			log.Printf("Adding email %s for city %s to retry queue.", user.Email, user.City)
			emailRetryMutex.Lock()
			emailRetryQueue[user.Email] = EmailRetry{
				RecipientEmail: user.Email,
				City:           user.City,
				WeatherInfo:    weatherInfo,
				NextAttempt:    time.Now().Add(RETRY_DELAY),
				Attempts:       1,
			}
			emailRetryMutex.Unlock()
		} else {
			emailRetryMutex.Lock()
			if _, wasInQueue := emailRetryQueue[user.Email]; wasInQueue {
				log.Printf("Email %s successfully sent, removing from retry queue.", user.Email)
				delete(emailRetryQueue, user.Email)
			}
			emailRetryMutex.Unlock()
		}
	}

	log.Printf("Finished main weather update processing for frequency: %s. Retry queue has %d emails.", frequency, len(emailRetryQueue))
}

func startEmailRetryScheduler() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	log.Println("Starting email retry scheduler.")

	for range ticker.C {
		log.Println("Checking email retry queue...")
		emailRetryMutex.Lock()

		emailsToRetryNow := []string{}
		for email, retry := range emailRetryQueue {
			if time.Now().After(retry.NextAttempt) {
				emailsToRetryNow = append(emailsToRetryNow, email)
			}
		}

		emailRetryMutex.Unlock()

		if len(emailsToRetryNow) == 0 {
			log.Println("No emails ready for retry.")
			continue
		}

		log.Printf("Attempting to retry %d email(s).", len(emailsToRetryNow))

		for _, email := range emailsToRetryNow {
			emailRetryMutex.Lock()
			retryEntry, exists := emailRetryQueue[email]
			emailRetryMutex.Unlock()

			if !exists {
				continue
			}

			if retryEntry.Attempts >= MAX_EMAIL_RETRIES {
				log.Printf("Email %s for city %s reached max retry attempts (%d), giving up.",
					retryEntry.RecipientEmail, retryEntry.City, MAX_EMAIL_RETRIES)
				emailRetryMutex.Lock()
				delete(emailRetryQueue, email)
				emailRetryMutex.Unlock()
				continue
			}

			log.Printf("Attempting retry #%d for email %s for city %s...",
				retryEntry.Attempts+1, retryEntry.RecipientEmail, retryEntry.City)
			err := sendFormattedEmail(retryEntry.RecipientEmail, retryEntry.City, retryEntry.WeatherInfo)

			emailRetryMutex.Lock()
			if err != nil {
				retryEntry.NextAttempt = time.Now().Add(RETRY_DELAY)
				retryEntry.Attempts++
				emailRetryQueue[email] = retryEntry
				log.Printf("Retry #%d failed for %s. Next attempt in %v.",
					retryEntry.Attempts, retryEntry.RecipientEmail, RETRY_DELAY)
			} else {
				log.Printf("Retry #%d succeeded for %s. Removing from queue.",
					retryEntry.Attempts, retryEntry.RecipientEmail)
				delete(emailRetryQueue, email)
			}
			emailRetryMutex.Unlock()
		}

		log.Printf("Finished checking retry queue. Queue size: %d", len(emailRetryQueue))
	}
}

func sendFormattedEmail(recipientEmail, city, formattedBody string) error {
	if settings.SMTP.Adres == "" || settings.SMTP.Username == "" || settings.SMTP.Password == "" {
		log.Println("Warning: SMTP settings are incomplete. Cannot send email.")
		return fmt.Errorf("SMTP settings are not configured")
	}

	message := gomail.NewMessage()
	message.SetHeader("From", settings.SMTP.Username)
	message.SetHeader("To", recipientEmail)
	message.SetHeader("Subject", fmt.Sprintf("Weather Update for %s (Retry)", city))
	message.SetBody("text/html", formattedBody)

	dialer := gomail.NewDialer(settings.SMTP.Adres, settings.SMTP.Port, settings.SMTP.Username, settings.SMTP.Password)

	if err := dialer.DialAndSend(message); err != nil {
		log.Printf("Email send failed to %s for city %s: %v", recipientEmail, city, err)
		return fmt.Errorf("failed to send email to %s: %w", recipientEmail, err)
	}

	log.Printf("Email sent successfully to %s for city %s (Retry)", recipientEmail, city)
	return nil
}
