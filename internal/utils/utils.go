package utils

import (
	"fmt"
	"time"
	"unicode"
)

// PrettyDuration returns a pretty version of the highest non-zero time unit
// 2 years and 3 months -> "2 years"
// 3 days and 17 hours -> "3 days"
// 1 hour and 7 minutes -> "1 hour"
func PrettyDuration(d time.Duration) string {
	const (
		microsecond = time.Microsecond
		millisecond = time.Millisecond
		second      = time.Second
		minute      = time.Minute
		hour        = time.Hour
		day         = hour * 24
		week        = day * 7
		month       = day * 30
		year        = month * 12
	)

	switch {
	case d >= year:
		years := d / year
		return fmt.Sprintf("%d year%s", years, PluraliseInt(int(years)))
	case d >= month:
		months := d / month
		return fmt.Sprintf("%d month%s", months, PluraliseInt(int(months)))
	case d >= week:
		weeks := d / week
		return fmt.Sprintf("%d week%s", weeks, PluraliseInt(int(weeks)))
	case d >= day:
		days := d / day
		return fmt.Sprintf("%d day%s", days, PluraliseInt(int(days)))
	case d >= hour:
		hours := d / hour
		return fmt.Sprintf("%d hour%s", hours, PluraliseInt(int(hours)))
	case d >= minute:
		minutes := d / minute
		return fmt.Sprintf("%d minute%s", minutes, PluraliseInt(int(minutes)))
	case d >= second:
		seconds := d / second
		return fmt.Sprintf("%d second%s", seconds, PluraliseInt(int(seconds)))
	case d >= millisecond:
		milliseconds := d / millisecond
		return fmt.Sprintf("%d ms", milliseconds)
	case d >= microsecond:
		microseconds := d / microsecond
		return fmt.Sprintf("%d μs", microseconds)
	}
	return "now"
}

// Returns an 's' if i is greater than 1.
// Example usage: fmt.Sprintf("%d second%s", num, PluraliseInt(num))
func PluraliseInt(i int) string {
	if i == 1 {
		return ""
	}
	return "s"
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

func CapitalizeFirstCharacter(s string) string {
	r := []rune(s)
	r[0] = unicode.ToTitle(r[0])
	return string(r)
}
