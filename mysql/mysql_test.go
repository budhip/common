package mysql_test

import (
	"encoding/base64"
	"testing"
)

func ca() []byte {
	strCA := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN1RENDQWwrZ0F3SUJBZ0lVR2hhdkpudzZTV2RJRFp2bVpQSWJyUjFETzJvd0NnWUlLb1pJemowRUF3SXcKZ2JFeEN6QUpCZ05WQkFZVEFrbEVNUlF3RWdZRFZRUUlEQXRFUzBrZ1NtRnJZWEowWVRFWU1CWUdBMVVFQnd3UApTbUZyWVhKMFlTQlRaV3hoZEdGdU1TRXdId1lEVlFRS0RCaFFWQ0JEWVhCcGRHRnNJRTVsZENCSmJtUnZibVZ6CmFXRXhFekFSQmdOVkJBc01DbFJsWTJodWIyeHZaM2t4RlRBVEJnTlZCQU1NREdSbGRpNWthVzFwYVM1cFpERWoKTUNFR0NTcUdTSWIzRFFFSkFSWVVjM2x6WVdSdGFXNUFZMkZ3YVhSaGJIZ3VhV1F3SGhjTk1qQXdOakE1TVRJMApOVEV3V2hjTk16QXdOakEzTVRJME5URXdXakNCc1RFTE1Ba0dBMVVFQmhNQ1NVUXhGREFTQmdOVkJBZ01DMFJMClNTQktZV3RoY25SaE1SZ3dGZ1lEVlFRSERBOUtZV3RoY25SaElGTmxiR0YwWVc0eElUQWZCZ05WQkFvTUdGQlUKSUVOaGNHbDBZV3dnVG1WMElFbHVaRzl1WlhOcFlURVRNQkVHQTFVRUN3d0tWR1ZqYUc1dmJHOW5lVEVWTUJNRwpBMVVFQXd3TVpHVjJMbVJwYldscExtbGtNU013SVFZSktvWklodmNOQVFrQkZoUnplWE5oWkcxcGJrQmpZWEJwCmRHRnNlQzVwWkRCWk1CTUdCeXFHU000OUFnRUdDQ3FHU000OUF3RUhBMElBQkdwZnZyL2doakVVK25uakRHakYKVXp1WUZSUE1JRzF5M0lXcVozMytYR0tTbERpTnVlclBFZGFuaG5ZclJWazd3b3RoNTZhYWp5VDVPbE9nakJTawpVWFdqVXpCUk1CMEdBMVVkRGdRV0JCUVBwS1F4bUl6N1dOOVZCVXVRbkxZMENnS2ppekFmQmdOVkhTTUVHREFXCmdCUVBwS1F4bUl6N1dOOVZCVXVRbkxZMENnS2ppekFQQmdOVkhSTUJBZjhFQlRBREFRSC9NQW9HQ0NxR1NNNDkKQkFNQ0EwY0FNRVFDSUhzMVFkVHl2U05PUE9NSWlMWERFZFA3dWl6VitEc0lwOGpzNStyMXVoMVZBaUJnMnRzcwpJYjJ2UGd3T1BJaVhXakVGZmxrdkFvakJ1WVJEdmNKbDBFcXR2Zz09Ci0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
	ca, _ := base64.StdEncoding.DecodeString(strCA)

	return ca
}

func TestDataSourceName(t *testing.T) {

	config := Config{
		Host:      "localhost",
		Port:      "3306",
		User:      "test",
		Password:  "test",
		Name:      "test",
		CA:        ca(),
		ParseTime: true,
		Location:  "Asia/Jakarta",
	}

	if got := dataSourceName(config); got == "" {
		t.Fatalf("bad config: %v", got)
	}
}

func TestDB(t *testing.T) {
	config := Config{
		Host:        "localhost",
		Port:        "3306",
		User:        "test",
		Password:    "test",
		Name:        "test",
		CA:          ca(),
		MaxOpen:     10,
		MaxIdle:     3,
		MaxLifetime: 3600,
		ParseTime:   true,
		Location:    "Asia/Jakarta",
	}

	if _, err := DB(config); err != nil {
		t.Fatalf("bad config: %v", err)
	}
}

