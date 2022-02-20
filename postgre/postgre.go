package postgre

import (
	"fmt"
	"time"

	"database/sql"
	"net/url"
)

const (
	// DefaultMaxOpen is default value for max open connection
	DefaultMaxOpen = 10
	// DefaultMaxIdle is default value for max idle connection
	DefaultMaxIdle = 10
	// DefaultMaxLifetime is default value for max connection lifetime in minutes
	DefaultMaxLifetime = 3
)

type Config struct {
	Host        string
	Port        string
	User        string
	Password    string
	Name        string
	MaxOpen     int
	MaxIdle     int
	MaxLifetime int // in minutes
	MaxIdleTime int // in minutes
	ServerName  string
	ParseTime   bool
	Location    string
}

func dataSourceName(config Config) string {
	connection := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.Host, config.Port, config.User, config.Password, config.Name)
	val := url.Values{}

	if config.ParseTime {
		val.Add("parseTime", "1")
	}
	if len(config.Location) > 0 {
		val.Add("loc", config.Location)
	}

	if len(val) == 0 {
		return connection
	}
	return fmt.Sprintf("%s?%s", connection, val.Encode())
}

// DB return new sql db
func DB(config Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", dataSourceName(config))
	if err != nil {
		return nil, err
	}

	if config.MaxOpen > 0 {
		db.SetMaxOpenConns(config.MaxOpen)
	} else {
		db.SetMaxOpenConns(DefaultMaxOpen)
	}

	if config.MaxIdle > 0 {
		db.SetMaxIdleConns(config.MaxIdle)
	} else {
		db.SetMaxIdleConns(DefaultMaxIdle)
	}

	if config.MaxLifetime > 0 {
		db.SetConnMaxLifetime(time.Duration(config.MaxLifetime) * time.Minute)
	} else {
		db.SetConnMaxLifetime(time.Duration(DefaultMaxLifetime) * time.Minute)
	}

	// if config.MaxIdleTime > 0 {
	// 	db.SetConnMaxIdleTime(time.Duration(config.MaxIdleTime) * time.Minute)
	// }

	return db, nil
}