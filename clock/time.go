package clock

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kelvins/sunrisesunset"
)

type Reading struct {
	Time, Sunrise, Sunset       time.Time
	BeforeSunrise, BeforeSunset bool
}

type event struct {
	desc string
	time time.Time
}

const timeFmt = "Mon Jan _2 15:04 MST 2006"
const timeFmtShort = "15:04"

func (e event) String(now time.Time) string {

	var naturalDur float64
	until := e.time.Sub(now).Abs()
	unit := "h"
	if until > 90*time.Minute {
		naturalDur = until.Hours()
	} else {
		naturalDur = until.Minutes()
		unit = "m"
	}

	return fmt.Sprintf("%7s at %6s (%4.1f%s)", e.desc, e.time.Format(timeFmtShort), naturalDur, unit)
}

func (r Reading) Render() string {
	eventRise := event{desc: "Sunrise", time: r.Sunrise}
	eventSet := event{desc: "Sunset", time: r.Sunset}
	night := r.BeforeSunrise
	events := []event{
		eventSet,
		eventRise,
	}
	if night {
		events = []event{
			eventRise,
			eventSet,
		}
	}
	return fmt.Sprintf("%s\n%s\n%s ", r.Time.Format(timeFmt), events[0].String(r.Time), events[1].String(r.Time))
}

type Coords struct {
	Lat, Lon float64
}

func (c Coords) Time(now time.Time) Reading {
	utc, err := strconv.Atoi(now.Format("-0700"))
	if err != nil {
		panic(err)
	}
	utcFloat := float64(utc) / 100.0
	p := sunrisesunset.Parameters{
		Latitude:  c.Lat,
		Longitude: c.Lon,
		UtcOffset: utcFloat,
		Date:      now,
	}
	rise, set, err := p.GetSunriseSunset()
	if err != nil {
		panic(err)
	}
	return Reading{
		Time:          now,
		Sunrise:       rise,
		Sunset:        set,
		BeforeSunrise: now.Before(rise),
		BeforeSunset:  now.Before(set),
	}

}
