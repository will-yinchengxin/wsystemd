package utils

import "time"

func GetCTime() string {
	return time.Now().Format(time.RFC3339)
}
