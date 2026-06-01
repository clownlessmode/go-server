package domain

import "time"

const beelineTimezoneOffsetSeconds = 3 * 60 * 60

func BeelineLocation() *time.Location {
	return time.FixedZone("MSK", beelineTimezoneOffsetSeconds)
}

func FormatBeelineDateTime(value time.Time) string {
	return value.In(BeelineLocation()).Format("2006-01-02T15:04:05")
}

func FormatBeelineDateTimeRFC3339(value time.Time) string {
	return value.In(BeelineLocation()).Format("2006-01-02T15:04:05-07:00")
}

func NowInBeeline() time.Time {
	return time.Now().In(BeelineLocation())
}
