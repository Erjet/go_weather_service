package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResponseOutput struct {
	Temperature int    `json:"temperature"`
	Humidity    int    `json:"humidity"`
	Description string `json:"description"`
}

type InputWeatherData struct {
	City string `json:"city"`
}

type InputEmailData struct {
	Email     string `json:"email"`
	City      string `json:"city"`
	Frequency string `json:"frequency"`
}

type City struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
}

type WeatherDescription struct {
	ID          uint   `gorm:"primaryKey"`
	Description string `gorm:"uniqueIndex"`
}

type Weather struct {
	ID              uint `gorm:"primaryKey"`
	CityID          uint
	City            City `gorm:"foreignKey:CityID"`
	DescriptionID   uint
	Description     WeatherDescription `gorm:"foreignKey:DescriptionID"`
	Temperature     int
	Humidity        int
	ObservationTime time.Time
}

// Перевизначення назви таблиці Wether для GORM
func (Weather) TableName() string {
	return "weather"
}

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email       string    `gorm:"uniqueIndex"`
	IsActivated bool      `gorm:"default:false"`
	Token       string    `gorm:"uniqueIndex"`
	Frequency   string
	City        string
}

type EmailRetry struct {
	RecipientEmail string
	City           string
	WeatherInfo    string
	NextAttempt    time.Time
	Attempts       int
}

type ServerSettings struct {
	Port               string `json:"Port"`
	DBConnectionString string `json:"DBConnectionString"`
	EmailSubscribeText string `json:"EmailSubscribeText"`
	EmailUpdateText    string `json:"EmailUpdateText"`
	LocalURL           string `json:"LocalURL"`
	IsSelfAutoMigrate  bool   `json:"IsSelfAutoMigrate"`
	SMTP               SMTPSettings
}
type SMTPSettings struct {
	Adres    string `json:"Adres"`
	Port     int    `json:"Port"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

var emailRetryQueue = make(map[string]EmailRetry)
var emailRetryMutex sync.Mutex

var DB *gorm.DB

var settings ServerSettings

const TOKEN_LENGHT int = 32
const MAX_EMAIL_RETRIES = 5
const RETRY_DELAY = 5 * time.Minute

var ErrEmailAlreadyExists = fmt.Errorf("email already exists")
var ErrInvalidFrequency = fmt.Errorf("invalid frequency")
var ErrCityNotFound = fmt.Errorf("city not found")
var ErrInvalidOrExpiredToken = fmt.Errorf("invalid or expired token")
var ErrDatabaseOperation = fmt.Errorf("database operation failed")
var ErrSettingsNotFoundInDB = fmt.Errorf("settings not found in database")
