package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func WebsiteHandler(res http.ResponseWriter, req *http.Request) {

	http.ServeFile(res, req, "index.html")
}

func WeatherHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	city := req.URL.Query().Get("city")
	if city == "" {
		http.Error(res, "Missing 'city' query parameter", http.StatusBadRequest)
		return
	}

	weather, err := GetWeatherData(city)
	if err != nil {
		if err == ErrCityNotFound {
			http.Error(res, "City not found or no weather data available", http.StatusNotFound)
		} else {
			log.Printf("Error getting weather data for city '%s': %v", city, err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	responseBody, err := json.Marshal(weather)
	if err != nil {
		log.Printf("Error marshalling weather response: %v", err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(responseBody)
}

func GetCitiesHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cityNames, err := GetCities()
	if err != nil {
		log.Printf("Error getting cities: %v", err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseBody, err := json.Marshal(cityNames)
	if err != nil {
		log.Printf("Error marshalling cities response: %v", err)
		http.Error(res, "Internal server error", http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.Write(responseBody)
}

func SubscriptionHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Error reading request body", http.StatusBadRequest)
		return
	}

	var data InputEmailData
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(res, "Error parsing request body", http.StatusBadRequest)
		return
	}

	err = SubscribeUser(data.Email, data.Frequency, data.City)
	if err != nil {
		if err == ErrEmailAlreadyExists {
			http.Error(res, "Email already subscribed", http.StatusConflict)
		} else if err == ErrInvalidFrequency {
			http.Error(res, "Invalid frequency provided", http.StatusBadRequest)
		} else if err == ErrCityNotFound {
			http.Error(res, "City not found for subscription", http.StatusBadRequest)
		} else {
			log.Printf("Error subscribing user: %v", err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte("Subscription successful. Please check your email for confirmation."))
}

func ConfirmSubscriptionHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathSegments := strings.Split(req.URL.Path, "/")
	if len(pathSegments) != 3 || pathSegments[0] != "" || pathSegments[1] != "confirm" {
		http.Error(res, "Invalid confirmation URL", http.StatusBadRequest)
		return
	}
	token := pathSegments[2]
	if token == "" {
		http.Error(res, "Missing confirmation token", http.StatusBadRequest)
		return
	}

	err := ConfirmUser(token)
	if err != nil {
		if err == ErrInvalidOrExpiredToken {
			http.Error(res, "Invalid or expired confirmation token", http.StatusNotFound)
		} else {
			log.Printf("Error confirming user: %v", err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Email confirmed successfully!"))
}

func UnsubscribeHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(res, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	pathSegments := strings.Split(req.URL.Path, "/")
	if len(pathSegments) != 3 || pathSegments[0] != "" || pathSegments[1] != "unsubscribe" {
		http.Error(res, "Invalid unsubscribe URL", http.StatusBadRequest)
		return
	}
	token := pathSegments[2]
	if token == "" {
		http.Error(res, "Missing unsubscribe token", http.StatusBadRequest)
		return
	}

	err := UnsubscribeUser(token)
	if err != nil {
		if err == ErrInvalidOrExpiredToken {
			http.Error(res, "Invalid or expired unsubscribe token", http.StatusNotFound)
		} else {
			log.Printf("Error unsubscribing user: %v", err)
			http.Error(res, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	res.WriteHeader(http.StatusOK)
	res.Write([]byte("Unsubscribed successfully."))
}
