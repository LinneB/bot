package utils

import (
	"fmt"
	"time"
)

// PrettyDuration returns a pretty version of the highest non-zero time unit
// 2 years and 3 months -> "2 years"
// 3 days and 17 hours -> "3 days"
// 1 hour and 7 minutes -> "1 hour"
func PrettyDuration(d time.Duration) string {
	micros := d.Microseconds()
	millis := d.Milliseconds()
	seconds := d.Seconds()
	minutes := d.Minutes()
	hours := d.Hours()
	days := hours / 24.0
	months := days / 30.0
	years := months / 12.0

	if years >= 2 {
		return fmt.Sprintf("%d years", int(years))
	}
	if years >= 1 {
		return fmt.Sprintf("%d year", int(years))
	}
	if months >= 2 {
		return fmt.Sprintf("%d months", int(months))
	}
	if months >= 1 {
		return fmt.Sprintf("%d month", int(months))
	}
	if days >= 2 {
		return fmt.Sprintf("%d days", int(days))
	}
	if days >= 1 {
		return fmt.Sprintf("%d day", int(days))
	}
	if hours >= 2 {
		return fmt.Sprintf("%d hours", int(hours))
	}
	if hours >= 1 {
		return fmt.Sprintf("%d hour", int(hours))
	}
	if minutes >= 2 {
		return fmt.Sprintf("%d minutes", int(minutes))
	}
	if minutes >= 1 {
		return fmt.Sprintf("%d minute", int(minutes))
	}
	if seconds >= 2 {
		return fmt.Sprintf("%d seconds", int(seconds))
	}
	if seconds >= 1 {
		return fmt.Sprintf("%d second", int(seconds))
	}
	if millis >= 1 {
		return fmt.Sprintf("%d ms", int(millis))
	}
	if micros >= 1 {
		return fmt.Sprintf("%d μs", int(micros))
	}
	return ""
}

func SplitStreamOnlineMessage(message string, users []string, length int) (messages []string) {
	buf := message
	for _, user := range users {
		combinedMessage := fmt.Sprintf(`%s @%s`, buf, user)
		if len(combinedMessage) > length {
			messages = append(messages, buf)
			buf = fmt.Sprintf("@%s", user)
		} else {
			buf = combinedMessage
		}
	}
	messages = append(messages, buf)
	return messages
}
