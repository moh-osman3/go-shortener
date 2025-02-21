package urls

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEmptyCounter(t *testing.T) {
	c := NewCounter()
	expectedSummary := fmt.Sprintf("Summary of shorturl:\n calls in the last day: %d calls\n calls in the last week: %d calls\n total calls since creation: %d calls\n", 0, 0, 0)
	assert.Equal(t, expectedSummary, c.GetSummary())
}

func TestCounter(t *testing.T) {
	c := NewCounter()

	expectedAll := 10
	expectedWeek := 10
	expectedDay := 10
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < expectedAll; i++ {
		// pick a day longer than a week ago
		randDay := 8 + rand.Intn(50)
		timestamp1 := time.Now().Add(-time.Duration(randDay) * 24 * time.Hour)
		c.AddCall(timestamp1)
	}
	for i := 0; i < expectedWeek; i++ {
		// pick a day longer one day and less than a week ago
		randDay := 2 + rand.Intn(5)
		timestamp1 := time.Now().Add(-time.Duration(randDay) * 24 * time.Hour)
		c.AddCall(timestamp1)
	}
	for i := 0; i < expectedDay; i++ {
		timestamp1 := time.Now()
		c.AddCall(timestamp1)
	}

	expectedSummary := fmt.Sprintf("Summary of shorturl:\n calls in the last day: %d calls\n calls in the last week: %d calls\n total calls since creation: %d calls\n", expectedDay, expectedDay+expectedWeek, expectedDay+expectedWeek+expectedAll)

	assert.Equal(t, expectedSummary, c.GetSummary())
}

func TestCounterEdge(t *testing.T) {
	c := NewCounter()

	// should not count as today, but count as this week
	exactlyOneDay := time.Now().Add(-24*time.Hour)
	// should not count as this week, but counts in total
	exactlyOneWeek := time.Now().Add(-7*24*time.Hour)

	c.AddCall(exactlyOneDay)
	c.AddCall(exactlyOneWeek)
	expectedSummary := fmt.Sprintf("Summary of shorturl:\n calls in the last day: %d calls\n calls in the last week: %d calls\n total calls since creation: %d calls\n", 0, 1, 2)

	assert.Equal(t, expectedSummary, c.GetSummary())
}