package rrule

import (
	"fmt"
	"testing"
	"time"
)

func runText(t *testing.T, text, expected string) {
	t.Helper()
	rrule, err := StrToRRule(text)
	if err != nil {
		t.Fatalf("StrToRRuleSet(%s): error = %s", text, err.Error())
	}

	if got := (&rruleText{RRule: *rrule}).String(); got != expected {
		t.Errorf("String() = '%v', want '%v'", got, expected)
	}
}

func Test_rruleText_SECONDLY(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		// MINUTELY
		{expected: "Every second", text: "FREQ=SECONDLY"},
		{expected: "Every other second", text: "FREQ=SECONDLY;INTERVAL=2"},
		{expected: "Every 3 seconds", text: "FREQ=SECONDLY;INTERVAL=3"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			runText(t, tt.text, tt.expected)
		})
	}
}

func Test_rruleText_MINUTELY(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		// MINUTELY
		{expected: "Every minute", text: "FREQ=MINUTELY"},
		{expected: "Every other minute", text: "FREQ=MINUTELY;INTERVAL=2"},
		{expected: "Every 3 minutes", text: "FREQ=MINUTELY;INTERVAL=3"},
		{expected: "Every 3 minutes at seconds 15 and 30", text: "FREQ=MINUTELY;INTERVAL=3;BYSECOND=15,30"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			runText(t, tt.text, tt.expected)
		})
	}
}

func Test_rruleText_HOURLY(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{expected: "Every hour", text: "FREQ=HOURLY"},
		{expected: "Every other hour", text: "FREQ=HOURLY;INTERVAL=2"},
		{expected: "Every 6 hours", text: "FREQ=HOURLY;INTERVAL=6"},
		{expected: "Every 4 hours", text: "INTERVAL=4;FREQ=HOURLY"},
		{expected: "Every 4 hours at minute 50", text: "INTERVAL=4;FREQ=HOURLY;BYMINUTE=50"},
		{expected: "Every 4 hours at minutes 15 and 23", text: "INTERVAL=4;FREQ=HOURLY;BYMINUTE=15,23"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			runText(t, tt.text, tt.expected)
		})
	}
}

func Test_rruleText_DAILY(t *testing.T) {
	tests := []struct {
		text     string
		expected string
	}{
		{expected: "Every day", text: "FREQ=DAILY"},
		{expected: "Every other day", text: "FREQ=DAILY;INTERVAL=2"},
		{expected: "Every 365 days", text: "FREQ=DAILY;INTERVAL=365"},
		{expected: "Every day at hour 10", text: "FREQ=DAILY;BYHOUR=10"},
		{expected: "Every day at hours 10, 12 and 17", text: "FREQ=DAILY;BYHOUR=10,12,17"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			runText(t, tt.text, tt.expected)
		})
	}
}

func Test_rruleText_WEEKLY(t *testing.T) {
	now := time.Now().UTC()
	weekday := now.Weekday()

	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{name: "Every %s", expected: fmt.Sprintf("Every %s", now.Weekday()), text: "FREQ=WEEKLY"},
		{name: "Every %s", expected: fmt.Sprintf("Every %s", weekday), text: "FREQ=WEEKLY;UNTIL=20771110T000000Z"},
		{name: "Every day", text: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU"},
		{name: "Every weekday", text: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"},
		{name: "Every Tuesday", text: "FREQ=WEEKLY;BYDAY=TU"},
		{name: "Every Monday and Wednesday", text: "FREQ=WEEKLY;BYDAY=MO,WE"},
		{name: "Every other %s", expected: fmt.Sprintf("Every other %s", weekday), text: "INTERVAL=2;FREQ=WEEKLY"},
		{name: "Every %s", expected: fmt.Sprintf("Every %s", weekday), text: "FREQ=WEEKLY;UNTIL=20660101T080000Z"},
		{name: "Every %s", expected: fmt.Sprintf("Every %s", weekday), text: "FREQ=WEEKLY;COUNT=20"},
		{name: "Every Monday", text: "FREQ=WEEKLY;BYDAY=MO"},
		{name: "The 3rd, 10th, 17th and 24th of every month", text: "FREQ=WEEKLY;BYMONTHDAY=3,10,17,24"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := tt.expected
			if tt.expected == "" {
				exp = tt.name
			}

			runText(t, tt.text, exp)
		})
	}
}

func Test_rruleText_MONTHLY(t *testing.T) {
	now := time.Now().UTC()
	month := now.Month()
	dayth := rruleText{}.stringSelect(recurStrings["nth_monthday"], now.Day(), "%{n}", now.Day())

	tests := []struct {
		name string
		text     string
		expected string
	}{
		{name:"The %s of every month", expected: fmt.Sprintf("The %s of every month", dayth), text: "FREQ=MONTHLY"},
		{name:"Every 6 months starting %s %s", expected: fmt.Sprintf("Every 6 months starting %s %s", month, dayth), text: "INTERVAL=6;FREQ=MONTHLY"},
		{name:"The 4th of every month", text: "FREQ=MONTHLY;BYMONTHDAY=4"},
		{name:"The 4th to last day of every month", text: "FREQ=MONTHLY;BYMONTHDAY=-4"},
		{name:"The 8th to last day of every month", text: "FREQ=MONTHLY;BYMONTHDAY=-8"},
		{name:"The 15th and last day of every month", text: "FREQ=MONTHLY;BYMONTHDAY=15,-1"},
		{name:"The 7th, 20th and 2nd to last day of every month", text: "FREQ=MONTHLY;BYMONTHDAY=7,20,-2"},
		{name:"The 3rd Tuesday of every month", text: "FREQ=MONTHLY;BYDAY=+3TU"},
		{name:"The 2nd Monday and 3rd Tuesday of every month", text: "FREQ=MONTHLY;BYDAY=+2MO,+3TU"},
		{name:"The 3rd to last Tuesday of every month", text: "FREQ=MONTHLY;BYDAY=-3TU"},
		{name:"The last Monday of every month", text: "FREQ=MONTHLY;BYDAY=-1MO"},
		{name:"The 2nd to last Friday of every month", text: "FREQ=MONTHLY;BYDAY=-2FR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := tt.expected
			if tt.expected == "" {
				exp = tt.name
			}

			runText(t, tt.text, exp)
		})
	}
}

func Test_rruleText_YEARLY(t *testing.T) {
	now := time.Now().UTC()
	month := now.Month()
	dayth := rruleText{}.stringSelect(recurStrings["nth_monthday"], now.Day(), "%{n}", now.Day())

	tests := []struct {
		text     string
		expected string
	}{
		{expected: fmt.Sprintf("Every %s %s", month, dayth), text: "FREQ=YEARLY"},
		{expected: "The 1st Friday of every year", text: "FREQ=YEARLY;BYDAY=+1FR"},
		{expected: "The 13th Friday of every year", text: "FREQ=YEARLY;BYDAY=+13FR"},
		{expected: "The 3rd day of every year", text: "FREQ=YEARLY;BYYEARDAY=3"},
		{expected: "The 3rd to last day of every year", text: "FREQ=YEARLY;BYYEARDAY=-3"},
		{expected: "The 25th day of every year", text: "FREQ=YEARLY;BYYEARDAY=25"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			runText(t, tt.text, tt.expected)
		})
	}
}
