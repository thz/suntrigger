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
	"strconv"
	"time"

	"github.com/KlausBrunner/gosolarpos"
)

func azimuth(when time.Time) (a float64) {
	a, _ = azimuthZenithDegrees(when)
	return
}

func zenith(when time.Time) (z float64) {
	_, z = azimuthZenithDegrees(when)
	return
}

func azimuthZenithDegrees(when time.Time) (a float64, z float64) {
	deltaT := gosolarpos.EstimateDeltaT(when)
	a, z = gosolarpos.Grena3(when,
		*flagLatitude,  // latitude (degrees)
		*flagLongitude, // longitude (degrees)
		deltaT,
		1000, // air pressure (hPa)
		20)   // air temperature (degrees centigrade)
	return
}

var (
	flagShowOnly = flag.Bool("show", false, "just calculate and output sun position")
	flagSimulate = flag.Bool("simulate", false, "let time pass unnatural quickly")

	flagLatitude  = flag.Float64("latitude", 0.0, "latitude (degrees) of observation location")
	flagLongitude = flag.Float64("longitude", 0.0, "longitude (degrees) of observation location")

	flagSunsetAction   = flag.String("sunset-action", "", "action to trigger on sunset")
	flagSunsetDegrees  = flag.String("sunset-degrees", "90", "degrees between zenith and sun to define sunset")
	flagSunriseAction  = flag.String("sunrise-action", "", "action to trigger on sunrise")
	flagSunriseDegrees = flag.String("sunrise-degrees", "90", "degrees between zenith and sun to define sunrise")

	sunsetDegrees, sunriseDegrees float64

	lastAzimuth, lastZenith float64
	firstReading            = true
)

const timeFormat = time.RFC3339

func nextReading(when time.Time) error {
	currentAzimuth, currentZenith := azimuthZenithDegrees(when)
	defer func() {
		lastZenith = currentZenith
		lastAzimuth = currentAzimuth
		firstReading = false
	}()

	for i, t := range triggers {
		if firstReading {
			fmt.Printf("Active trigger %d: %s\n", i, t)
		} else if t.Kind == "S" { // sunset
			if lastZenith < t.Degrees && currentZenith >= t.Degrees {
				fmt.Printf("%s Sunset at %.4f (trigger: %.4f) degrees. Trigger %d firing.\n",
					when.Format(timeFormat), currentZenith, t.Degrees, i)
				if err := t.execute(); err != nil {
					fmt.Printf("Trigger %d (%s:%.4f) executed with error: %s\n", i, t.Kind, t.Degrees, err)
				}
			}
		} else if t.Kind == "R" { // sunrise
			if lastZenith > t.Degrees && currentZenith <= t.Degrees {
				fmt.Printf("%s Sunrise at %.4f (trigger: %.4f) degrees. Trigger %d firing.\n",
					when.Format(timeFormat), currentZenith, t.Degrees, i)
				if err := t.execute(); err != nil {
					fmt.Printf("Trigger %d (%s:%.4f) executed with error: %s\n", i, t.Kind, t.Degrees, err)
				}
			}
		} else if t.Kind == "A" { // Azimuth
			if (lastAzimuth < t.Degrees && currentAzimuth >= t.Degrees) ||
				(lastAzimuth > currentAzimuth && // 360 deg rollover
					(currentAzimuth >= t.Degrees || lastAzimuth < t.Degrees)) {
				fmt.Printf("%s Sun passed azimuth of %.4f (trigger: %.4f) degrees (Zenith angle at %.4f degrees). Trigger %d firing.\n",
					when.Format(timeFormat), currentAzimuth, t.Degrees, currentZenith, i)
				if err := t.execute(); err != nil {
					fmt.Printf("Trigger %d (%s:%.4f) executed with error: %s\n", i, t.Kind, t.Degrees, err)
				}
			}
		}
	}

	if firstReading {
		fmt.Printf("%s first reading: zenith angle: %.4f. Azimuth: %.4f\n", when.Format(timeFormat), currentZenith, currentAzimuth)
	} else if currentZenith < lastZenith {
		fmt.Printf("%s The sun is rising: %.4f. Azimuth: %.4f\n", when.Format(timeFormat), currentZenith, currentAzimuth)
	} else if currentZenith > lastZenith {
		fmt.Printf("%s The sun is setting: %.4f Azimuth: %.4f\n", when.Format(timeFormat), currentZenith, currentAzimuth)
	} else {
		fmt.Printf("%s The sun is standing still: at %.4f degrees / Azimuth: %.4f.\n", when.Format(timeFormat), currentAzimuth)
	}
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

	// old style triggers
	if *flagSunriseAction != "" {
		triggers = append(triggers, trigger{Kind: "R", Degrees: sunriseDegrees, Action: *flagSunriseAction})
	}
	if *flagSunsetAction != "" {
		triggers = append(triggers, trigger{Kind: "S", Degrees: sunsetDegrees, Action: *flagSunsetAction})
	}

	for _, arg := range flag.Args() {
		t := ParseTrigger(arg)
		if t.Kind == "" {
			return fmt.Errorf("invalid remaining argument (not a trigger): %s", arg)
		}
		triggers = append(triggers, t)
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

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "\nUsage of %s [flags] [generic triggers]:\n", os.Args[0])

		flag.PrintDefaults()
		fmt.Fprintf(flag.CommandLine.Output(), `
The -sunXXX-action and -sunXXX-degrees flags are the older and less generic
way of specifying triggers based on the sun's angle to zenith. The generic
trigger syntax ([generic triggers]) can be given multiple times on the
command line and allows triggers to be defined based on the sun's angle to
the zenith (rising or setting) and based on the sun's azimuth. The generic
trigger syntax is: K:D:C where K is one of "S" (for a setting sun
trigger), "R" (for a rising sun trigger), and "A" (for azimuth trigger).
D specifies the degrees of the trigger (angle between zenith and sun or
angle between North and the sun (azimuth)). C specifies a command which is
passed to "sh -c" for execution when the trigger fires.

Examples for generic triggers:

"S:96:echo 'End of twilight. Good night!'"
"R:96:echo 'Beginning of twilight. Good morning!'"
"A:180:echo 'Sun is in the South.'"

The old [-sunrise-action 'echo Morning' -sunrise-degrees 90] can be expressed
by a generic trigger as "R:90:echo Morning".
`)

	}

	if err := parseFlags(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid flags: %v\n", err)
		os.Exit(1)
	}

	done := make(chan bool) // currently no one is ever stopping us
	if *flagShowOnly {
		now := time.Now()
		azimuthDegrees, zenithDegrees := azimuthZenithDegrees(now)
		fmt.Printf("%s Sun located at %.4f / %.4f degrees (azimuth / zenith).\n", now.Format(timeFormat), azimuthDegrees, zenithDegrees)
		for i, t := range triggers {
			fmt.Printf("Trigger %d: %s\n", i, t)
		}
	} else if *flagSimulate {
		simulate(time.NewTicker(1*time.Second), done)
	} else {
		loop(time.NewTicker(60*time.Second), done)
	}
}
