package urls

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Counter struct {
	NumCalls map[Date]int64
	lock sync.RWMutex
}

type Date struct {
	Year int
	Month int
	Day int
}

func (d Date) MarshalText() ([]byte, error) {
    return []byte(fmt.Sprintf("%d-%d-%d", d.Year, d.Month, d.Day)), nil
}

func (d *Date) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), "-")

	if len(parts) != 3 {
		return errors.New("found date with incorrect format")
	}

	year, yerr := strconv.Atoi(parts[0])
	month, merr := strconv.Atoi(parts[1])
	day, derr := strconv.Atoi(parts[2])

	if yerr != nil || merr != nil || derr != nil {
		return errors.New("error converting string to integers")
	}


	d.Year = year
	d.Month = month
	d.Day = day
    return nil
}

func NewDate(year,month,day int) Date {
	return Date{
		Year: year,
		Month: month,
		Day: day,
	}
}

func NewCounter() *Counter {
	return &Counter{
		NumCalls: make(map[Date]int64),
	}
}

func (c *Counter) AddCall(timestamp time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()
	y, m, d := timestamp.Date()

	key := NewDate(int(y), int(m), int(d))

	if _, ok := c.NumCalls[key]; !ok {
		// initialize key if it doesn't exist
		c.NumCalls[key] = 0
	}

	c.NumCalls[key] += 1
}

// Todo: In the future this function should return each summary type separately
// i.e. scanning over all timestamps is inefficient and we only really need to store
// timestamps if they occured in the last 7 days and then keep a simple count for total calls.
// or replace this with an otel instrumentation?
func (c *Counter) GetSummary() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	nowTime := time.Now()
	nowYear, nowMonth, nowDay := nowTime.Date()
	dayTotal := int64(0)
	weekTotal := int64(0)
	allTotal := int64(0)

	for key, val := range c.NumCalls {
		allTotal += val

		if key.Year == int(nowYear) && key.Month == int(nowMonth) && key.Day == int(nowDay) {
			dayTotal += val
		}

		// this is a crude measurement of the past week
		keyTimeStamp := time.Date(key.Year, time.Month(key.Month), key.Day, 0, 0, 0, 0, time.Local)

		if keyTimeStamp.Before(nowTime) && nowTime.Add(-7 * 24 * time.Hour).Before(keyTimeStamp) {
			weekTotal += val
		}
	}

	return fmt.Sprintf("Summary of shorturl:\n calls in the last day: %d calls\n calls in the last week: %d calls\n total calls since creation: %d calls\n", dayTotal, weekTotal, allTotal)
}