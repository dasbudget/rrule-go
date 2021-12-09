package rrule

import "time"

// Next is a generator of time.Time.
// It returns false of Ok if there is no value to generate.
type Next func() (value time.Time, ok bool)


// Previous is a generator of time.Time.
// It returns false of Ok if there is no value to generate.
type Previous func() (value time.Time, ok bool)

type Iterator interface {
	Next() (value time.Time, ok bool)
	Previous() (value time.Time, ok bool)
}
