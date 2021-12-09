package rrule

import (
	"fmt"
	"math"
	"strings"
	"time"
)

// port of https://github.com/jakubroztocil/rrule/blob/ab9c564a83de2f9688d6671f2a6df273ceb902bf/src/nlp/totext.ts

var supportedOptions = map[Frequency][]string{
	MINUTELY: {"count", "until", "interval", "byweekday", "bymonthday", "bymonth"},
	HOURLY:   {"count", "until", "interval", "byweekday", "bymonthday", "bymonth"},
	DAILY:    {"count", "until", "interval", "byweekday", "bymonthday", "bymonth", "byhour"},
	WEEKLY:   {"count", "until", "interval", "byweekday", "bymonthday", "bymonth"},
	MONTHLY:  {"count", "until", "interval", "byweekday", "bymonthday", "bymonth"},
	YEARLY:   {"count", "until", "interval", "byweekday", "bymonthday", "bymonth", "byeweekno", "byyearday"},
}

const (
	// PrettyDateFormat is date format used to print human friendly dates
	PrettyDateFormat = "January 2, 2006"
)

type rruleText struct {
	text  []string
	rrule *RRule
}

func (t *rruleText) String() string {
	if _, ok := supportedOptions[t.rrule.freq]; !ok {
		return "RRule error: Unable to fully convert this rrule to text"
	}

	fmt.Println(t.rrule.Options.String())

	t.text = []string{"Every"}
	switch t.rrule.freq {
	case MINUTELY:
		t.minutely()
	case HOURLY:
		t.hourly()
	case DAILY:
		t.daily()
	case WEEKLY:
		t.weekly()
	case MONTHLY:
		t.monthly()
	case YEARLY:
		t.yearly()
	}

	if t.hasUntil() {
		t.add("until")
		t.add(t.rrule.until.Format(PrettyDateFormat))
	} else if t.rrule.count > 0 {
		t.add(fmt.Sprintf("for %d %s", t.rrule.count, plural(t.rrule.count, "time", false)))
	}

	if !t.isFullyConvertible() {
		t.add("(~ approximate)")
	}

	return strings.Join(t.text, "")
}

func (t *rruleText) add(texts ...string) {
	var toAdd []string
	for _, text := range texts {
		toAdd = append(toAdd, " ", text)
	}

	t.text = append(t.text, toAdd...)
}

func (t *rruleText) isFullyConvertible() bool {
	_, ok := supportedOptions[t.rrule.freq]
	if !ok {
		return false
	}

	if !t.rrule.OrigOptions.Until.IsZero() && t.rrule.OrigOptions.Count > 0 {
		return false
	}

	return true
}

func (t *rruleText) minutely() {
	interval := t.rrule.Options.Interval
	if interval != 1 {
		t.add(t.niceInterval())
	}

	t.add(plural(interval, "minute", true))
}

func (t *rruleText) hourly() {
	interval := t.rrule.Options.Interval
	if interval != 1 {
		t.add(fmt.Sprint(interval))
	}

	t.add(plural(interval, "hour", false))
}

func (t *rruleText) daily() {
	interval := t.rrule.Options.Interval
	if interval != 1 {
		t.add(fmt.Sprint(interval))
	}

	if len(t.rrule.byweekday) > 0 && t.isWeekdays() {
		t.add(plural(interval, "weekday", false))
	} else {
		t.add(plural(interval, "day", false))
	}

	if len(t.rrule.OrigOptions.Bymonth) > 0 {
		t.add("in")
		t._byMonth()
	}

	if len(t.rrule.bymonthday) > 0 {
		t._byMonthDay()
	} else if len(t.rrule.byweekday) > 0 {
		t._byWeekday()
	} else if len(t.rrule.OrigOptions.Byhour) > 0 {
		t._byHour()
	}
}

func (t *rruleText) weekly() {
	interval := t.rrule.interval
	if interval != 1 {
		t.add(t.niceInterval())
	}

	if len(t.rrule.byweekday) > 0 && t.isWeekdays() {
		if interval == 1 {
			t.add(plural(interval, "weekday", false))
		} else {
			t.add("on", "weekdays")
		}
	} else if len(t.rrule.byweekday) > 0 && t.isEveryDay() {
		t.add(plural(interval, "day", false))
	} else {
		if len(t.rrule.OrigOptions.Bymonth) > 0 {
			t.add("in")
			t._byMonth()
		}

		if len(t.rrule.bymonthday) > 0 {
			t._byMonthDay()
		} else if len(t.rrule.byweekday) > 0 {
			t._byWeekday()
		}
	}
}

func (t *rruleText) monthly() {

}

func (t *rruleText) yearly() {

}

func plural(count int, word string, useOther bool) string {
	if count == 1 || (count == 2 && useOther){
		return word
	}

	return word + "s"
}

func (t *rruleText) _byMonthDay() {
	if t.isAllWeeks() {
		t.add("on")
		t.add(
			t.list(t.rrule.byweekday, func(idx int) string {
				return weekdayText(Weekday{
					weekday: t.rrule.byweekday[idx],
					n:       0,
				})
			}, "or"),
		)
		t.add("the")
		t.add(t.list(t.rrule.bymonthday, nthText, "or"))
	} else {
		t.add("on the")
		t.add(t.list(t.rrule.bymonthday, nthText, "and"))
	}
}

func (t *rruleText) _byWeekday() {
	if t.isAllWeeks() && !t.isWeekdays() {
		t.add(t.list(t.rrule.byweekday, func(idx int) string {
			return weekdayText(Weekday{
				weekday: t.rrule.byweekday[idx],
				n:       0,
			})
		}, "and"))
	}

	if t.isNWeeks() {
		if t.isAllWeeks() {
			t.add("and")
		}

		t.add("on the", t.list(t.rrule.bynweekday, func(idx int) string {
			return weekdayText(t.rrule.bynweekday[idx])
		}, "and"))
	}
}

func (t *rruleText) isWeekdays() bool {
	return len(t.rrule.byweekday) == 5 &&
		!contains(t.rrule.byweekday, SA.weekday) &&
		!contains(t.rrule.byweekday, SU.weekday)
}

func (t *rruleText) isEveryDay() bool {
	return len(t.rrule.byweekday) == 7
}

func (t *rruleText) isAllWeeks() bool {
	return len(t.rrule.byweekday) > 0
}

func (t *rruleText) isNWeeks() bool {
	return len(t.rrule.bynweekday) > 0
}

func (t *rruleText) someWeeks() []Weekday {
	var weeks []Weekday
	for _, weekday := range t.rrule.bynweekday {
		if weekday.N() != 0 {
			weeks = append(weeks, weekday)
		}
	}

	return weeks
}

func (t *rruleText) list(vals interface{}, text func(idx int) string, s string) string {
	ret := strings.Builder{}

	valsSlice := interfaceSlice(vals)

	if text == nil {
		text = func(idx int) string {
			return fmt.Sprint(valsSlice[idx])
		}
	}

	for i := range valsSlice {
		if i != 0 {
			if i == len(valsSlice)-1 {
				ret.WriteString(" ")
				ret.WriteString(s)
				ret.WriteString(" ")
			} else {
				ret.WriteString(", ")
			}
		}

		ret.WriteString(text(i))
	}

	return ret.String()
}

func (t *rruleText) _byHour() {
	t.add("at")
	t.add(t.list(t.rrule.OrigOptions.Byhour, nil, "and"))
}

func (t *rruleText) _byMonth() {
	t.add(t.list(t.rrule.Options.Bymonth, func(idx int) string {
		return time.Month(t.rrule.Options.Bymonth[idx]).String()
	}, "and"))
}

func (t *rruleText) hasUntil() bool {
	if t.rrule.until.Equal(t.rrule.GetDTStart().Add(maxDuration)) {
		return false
	}

	return !t.rrule.until.IsZero()
}

func (t *rruleText) niceInterval() string {
	if t.rrule.interval == 2 {
		return "other"
	}

	return fmt.Sprint(t.rrule.interval)
}

func nthText(num int) string {
	if num == -1 {
		return "last"
	}

	suffix := "th"
	npos := int(math.Abs(float64(num)))
	switch npos {
	case 1, 21, 31:
		suffix = "st"
	case 2, 22:
		suffix = "nd"
	case 3, 23:
		suffix = "rd"
	default:
		suffix = "th"
	}

	nth := fmt.Sprintf("%d%s", npos, suffix)
	if num < 0 {
		return nth + " last"
	}

	return nth
}

func monthText(num int) string {
	return time.Month(num).String()
}

func weekdayText(weekday Weekday) string {
	nth := ""
	if weekday.N() > 0 {
		nth = nthText(weekday.N()) + " "
	}

	return fmt.Sprintf("%s%s", nth, weekday.ToWeekday())
}
