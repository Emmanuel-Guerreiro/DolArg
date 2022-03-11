package main

import (
	"strings"
	"time"
)

func ISOTimestamp() string {
	t := time.Now().String()
	times := strings.Split(t, " ")[:2]
	ISOT := strings.Join(times, " ")

	return ISOT
}
