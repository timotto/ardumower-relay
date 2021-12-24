package main

import (
	"fmt"
	"github.com/timotto/ardumower-relay/test/fake_app"
	"math"
	"net/http"
	"time"
)

type rttMeasurementRun struct {
	app *fake_app.FakeApp

	start   time.Time
	stop    time.Time
	timeout time.Time
	limit   int

	Results  []time.Duration
	Errors   []error
	Duration time.Duration

	AvgRtt time.Duration
	MinRtt time.Duration
	MaxRtt time.Duration

	ErrorRate float64
}

func runRttMeasurement(app *fake_app.FakeApp, maxDuration time.Duration, maxIterations int) *rttMeasurementRun {
	now := time.Now()
	r := &rttMeasurementRun{
		app:     app,
		start:   now,
		timeout: now.Add(maxDuration),
		limit:   maxIterations,
	}

	r.run()
	r.stats()

	return r
}

func (r *rttMeasurementRun) Add(others ...*rttMeasurementRun) {
	for _, o := range others {
		r.Results = append(r.Results, o.Results...)
		r.Errors = append(r.Errors, o.Errors...)
	}
	r.stats()
}

func (r *rttMeasurementRun) run() {
	for time.Now().Before(r.timeout) && (r.limit == 0 || r.resultCount() < r.limit) {
		if result, err := r.measure(); err != nil {
			r.Errors = append(r.Errors, err)
		} else {
			r.Results = append(r.Results, result)
		}
	}
	r.stop = time.Now()
}

func (r *rttMeasurementRun) stats() {
	r.Duration = r.stop.Sub(r.start)

	if r.resultCount() == 0 {
		r.ErrorRate = 1
		return
	}

	r.ErrorRate = float64(len(r.Errors)) / float64(r.resultCount())

	if len(r.Results) == 0 {
		return
	}

	var (
		avg float64 = 0
		min         = math.MaxFloat64
		max         = -math.MaxFloat64
	)

	for _, result := range r.Results {
		seconds := result.Seconds()
		avg += seconds
		if seconds < min {
			min = seconds
		}
		if seconds > max {
			max = seconds
		}
	}
	avg /= float64(len(r.Results))

	r.AvgRtt = fromSeconds(avg)
	r.MinRtt = fromSeconds(min)
	r.MaxRtt = fromSeconds(max)
}

func (r *rttMeasurementRun) resultCount() int {
	return len(r.Results) + len(r.Errors)
}

func (r *rttMeasurementRun) measure() (time.Duration, error) {
	t0 := time.Now()
	res, err := r.app.Send("AT+1\n")
	if err != nil {
		return 0, err
	} else if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("response status is %v", res.Status)
	}

	return time.Now().Sub(t0), nil
}

func fromSeconds(seconds float64) time.Duration {
	return time.Duration(float64(time.Second) * seconds)
}
