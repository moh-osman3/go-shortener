package urls

import (
	"fmt"
	"time"
)

type counter struct {
	numCalls map[Date]int64
}

type Date struct {
	year int
	month int
	day int

}
func NewDate(year,month,day int) Date {
	return Date{
		year: year,
		month: month,
		day: day,
	}
}

func NewCounter() *counter {
	return &counter{
		numCalls: make(map[Date]int64),
	}
}

func (c *counter) AddCall(timestamp time.Time) {
	y, m, d := timestamp.Date()

	key := NewDate(int(y), int(m), int(d))

	if _, ok := c.numCalls[key]; !ok {
		// initialize key if it doesn't exist
		c.numCalls[key] = 0
	}

	c.numCalls[key] += 1
}

// Todo: In the future this function should return each summary type separately
// i.e. scanning over all timestamps is inefficient and we only really need to store
// timestamps if they occured in the last 7 days and then keep a simple count for total calls.
// or replace this with an otel instrumentation?
func (c *counter) GetSummary() string {
	nowTime := time.Now()
	nowYear, nowMonth, nowDay := nowTime.Date()
	dayTotal := int64(0)
	weekTotal := int64(0)
	allTotal := int64(0)

	for key, val := range c.numCalls {
		allTotal += val

		if key.year == int(nowYear) && key.month == int(nowMonth) && key.day == int(nowDay) {
			dayTotal += val
		}

		// this is a crude measurement of the past week
		keyTimeStamp := time.Date(key.year, time.Month(key.month), key.day, 0, 0, 0, 0, time.Local)

		if keyTimeStamp.Before(nowTime) && nowTime.Add(-7 * 24 * time.Hour).Before(keyTimeStamp) {
			weekTotal += val
		}
	}

	return fmt.Sprintf("Summary of shorturl:\n calls in the last day: %d calls\n calls in the last week: %d calls\n total calls since creation: %d calls\n", dayTotal, weekTotal, allTotal)
}