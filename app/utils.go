package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GetSubscribedUsers(frequency string) ([]User, error) {
	var users []User

	result := DB.Where("is_activated = ? AND frequency = ?", true, frequency).Find(&users)
	if result.Error != nil {

		return nil, fmt.Errorf("%w: error fetching subscribed users for frequency '%s': %v", ErrDatabaseOperation, frequency, result.Error)
	}
	return users, nil
}

func formatWeatherInfo(weather *ResponseOutput, city string) string {
	if weather == nil {
		return fmt.Sprintf("Could not retrieve weather information for %s.", city)
	}

	return fmt.Sprintf(
		"Latest weather update for %s:\n\n"+
			"Temperature: %d°C\n"+
			"Humidity: %d%%\n"+
			"Description: %s\n\n"+
			"Stay updated!",
		city, weather.Temperature, weather.Humidity, weather.Description,
	)
}

func GenerateToken() (string, error) {
	b := make([]byte, TOKEN_LENGHT)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to read random bytes: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(b)
	return token, nil
}

func GetCities() ([]string, error) {
	var cities []City
	result := DB.Find(&cities)
	if result.Error != nil {
		return nil, fmt.Errorf("%w: error fetching cities: %v", ErrDatabaseOperation, result.Error)
	}

	cityNames := make([]string, len(cities))
	for i, city := range cities {
		cityNames[i] = city.Name
	}

	return cityNames, nil
}
