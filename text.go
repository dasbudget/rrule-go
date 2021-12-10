package rrule

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode"
)

// port of https://github.com/rlanvin/php-rrule/blob/30f9170f3af4ed2fc84c67482d67bc316c442d2e/src/RRule.php#L2078

var (
	recurStrings = map[interface{}]interface{}{
		SECONDLY: map[interface{}]string{
			1:   "Every second",
			2:   "Every other second",
			nil: "Every %{interval} seconds",
		},
		MINUTELY: map[interface{}]string{
			1:   "Every minute",
			2:   "Every other minute",
			nil: "Every %{interval} minutes",
		},
		HOURLY: map[interface{}]string{
			1:   "Every hour",
			2:   "Every other hour",
			nil: "Every %{interval} hours",
		},
		DAILY: map[interface{}]string{
			1:   "Every day",
			2:   "Every other day",
			nil: "Every %{interval} days",
		},
		WEEKLY: map[interface{}]string{
			1:   "Every %{weekdays}",
			2:   "Every other %{weekdays}",
			nil: "Every %{interval} %{weekdays}",
		},
		MONTHLY: map[interface{}]string{
			1:   "the %{x} of every month",
			2:   "the %{x} of every other month",
			nil: "Every %{interval} months starting %{month} %{nth}",
		},
		YEARLY: map[interface{}]string{
			1:   "Every %{month} %{nth}",
			2:   "Every other %{month} %{nth}",
			nil: "Every %{interval} years starting %{month} %{nth}",
		},
		"dtstart":  ", starting from %{date}",
		"infinite": ", forever",
		"until":    ", until %{date}",
		"count": map[interface{}]string{
			1:   ", one time",
			nil: ", %{count} times",
		},
		"and": "and ",
		"x_of_the_y": map[Frequency]string{
			YEARLY:  "the %{x} of every year", // e.g. the first Monday of the year, or the first day of the year
			MONTHLY: "the %{x} of every month",
		},
		"bymonth":   " in %{months}",
		"byweekday": " %{weekdays}",
		"nth_weekday": map[interface{}]string{
			1:   "1st %{weekday}", // e.g. the first Monday
			2:   "2nd %{weekday}",
			3:   "3rd %{weekday}",
			nil: "%{n}th %{weekday}",
		},
		"-nth_weekday": map[interface{}]string{
			-1:  "last %{weekday}", // e.g. the last Monday
			-2:  "2nd to last %{weekday}",
			-3:  "3rd to last %{weekday}",
			nil: "%{n}th to the last %{weekday}",
		},
		"byweekno": map[interface{}]string{
			1:   " on week %{weeks}",
			nil: " on weeks number %{weeks}",
		},
		"bymonthday": " on %{monthdays}",
		"nth_monthday": map[interface{}]string{
			1:   "1st",
			2:   "2nd",
			3:   "3rd",
			21:  "21st",
			22:  "22nd",
			23:  "23rd",
			31:  "31st",
			nil: "%{n}th",
		},
		"-nth_monthday": map[interface{}]string{
			-1:  "last day",
			-2:  "2nd to last day",
			-3:  "3rd to last day",
			-21: "21st to last day",
			-22: "22nd to last day",
			-23: "23rd to last day",
			-31: "31st to last day",
			nil: "%{n}th to last day",
		},
		"byyearday": map[interface{}]string{
			1:   "%{yeardays} day",
			nil: "%{yeardays} days",
		},
		"nth_yearday": map[interface{}]string{
			1:   "1st",
			2:   "2nd",
			3:   "3rd",
			nil: "%{n}th",
		},
		"-nth_yearday": map[interface{}]string{
			-1:  "last",
			-2:  "2nd to last",
			-3:  "3rd to last",
			nil: "%{n}th to last",
		},
		"byhour": map[interface{}]string{
			1:   " at hour %{hours}",
			nil: " at hours %{hours}",
		},
		"nth_hour": "%{n}",
		"byminute": map[interface{}]string{
			1:   " at minute %{minutes}",
			nil: " at minutes %{minutes}",
		},
		"nth_minute": "%{n}",
		"bysecond": map[interface{}]string{
			1:   " at second %{seconds}",
			nil: " at seconds %{seconds}",
		},
		"nth_second": "%{n}",
		"bysetpos":   ", but only %{setpos} instance of this set",
		"nth_setpos": map[interface{}]string{
			1:   "the first",
			2:   "the second",
			3:   "the third",
			nil: "the %{n}th",
		},
		"-nth_setpos": map[interface{}]string{
			-1:  "the last",
			-2:  "the penultimate",
			-3:  "the antepenultimate",
			nil: "the %{n}th to the last",
		},
	}

	order = []string{"freq",
		"byweekday",
		"bymonth",
		"byweekno",
		"byyearday",
		"bymonthday",
		"byhour",
		"byminute",
		"bysecond",
		"bysetpos",
	}
)

type rruleText struct {
	RRule
	parts map[string]string
}

func (r *rruleText) String() string {
	r.parts = map[string]string{}
	for _, s := range order {
		r.parts[s] = ""
	}

	switch r.freq {
	case SECONDLY, MINUTELY, HOURLY, DAILY:
		r.parts["freq"] = r.selectString(recurStrings[r.freq], r.interval, "%{interval}", r.interval)
	}

	r.daily()
	r.weekly()
	r.monthly()
	r.yearly()

	// shared
	r._byHour()
	r._byMinute()
	r._bySecond()

	b := strings.Builder{}
	for _, str := range order {
		b.WriteString(r.parts[str])
	}

	s := []rune(b.String())
	if len(s) > 0 {
		s[0] = unicode.ToUpper(s[0])
	}
	return string(s)
}

func (r *rruleText) daily() {
	if r.freq != DAILY {
		return
	}

	r._byHour()
}

func (r *rruleText) weekly() {
	if r.freq != WEEKLY {
		return
	}

	r._byWeekday()
	r._byMonthDay()
}

func (r *rruleText) monthly() {
	if r.freq != MONTHLY {
		return
	}

	if len(r.OrigOptions.Byweekday) > 0 {
		r._byWeekday()
	} else {
		r._byMonthDay()
	}
}

func (r *rruleText) yearly() {
	if r.freq != YEARLY {
		return
	}

	if len(r.OrigOptions.Bymonth) > 0 {
		r._byMonth()
	} else if len(r.OrigOptions.Byyearday) > 0 {
		r._byYearDay()
	} else if len(r.OrigOptions.Byweekno) > 0 {
		r._byWeekno()
	} else if len(r.OrigOptions.Byweekday) > 0 {
		r._byWeekday()
	} else {
		r.parts["bymonthday"] = r.selectString(
			recurStrings[YEARLY],
			r.interval,
			"%{interval}", r.interval,
			"%{month}", r.dtstart.Month(),
			"%{nth}", r.selectString(
				recurStrings["nth_monthday"],
				r.dtstart.Day(),
				"%{n}", int(math.Abs(float64(r.dtstart.Day()))),
			),
		)
	}
}

func (r *rruleText) _byHour() {
	if len(r.OrigOptions.Byhour) == 0 {
		return
	}

	tmp := make([]interface{}, len(r.OrigOptions.Byhour))
	for i, hr := range r.OrigOptions.Byhour {
		tmp[i] = r.selectString(recurStrings["nth_hour"], hr, "%{n}", hr)
	}

	r.parts["byhour"] = r.selectString(
		recurStrings["byhour"],
		len(tmp),
		"%{hours}", r.join(tmp),
	)
}

func (r *rruleText) _bySecond() {
	if len(r.OrigOptions.Bysecond) == 0 {
		return
	}

	tmp := make([]interface{}, len(r.OrigOptions.Bysecond))
	for i, sec := range r.OrigOptions.Bysecond {
		tmp[i] = r.selectString(recurStrings["nth_second"], sec, "%{n}", sec)
	}

	r.parts["bysecond"] = r.selectString(
		recurStrings["bysecond"],
		len(tmp),
		"%{seconds}", r.join(tmp),
	)
}

func (r *rruleText) _byMinute() {
	if len(r.OrigOptions.Byminute) == 0 {
		return
	}

	tmp := make([]interface{}, len(r.OrigOptions.Byminute))
	for i, sec := range r.OrigOptions.Byminute {
		tmp[i] = r.selectString(recurStrings["nth_minute"], sec, "%{n}", sec)
	}

	r.parts["byminute"] = r.selectString(
		recurStrings["byminute"],
		len(tmp),
		"%{minutes}", r.join(tmp),
	)
}

func (r *rruleText) _byWeekday() {
	if len(r.OrigOptions.Byweekday) == 0 && r.freq != WEEKLY {
		return
	}

	if len(r.byweekday) > 0 {
		var tmp []interface{}
		if r.isWeekdays() {
			tmp = []interface{}{"weekday"}
		} else if r.isEveryDay() {
			tmp = []interface{}{"day"}
		} else {
			tmp = make([]interface{}, len(r.byweekday))
			for i, day := range r.byweekday {
				tmp[i] = (&Weekday{weekday: day}).ToWeekday()
			}
		}

		r.parts["byweekday"] = r.selectString(
			recurStrings[WEEKLY],
			r.interval,
			"%{weekdays}", r.join(tmp),
		)
	}

	if len(r.bynweekday) > 0 {
		tmp := make([]interface{}, len(r.bynweekday))
		for i, day := range r.bynweekday {
			selection := "nth_weekday"
			if day.N() < 0 {
				selection = "-nth_weekday"
			}

			tmp[i] = r.selectString(
				recurStrings[selection],
				day.N(),
				"%{n}", int(math.Abs(float64(day.N()))),
				"%{weekday}", day.ToWeekday(),
			)
		}

		r.parts["bymonthday"] = r.selectString(
			recurStrings["x_of_the_y"],
			r.freq,
			"%{x}", r.join(tmp),
			"%{weekdays}", r.join(tmp),
			"%{interval}", r.interval,
			"%{day}", r.dtstart.Day(),
		)
	}
}

func (r *rruleText) _byMonthDay() {
	if len(r.OrigOptions.Bymonthday) == 0 && r.freq != MONTHLY {
		return
	}

	monthdays := append(r.bymonthday, r.bynmonthday...)
	if len(monthdays) > 0 {
		tmp := make([]interface{}, len(monthdays))
		for i, day := range monthdays {
			selection := "nth_monthday"
			if day < 0 {
				selection = "-nth_monthday"
			}

			tmp[i] = r.selectString(
				recurStrings[selection],
				day,
				"%{n}", int(math.Abs(float64(day))),
			)
		}

		r.parts["bymonthday"] = r.selectString(
			recurStrings[MONTHLY],
			r.interval,
			"%{x}", r.join(tmp),
			"%{interval}", r.interval,
			"%{month}", r.dtstart.Month(),
			"%{nth}", r.selectString(
				recurStrings["nth_monthday"],
				r.dtstart.Day(),
				"%{n}", int(math.Abs(float64(r.dtstart.Day()))),
			),
		)
	} else {
		nth := r.selectString(recurStrings["nth_monthday"], r.dtstart.Day(), "%{n}", r.dtstart.Day())
		r.parts["freq"] = r.selectString(recurStrings[r.freq], r.interval,
			"%{interval}", fmt.Sprint(r.interval),
			"%{month}", r.dtstart.Month().String(),
			"%{x}", nth,
			"%{nth}", nth,
		)
	}
}

func (r *rruleText) _byMonth() {
	if len(r.bymonth) == 0 {
		return
	}

	tmp := make([]interface{}, len(r.bymonth))
	for i, mo := range r.bymonth {
		tmp[i] = time.Month(mo)
	}

	r.parts["bymonth"] = r.selectString(
		recurStrings["bymonth"],
		nil,
		"%{months}", r.join(tmp),
	)
}

func (r *rruleText) _byYearDay() {
	if len(r.byyearday) == 0 {
		return
	}

	tmp := make([]interface{}, len(r.byyearday))
	for i, yd := range r.byyearday {
		selection := "nth_yearday"
		if yd < 0 {
			selection = "-nth_yearday"
		}

		tmp[i] = r.selectString(
			recurStrings[selection],
			yd,
			"%{n}", int(math.Abs(float64(yd))),
		)
	}

	r.parts["byyearday"] = r.selectString(
		recurStrings["x_of_the_y"],
		YEARLY,
		"%{x}", r.selectString(
			recurStrings["byyearday"],
			len(tmp),
			"%{yeardays}", r.join(tmp),
		),
	)
}

func (r *rruleText) _byWeekno() {
	if len(r.OrigOptions.Byweekno) == 0 {
		return
	}

	tmp := make([]interface{}, len(r.OrigOptions.Byweekno))
	for i, mo := range r.OrigOptions.Byweekno {
		tmp[i] = mo
	}

	r.parts["byweekno"] = r.selectString(
		recurStrings["byweekno"],
		len(tmp),
		"%{weeks}", r.join(tmp),
	)
}

func (r *rruleText) isWeekdays() bool {
	return len(r.byweekday) == 5 &&
		!contains(r.byweekday, SA.weekday) &&
		!contains(r.byweekday, SU.weekday)
}

func (r *rruleText) isEveryDay() bool {
	return len(r.byweekday) == 7
}

//func humanFriendlyText(r *RRule) string {
//	// Every (INTERVAL) FREQ...
//
//	// BYXXX rules
//
//	// todo BYYEARDAY
//
//

//
//	// todo BYMINUTE
//	// todo BYSECOND
//	// todo bysetpos
//

//}

func (r rruleText) selectString(vals interface{}, key interface{}, oldnew ...interface{}) string {
	format := ""

	switch v := vals.(type) {
	case string:
		format = v
	case map[interface{}]string:
		if key == nil {
			format = v[nil]
		} else {
			if s, ok := v[key.(interface{})]; ok {
				format = s
			} else {
				format = v[nil]
			}
		}
	case map[Frequency]string:
		format = v[key.(Frequency)]
	}

	kvs := make([]string, len(oldnew))
	for i, s := range oldnew {
		kvs[i] = fmt.Sprint(s)
	}
	return strings.NewReplacer(kvs...).Replace(format)
}

func (r rruleText) join(vals []interface{}) string {
	if len(vals) == 1 {
		return fmt.Sprint(vals[0])
	}

	conjunct := "and"

	total := len(vals)
	last := vals[total-1]
	s := strings.Builder{}

	for i, val := range vals[:len(vals)-1] {
		s.WriteString(fmt.Sprint(val))
		if i == total-2 {
			s.WriteString(" ")
		} else {
			s.WriteString(", ")
		}
	}

	s.WriteString(conjunct)
	s.WriteString(" ")
	s.WriteString(fmt.Sprint(last))

	return s.String()
}
