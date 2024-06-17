package referrals

import (
	"testing"
	"time"
)

func TestGetRunnerInterval(t *testing.T) {
	var testCases = []struct {
		regularInterval time.Duration
		lastRunTime     time.Time
		expected        time.Duration
	}{
		{time.Duration(60) * time.Second, time.Now().UTC(), 60},
		{time.Duration(60) * time.Second, time.Now().UTC().Add(time.Duration(-10) * time.Second), 50},
		{time.Duration(60) * time.Second, time.Now().UTC().Add(time.Duration(-60) * time.Second), 0},
		{time.Duration(60) * time.Second, time.Now().UTC().Add(time.Duration(-61) * time.Second), 0},
	}

	for i, test := range testCases {
		expected := test.expected
		calculated := getRunnerInterval(test.regularInterval, test.lastRunTime)

		if expected != calculated {
			t.Errorf("loop = %d, for regularInterval = %d lastRunTime = %s, expected %d. got %d.",
				i, test.regularInterval, test.lastRunTime, expected, calculated)
		}
	}
}
