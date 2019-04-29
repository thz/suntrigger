## suntrigger - actions triggered by the sun

### What?

A tool to trigger actions based on the sun's position.

### How?

This tool takes the observers location as input and continuously calculates the sun's position (azimuth and angle to zenith). The sunrise and sunset can be user-defined by giving the angle to zenith. This allows to trigger based on the sun dropping below the horizon or start/end of civil/nautic/astronomical dusk or dawn.

### Why?

I control my automated shutters with this.

### Usage:

```
Usage of ./suntrigger
  -latitude float
    	latitude (degrees) of observation location
  -longitude float
    	longitude (degrees) of observation location
  -simulate
    	let time pass unnatural quickly
  -sunrise-action string
    	action to trigger on sunrise
  -sunrise-degrees string
    	degrees between zentih and sun to define sunrise (default "90")
  -sunset-action string
    	action to trigger on sunset
  -sunset-degrees string
    	degrees between zentih and sun to define sunset (default "90")
```
