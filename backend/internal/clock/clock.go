package clock

import "time"

// Beijing is the sole civil-time zone used by GoPuppy. Day-boundary and age
// math must never use time.Now().UTC() date parts.
var Beijing = time.FixedZone("CST", 8*3600)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func Today() time.Time {
	n := Now()
	return time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, Beijing)
}

func DateOf(t time.Time) time.Time {
	t = t.In(Beijing)
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, Beijing)
}

func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", s, Beijing)
}

func FormatDate(t time.Time) string {
	return t.In(Beijing).Format("2006-01-02")
}

func FormatDateTime(t time.Time) string {
	return t.In(Beijing).Format("2006-01-02 15:04:05")
}

func Hour(t time.Time) int {
	return t.In(Beijing).Hour()
}
