package postgre

import (
	"database/sql"
	"fmt"
	"net/url"
)

type Config struct {
	Host        string
	Port        string
	User        string
	Password    string
	Name        string
}

func dataSourceName(config Config) string {
	connection := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.Name)
	val := url.Values{}

	return fmt.Sprintf("%s?%s", connection, val.Encode())
}

// DB return new sql db
func DB(config Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName(config))
	if err != nil {
		return nil, err
	}

	return db, nil
}