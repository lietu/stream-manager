package utils

import (
	"log"
	"time"
)

var duplicateDurationLimit = 10 * time.Minute // How frequently is an identical message ok?
var lastLogs = map[string]string{}            // List of last messages per category
var lastLogTimes = map[string]time.Time{}     // List of when last message in each category was

func StatusLog(category string, status string) {
	old, ok := lastLogs[category]
	if !ok || old != status {
		showStatusLog(category, status)
	} else if durationLimitExceeded(category) {
		showStatusLog(category, status)
	}

	// Ignored, identical as the last message, and time limit hasn't passed
}

func durationLimitExceeded(category string) bool {
	return time.Since(lastLogTimes[category]) > duplicateDurationLimit
}

func showStatusLog(category string, status string) {
	lastLogs[category] = status
	lastLogTimes[category] = time.Now()
	log.Print(status)
}
