package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	settings = serverInit()

	if err := InitDatabase(settings.DBConnectionString); err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err == nil && sqlDB != nil {
		defer sqlDB.Close()
	}

	finalSettings := loadFinalSettings()

	settings.Port = finalSettings.Port
	settings.EmailSubscribeText = finalSettings.EmailSubscribeText
	settings.LocalURL = finalSettings.LocalURL
	settings.SMTP = finalSettings.SMTP

	//TODO: В розширеній імплементації цей механізм слід замінити повноцінним планувальником який підтримує логіку виконання пропущених завдань, наприклад, go-quartz.

	hourlyTicker := time.NewTicker(1 * time.Hour)
	go func() {
		SendWeatherUpdates("hourly")
		for range hourlyTicker.C {
			SendWeatherUpdates("hourly")
		}
	}()

	dailyTicker := time.NewTicker(24 * time.Hour)
	go func() {
		SendWeatherUpdates("daily")
		for range dailyTicker.C {
			SendWeatherUpdates("daily")
		}
	}()

	go startEmailRetryScheduler()

	http.HandleFunc("/", WebsiteHandler)
	http.HandleFunc("/weather", WeatherHandler)
	http.HandleFunc("/subscribe", SubscriptionHandler)
	http.HandleFunc("/confirm/", ConfirmSubscriptionHandler)
	http.HandleFunc("/unsubscribe/", UnsubscribeHandler)
	http.HandleFunc("/getCities", GetCitiesHandler)

	log.Printf("Server up on port %s", settings.Port)
	log.Fatal(http.ListenAndServe(settings.Port, nil))
}

func serverInit() ServerSettings {
	log.Println("Initializing database connection string...")
	log.Println("Attempting to load DBConnectionString from environment variable (DATABASE_URL)...")

	var loadedSettings ServerSettings
	dbURL, exists := os.LookupEnv("DATABASE_URL")
	if exists && dbURL != "" {
		loadedSettings.DBConnectionString = dbURL
		log.Println("DBConnectionString loaded from environment.")
		return loadedSettings
	}
	log.Println("DBConnectionString not found in environment.")
	log.Println("Attempting to load DBConnectionString from settings.json...")

	fileSettings, err := loadSettingsFromFile("settings.json")
	if err == nil && fileSettings.DBConnectionString != "" {
		loadedSettings.DBConnectionString = fileSettings.DBConnectionString
		log.Println("DBConnectionString loaded from settings.json.")
		return loadedSettings
	}
	if err != nil {
		log.Printf("Error reading settings.json for DBConnectionString: %v", err)
	} else {
		log.Println("DBConnectionString not found in settings.json.")
	}

	loadedSettings = applyDefaultSettings()

	log.Println("DBConnectionString not found in environment or settings.json. Using defalt parameters, InitDatabase may fail.")
	return ServerSettings{}
}

func loadFinalSettings() ServerSettings {
	var finalsettings ServerSettings

	log.Println("Attempting to load settings from database...")
	dbSettings, err := loadFinalSettingsFromDB()
	if err == nil && dbSettings != nil {
		log.Println("Settings loaded from database. Overriding file/defaults...")
		mergeSettings(&finalsettings, *dbSettings)
	} else if errors.Is(err, ErrSettingsNotFoundInDB) {
		log.Println("Settings table empty or not found in database. Skipping database load.")
	} else {
		log.Printf("Error loading settings from database: %v. This might indicate a DB issue.", err)
	}

	log.Println("Attempting to load settings from settings.json...")
	fileSettings, err := loadSettingsFromFile("settings.json")
	if err == nil {
		log.Println("settings.json found. Overriding defaults...")
		mergeSettings(&finalsettings, fileSettings)
	} else {
		log.Printf("settings.json not found or error reading it: %v. Skipping file load.", err)
	}

	log.Println("Checking for PORT environment variable override...")
	envPort, exists := os.LookupEnv("PORT")
	if exists && envPort != "" {
		if !strings.HasPrefix(envPort, ":") {
			envPort = ":" + envPort
		}
		finalsettings.Port = envPort
		log.Println("Port overridden by PORT environment variable.")
	}

	log.Println("Settings initialization complete.")
	return finalsettings
}

func loadSettingsFromFile(filename string) (ServerSettings, error) {
	var fileSettings ServerSettings

	settingsFile, err := os.Open(filename)
	if err != nil {
		return fileSettings, fmt.Errorf("error opening settings file %s: %w", filename, err)
	}
	defer settingsFile.Close()

	settingsFileByte, err := io.ReadAll(settingsFile)
	if err != nil {
		return fileSettings, fmt.Errorf("error reading settings file %s: %w", filename, err)
	}

	err = json.Unmarshal(settingsFileByte, &fileSettings)
	if err != nil {
		return fileSettings, fmt.Errorf("error unmarshalling settings file %s: %w", filename, err)
	}

	return fileSettings, nil
}

func applyDefaultSettings() ServerSettings {
	log.Println("Applying hardcoded default settings.")
	return ServerSettings{
		Port:               ":8080",
		EmailSubscribeText: "Thank you for subscribing to our weather report service! \n\nTo confirm subscription please visit this page {1}\nIf you want to unsubscribe visit this page {2}",
		LocalURL:           "http://localhost:8080",
		SMTP: SMTPSettings{
			Adres:    "smtp.gmail.com",
			Port:     587,
			Username: "pikkljuihilhu@gmail.com",
			Password: "ueyg cmai exsb rsxm",
		},
	}
}

func loadFinalSettingsFromDB() (*ServerSettings, error) {
	var loadedSettings ServerSettings
	settingsMap := make(map[string]string)

	rows, err := DB.Table("application_settings").Select("key, value").Rows()
	if err != nil {
		log.Printf("Warning: Could not load settings from database: %v", err)
		return nil, ErrSettingsNotFoundInDB
	}
	defer rows.Close()

	foundSettings := false
	for rows.Next() {
		var key string
		var value string
		if err := rows.Scan(&key, &value); err != nil {
			log.Printf("Error scanning settings row from database: %v", err)
			continue
		}
		settingsMap[key] = value
		foundSettings = true
	}

	if !foundSettings && err == nil {

		return nil, ErrSettingsNotFoundInDB
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error during rows iteration for database settings: %v", err)
		return nil, ErrSettingsNotFoundInDB
	}

	if port, ok := settingsMap["Port"]; ok {
		loadedSettings.Port = port
	}
	if EmailSubscribeText, ok := settingsMap["EmailSubscribeText"]; ok {
		loadedSettings.EmailSubscribeText = EmailSubscribeText
	}
	if localURL, ok := settingsMap["LocalURL"]; ok {
		loadedSettings.LocalURL = localURL
	}
	if smtpAdres, ok := settingsMap["SMTP_Adres"]; ok {
		loadedSettings.SMTP.Adres = smtpAdres
	}
	if smtpPortStr, ok := settingsMap["SMTP_Port"]; ok {
		if smtpPort, err := strconv.Atoi(smtpPortStr); err == nil {
			loadedSettings.SMTP.Port = smtpPort
		} else {
			log.Printf("Warning: Invalid SMTP_Port value from database: %s, error: %v", smtpPortStr, err)
		}
	}
	if smtpUsername, ok := settingsMap["SMTP_Username"]; ok {
		loadedSettings.SMTP.Username = smtpUsername
	}
	if smtpPassword, ok := settingsMap["SMTP_Password"]; ok {
		loadedSettings.SMTP.Password = smtpPassword
	}

	// Повернути налаштування, навіть якщо деякі ключі були відсутні (будуть використані нульові значення)
	return &loadedSettings, nil
}

func mergeSettings(dest *ServerSettings, source ServerSettings) {

	if source.Port != "" {
		dest.Port = source.Port
	}
	if source.DBConnectionString != "" {
		dest.DBConnectionString = source.DBConnectionString
	}
	if source.EmailSubscribeText != "" {
		dest.EmailSubscribeText = source.EmailSubscribeText
	}
	if source.LocalURL != "" {
		dest.LocalURL = source.LocalURL
	}

	if source.SMTP.Adres != "" {
		dest.SMTP.Adres = source.SMTP.Adres
	}
	if source.SMTP.Port != 0 {
		dest.SMTP.Port = source.SMTP.Port
	}
	if source.SMTP.Username != "" {
		dest.SMTP.Username = source.SMTP.Username
	}
	if source.SMTP.Password != "" {
		dest.SMTP.Password = source.SMTP.Password
	}
}

func InitDatabase(dsn string) error {
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("error opening database connection with GORM: %w", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("error getting underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("error pinging database after GORM connection: %w", err)
	}

	log.Println("Database initialized successfully with GORM")
	return nil
}
