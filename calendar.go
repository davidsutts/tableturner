package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"
)

const layout = "15.04 02 Jan 2006"

func generateUID(start, stop string) string {
	var str string
	for _, b := range []byte(start) {
		str += fmt.Sprintf("%02x", b)
	}
	str += "-"
	for _, b := range []byte(stop) {
		str += fmt.Sprintf("%02x", b)
	}

	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	str += "-" + fmt.Sprintf("%02x", sha256.Sum256(b))

	return str
}

func parseStartEnd(start, end, date string) (time.Time, time.Time, error) {
	loc, err := time.LoadLocation("Australia/Adelaide")
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("could not find timezone: %v", err)
	}
	st, err := time.ParseInLocation(layout, start+" "+date, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("could not parse start time: %v", err)
	}

	en, err := time.ParseInLocation(layout, end+" "+date, loc)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("could not parse end time: %v", err)
	}
	return st, en, nil
}
