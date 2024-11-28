package helpers

import (
	"log"
	"time"
)

func GetTimefromStr(str string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05Z", str)
	if err != nil {
		log.Println("Error parsing time!")
		return time.Unix(0, 0)
	} else {
		return t
	}
}
