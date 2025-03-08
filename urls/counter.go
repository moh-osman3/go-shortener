package urls

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

const secondsInDay = 86400

type Count struct {
	count    int64
	lastUnix int64
}

type Counter struct {
	WeekBuffer []Count
	TotalCalls int64
	lock       sync.RWMutex
}

func (c Count) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf("%d:%d", c.count, c.lastUnix)), nil
}

func (c *Count) UnmarshalText(text []byte) error {
	parts := strings.Split(string(text), ":")

	if len(parts) != 2 {
		return errors.New("found text with incorrect format")
	}

	count, cerr := strconv.ParseInt(parts[0], 10, 64)
	lastUnix, lerr := strconv.ParseInt(parts[1], 10, 64)

	if cerr != nil || lerr != nil {
		return errors.New("error converting string to integers")
	}
	c.count = count
	c.lastUnix = lastUnix
	return nil
}

func NewCounter() *Counter {
	return &Counter{
		TotalCalls: 0,
		WeekBuffer: make([]Count, 7),
	}
}

func (c *Counter) AddCall(timestamp time.Time) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.TotalCalls += 1

	// calculate index in ring buffer by days_since_unix_epoch % 7
	seconds := timestamp.Unix()
	days := seconds / secondsInDay

	key := days % 7

	isStale := (seconds-secondsInDay > c.WeekBuffer[key].lastUnix)

	if isStale {
		c.WeekBuffer[key].count = 0
	}

	c.WeekBuffer[key].count += 1
	c.WeekBuffer[key].lastUnix = seconds
}

// Todo: In the future this function should return each summary type separately
// i.e. scanning over all timestamps is inefficient and we only really need to store
// timestamps if they occured in the last 7 days and then keep a simple count for total calls.
// or replace this with an otel instrumentation?
func (c *Counter) GetSummary() string {
	c.lock.RLock()
	defer c.lock.RUnlock()
	nowTime := time.Now()
	nowSeconds := nowTime.Unix()
	dayTotal := int64(0)
	weekTotal := int64(0)
	allTotal := c.TotalCalls

	for _, count := range c.WeekBuffer {
		if nowSeconds-secondsInDay < count.lastUnix {
			dayTotal = count.count
		}

		if nowSeconds-7*secondsInDay < count.lastUnix {
			weekTotal += count.count

		}
	}

	return fmt.Sprintf("Summary of shorturl:\n calls in the last day: %d calls\n calls in the last week: %d calls\n total calls since creation: %d calls\n", dayTotal, weekTotal, allTotal)
}
