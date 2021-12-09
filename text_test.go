package rrule

import (
	"fmt"
	"testing"
	"time"
)

func Test_rruleText_String(t *testing.T) {
	now := time.Now().UTC()
	weekday := now.Weekday()
	day := now.Day()
	dayth := nthText(day)

	tests := []struct {
		text     string
		expected string
	}{
		{expected: "Every day", text: "FREQ=DAILY"},
		{expected: "Every day at 10, 12 and 17", text: "FREQ=DAILY;BYHOUR=10,12,17"},
		{expected: fmt.Sprintf("Every %s", now.Weekday()), text: "FREQ=WEEKLY"},
		{expected: "Every hour", text: "FREQ=HOURLY"},
		{expected: "Every 4 hours", text: "INTERVAL=4;FREQ=HOURLY"},
		{expected: "Every Tuesday", text: "FREQ=WEEKLY;BYDAY=TU"},
		{expected: "Every Monday, and Wednesday", text: "FREQ=WEEKLY;BYDAY=MO,WE"},
		{expected: "Every weekday", text: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR"},
		{expected: fmt.Sprintf("Every other %s", weekday), text: "INTERVAL=2;FREQ=WEEKLY"},
		{expected: fmt.Sprintf("The %s of every month", dayth), text: "FREQ=MONTHLY"},
		{expected: "Every 6 months", text: "INTERVAL=6;FREQ=MONTHLY"},
		{expected: "Every year", text: "FREQ=YEARLY"},
		{expected: "Every year on the 1st Friday", text: "FREQ=YEARLY;BYDAY=+1FR"},
		{expected: "Every year on the 13th Friday", text: "FREQ=YEARLY;BYDAY=+13FR"},
		{expected: "The 4th of every month", text: "FREQ=MONTHLY;BYMONTHDAY=4"},
		{expected: "The 4th last day of month", text: "FREQ=MONTHLY;BYMONTHDAY=-4"},
		{expected: "The 3rd Tuesday of every month", text: "FREQ=MONTHLY;BYDAY=+3TU"},
		{expected: "The 3rd last Tuesday of every month", text: "FREQ=MONTHLY;BYDAY=-3TU"},
		{expected: "The last Monday of every month", text: "FREQ=MONTHLY;BYDAY=-1MO"},
		{expected: "The 2nd last Friday of every month", text: "FREQ=MONTHLY;BYDAY=-2FR"},
		{expected: fmt.Sprintf("Every %s until January 1, 2066", weekday), text: "FREQ=WEEKLY;UNTIL=20660101T080000Z"},
		{expected: fmt.Sprintf("Every %s for 20 times", weekday), text: "FREQ=WEEKLY;COUNT=20"},
		{expected: "Every Monday", text: "FREQ=WEEKLY;BYDAY=MO"},
		{expected: "Every week on the 3rd, 10th, 17th and 24th", text: "FREQ=WEEKLY;BYMONTHDAY=3,10,17,24"},
		{expected: "Every day", text: "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU"},
		{expected: "Every minute", text: "FREQ=MINUTELY"},
		{expected: "Every other minute", text: "FREQ=MINUTELY;INTERVAL=2"},
		{expected: "Every 3 minutes", text: "FREQ=MINUTELY;INTERVAL=3"},
		{expected: fmt.Sprintf("Every %s until November 10, 2077", weekday), text: "FREQ=WEEKLY;UNTIL=20771110T000000Z"},
	}
	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			rrule, err := StrToRRule(tt.text)
			if err != nil {
				t.Fatalf("StrToRRuleSet(%s): error = %s", tt.text, err.Error())
			}

			if got := (&rruleText{rrule: rrule}).String(); got != tt.expected {
				t.Errorf("String() = '%v', want '%v'", got, tt.expected)
			}
		})
	}
}
