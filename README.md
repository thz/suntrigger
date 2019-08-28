## suntrigger - actions triggered by the sun

### What?

A tool to trigger actions based on the sun's position.

### How?

This tool takes the observers location as input and continuously calculates the sun's position (azimuth and angle to zenith). The sunrise and sunset can be user-defined by giving the angle to zenith. This allows to trigger based on the sun dropping below the horizon or start/end of civil/nautic/astronomical dusk or dawn.

### Why?

I control my automated shutters with this.

### Usage:

```
Usage of ./suntrigger [flags] [generic triggers]
  -latitude float
    	latitude (degrees) of observation location
  -longitude float
    	longitude (degrees) of observation location
  -simulate
    	let time pass unnatural quickly
  -sunrise-action string
    	action to trigger on sunrise
  -sunrise-degrees string
    	degrees between zenith and sun to define sunrise (default "90")
  -sunset-action string
    	action to trigger on sunset
  -sunset-degrees string
    	degrees between zentih and sun to define sunset (default "90")

The -sunXXX-action and -sunXXX-degrees flags are the older and less generic
way of specifying triggers based on the sun's angle to zenith. The generic
trigger syntax (`[generic triggers]`) can be given multiple times on the
command line and allows triggers to be defined based on the sun's angle to
the zenith (rising or setting) and based on the sun's azimuth. The generic
trigger syntax is: `K:D:C` where `K` is one of `S` (for a setting sun
trigger), `R` (for a rising sun trigger), and `A` (for azimuth trigger).
`D` specifies the degrees of the trigger (angle between zenith and sun or
angle between North and the sun (azimuth)). `C` specifies a command which is
passed to `sh -c` for execution when the trigger fires.

Examples for generic triggers:

"S:96:echo 'End of twilight. Good night!'"
"R:96:echo 'Beginning of twilight. Good morning!'"
"A:180:echo 'Sun is in the South.'"

The old `-sunrise-action "echo Morning" -sunrise-degrees 90` can be expressed
by a generic trigger as `"R:90:echo Morning"`.

```
