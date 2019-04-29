// Copyright 2019 Tobias Hintze
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// github.com/thz/suntrigger
// A tool to trigger actions based on the sun's position.

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/KlausBrunner/gosolarpos"
)

func zenith(when time.Time) float64 {
	deltaT := gosolarpos.EstimateDeltaT(when)
	_, zenithAngle := gosolarpos.Grena3(when,
		*flagLatitude,  // latitude (degrees)
		*flagLongitude, // longitude (degrees)
		deltaT,
		1000, // air pressure (hPa)
		20)   // air temperature (degrees centigrade)
	return zenithAngle
}

var (
	flagSimulate = flag.Bool("simulate", false, "let time pass unnatural quickly")

	flagLatitude  = flag.Float64("latitude", 0.0, "latitude (degrees) of observation location")
	flagLongitude = flag.Float64("longitude", 0.0, "longitude (degrees) of observation location")

	flagSunsetAction              = flag.String("sunset-action", "", "action to trigger on sunset")
	flagSunsetDegrees             = flag.String("sunset-degrees", "90", "degrees between zentih and sun to define sunset")
	flagSunriseAction             = flag.String("sunrise-action", "", "action to trigger on sunrise")
	flagSunriseDegrees            = flag.String("sunrise-degrees", "90", "degrees between zentih and sun to define sunrise")
	sunsetDegrees, sunriseDegrees float64

	lastReading float64
)

func nextReading(when time.Time) error {
	const timeFormat = time.RFC3339
	reading := zenith(when)
	defer func() { lastReading = reading }()

	// check for sunset
	if *flagSunsetAction != "" {
		if lastReading != 0.0 && lastReading < sunsetDegrees && reading >= sunsetDegrees {
			fmt.Printf("%s Sunset at %.2f degrees. Triggering action.\n", when.Format(timeFormat), reading)
			return executeTrigger(*flagSunsetAction)
		}
	}

	// check for sunrise
	if *flagSunriseAction != "" {
		if lastReading != 0.0 && lastReading > sunriseDegrees && reading <= sunriseDegrees {
			fmt.Printf("%s Sunrise at %.2f degrees. Triggering action.\n", when.Format(timeFormat), reading)
			return executeTrigger(*flagSunriseAction)
		}
	}

	if reading < lastReading {
		fmt.Printf("%s The sun is rising: %.4f --> %.4f\n", when.Format(timeFormat), reading, sunriseDegrees)
	} else if reading > lastReading {
		fmt.Printf("%s The sun is setting: %.4f --> %.4f\n", when.Format(timeFormat), reading, sunsetDegrees)
	} else {
		fmt.Printf("%s The sun is standing still: at %.4f degrees.\n", when.Format(timeFormat), reading)
	}
	return nil
}

func executeTrigger(cmdstring string) error {
	cmd := exec.Command("sh", "-c", cmdstring)
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	fmt.Printf("%sEOF\n", stdoutStderr)
	return nil
}

func parseFlags() error {
	flag.Parse()
	var (
		f   float64
		err error
	)

	if *flagSunsetDegrees != "" {
		f, err = strconv.ParseFloat(*flagSunsetDegrees, 64)
		if err != nil {
			return err
		}
		sunsetDegrees = f
	}

	if *flagSunriseDegrees != "" {
		f, err = strconv.ParseFloat(*flagSunriseDegrees, 64)
		if err != nil {
			return err
		}
		sunriseDegrees = f
	}

	if *flagLatitude == 0.0 || *flagLongitude == 0.0 {
		return fmt.Errorf("observation point must be specified with latitude and longitude")
		// Actual latitudes and longitudes of 0.0 are not supported. Sorry.
	}

	return nil
}

func simulate(ticker *time.Ticker, done chan bool) {
	defer ticker.Stop()
	when := time.Now()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			when = when.Add(time.Minute * 15)
			nextReading(when)
		}
	}
}

func loop(ticker *time.Ticker, done chan bool) {
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case now := <-ticker.C:
			nextReading(now)
		}
	}
}

func main() {

	if err := parseFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid flags: %v\n", err)
		os.Exit(1)
	}

	done := make(chan bool) // currently no one is ever stopping us
	if *flagSimulate {
		simulate(time.NewTicker(1*time.Second), done)
	} else {
		loop(time.NewTicker(60*time.Second), done)
	}
}
