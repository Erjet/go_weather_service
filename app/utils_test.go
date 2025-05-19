package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newMockGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqldb, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqldb,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm connection: %v", err)
	}

	return gormDB, mock
}

func setMockDB(mockDB *gorm.DB) {
	DB = mockDB
}

func TestGetSubscribedUsers_Success(t *testing.T) {
	mockDB, mock := newMockGormDB(t)
	setMockDB(mockDB)

	frequency := "daily"

	expectedUsers := []User{
		{ID: uuid.New(), Email: "user1@example.com", IsActivated: true, Token: "token1", Frequency: "daily", City: "Kyiv"},
		{ID: uuid.New(), Email: "user2@example.com", IsActivated: true, Token: "token2", Frequency: "daily", City: "London"},
	}

	rows := sqlmock.NewRows([]string{"id", "email", "is_activated", "token", "frequency", "city"})
	for _, user := range expectedUsers {
		rows.AddRow(user.ID, user.Email, user.IsActivated, user.Token, user.Frequency, user.City)
	}

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE is_activated = \$1 AND frequency = \$2`).
		WithArgs(true, frequency).
		WillReturnRows(rows)

	users, err := GetSubscribedUsers(frequency)

	if err != nil {
		t.Errorf("GetSubscribedUsers returned an unexpected error: %v", err)
	}
	if len(users) != len(expectedUsers) {
		t.Errorf("GetSubscribedUsers returned incorrect number of users. Got %d, Expected %d", len(users), len(expectedUsers))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetSubscribedUsers_NoUsersFound(t *testing.T) {
	mockDB, mock := newMockGormDB(t)
	setMockDB(mockDB)

	frequency := "hourly"

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE is_activated = \$1 AND frequency = \$2`).
		WithArgs(true, frequency).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "is_activated", "token", "frequency", "city"}))

	users, err := GetSubscribedUsers(frequency)

	if err != nil {
		t.Errorf("GetSubscribedUsers returned an unexpected error: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("GetSubscribedUsers returned users when none were expected. Got %d", len(users))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetSubscribedUsers_DatabaseError(t *testing.T) {
	mockDB, mock := newMockGormDB(t)
	setMockDB(mockDB)

	frequency := "daily"
	dbError := errors.New("simulated database error")

	mock.ExpectQuery(`SELECT \* FROM "users" WHERE is_activated = \$1 AND frequency = \$2`).
		WithArgs(true, frequency).
		WillReturnError(dbError)

	users, err := GetSubscribedUsers(frequency)

	if users != nil {
		t.Errorf("GetSubscribedUsers returned users when an error was expected: %+v", users)
	}
	if err == nil {
		t.Error("GetSubscribedUsers did not return an error when a database error occurred")
	}

	if !errors.Is(err, ErrDatabaseOperation) || !strings.Contains(err.Error(), dbError.Error()) {
		t.Errorf("GetSubscribedUsers returned incorrect error. Got %v, Expected to wrap %v and contain '%v'", err, ErrDatabaseOperation, dbError)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestFormatWeatherInfo_WithData(t *testing.T) {
	weather := &ResponseOutput{
		Temperature: 25,
		Humidity:    50,
		Description: "Sunny",
	}
	city := "Paris"

	expectedOutput := "Latest weather update for Paris:\n\nTemperature: 25°C\nHumidity: 50%\nDescription: Sunny\n\nStay updated!"

	formattedString := formatWeatherInfo(weather, city)

	if formattedString != expectedOutput {
		t.Errorf("formatWeatherInfo returned incorrect string.\nGot:\n%s\nExpected:\n%s", formattedString, expectedOutput)
	}
}

func TestFormatWeatherInfo_NilWeather(t *testing.T) {
	var weather *ResponseOutput = nil
	city := "Berlin"

	expectedOutputSubstring := "Could not retrieve weather information for Berlin."

	formattedString := formatWeatherInfo(weather, city)

	if !strings.Contains(formattedString, expectedOutputSubstring) {
		t.Errorf("formatWeatherInfo returned incorrect string for nil weather.\nGot:\n%s\nExpected to contain:\n%s", formattedString, expectedOutputSubstring)
	}
}

func TestGenerateToken_Success(t *testing.T) {

	token, err := GenerateToken()

	if err != nil {
		t.Errorf("GenerateToken returned an unexpected error: %v", err)
	}
	if len(token) == 0 {
		t.Error("GenerateToken returned an empty token")
	}
	if len(token) < 40 {
		t.Errorf("GenerateToken returned a token that is too short. Got length %d", len(token))
	}
}

func TestGetCities_Success(t *testing.T) {
	mockDB, mock := newMockGormDB(t)
	setMockDB(mockDB)

	expectedCities := []string{"Kyiv", "London", "New York"}

	rows := sqlmock.NewRows([]string{"id", "name"})
	for i, city := range expectedCities {
		rows.AddRow(uint(i+1), city)
	}

	mock.ExpectQuery(`SELECT \* FROM "cities"`).
		WillReturnRows(rows)

	cities, err := GetCities()

	if err != nil {
		t.Errorf("GetCities returned an unexpected error: %v", err)
	}
	if len(cities) != len(expectedCities) {
		t.Errorf("GetCities returned incorrect number of cities. Got %d, Expected %d", len(cities), len(expectedCities))
	}
	for i, city := range cities {
		if city != expectedCities[i] {
			t.Errorf("GetCities returned incorrect city name at index %d. Got %s, Expected %s", i, city, expectedCities[i])
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetCities_NoCitiesFound(t *testing.T) {
	mockDB, mock := newMockGormDB(t)
	setMockDB(mockDB)

	mock.ExpectQuery(`SELECT \* FROM "cities"`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))

	cities, err := GetCities()

	if err != nil {
		t.Errorf("GetCities returned an unexpected error: %v", err)
	}
	if len(cities) != 0 {
		t.Errorf("GetCities returned cities when none were expected. Got %d", len(cities))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}

func TestGetCities_DatabaseError(t *testing.T) {
	mockDB, mock := newMockGormDB(t)
	setMockDB(mockDB)

	dbError := errors.New("simulated database error")

	mock.ExpectQuery(`SELECT \* FROM "cities"`).
		WillReturnError(dbError)

	cities, err := GetCities()

	if cities != nil {
		t.Errorf("GetCities returned cities when an error was expected: %+v", cities)
	}
	if err == nil {
		t.Error("GetCities did not return an error when a database error occurred")
	}

	if !errors.Is(err, ErrDatabaseOperation) || !strings.Contains(err.Error(), dbError.Error()) {
		t.Errorf("GetCities returned incorrect error. Got %v, Expected to wrap %v and contain '%v'", err, ErrDatabaseOperation, dbError)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
