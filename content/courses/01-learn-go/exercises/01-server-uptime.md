---
title: Server Uptime
difficulty: easy
after: variables
hints:
  - 'Work from the biggest unit down. A day has `24 * 60 * 60` seconds, so `uptime / 86400` is the number of whole days and `uptime % 86400` is what''s left over.'
  - 'Repeat the same trick with what''s left: divide by `3600` for hours, keep the remainder, divide that by `60` for minutes, and the final remainder is the seconds.'
  - 'For the second line, `uptime` is an `int`, so convert it first: `float64(uptime) / 3600`, then print it with `%.2f`.'
exercise:
  starter: |
    package main

    import "fmt"

    func main() {
    	uptime := 93784 // seconds since the Textio server started

    	// 1. Split uptime into whole days, hours, minutes and seconds
    	//    using integer division (/) and remainder (%).
    	_ = uptime // Go rejects unused variables; delete this line once you use uptime
    	days := 0
    	hours := 0
    	minutes := 0
    	seconds := 0

    	// 2. Work out the uptime in hours as a float64.
    	totalHours := 0.0

    	fmt.Printf("Uptime: %dd %dh %dm %ds\n", days, hours, minutes, seconds)
    	fmt.Printf("Hours: %.2f\n", totalHours)
    }
  solution: |
    package main

    import "fmt"

    func main() {
    	uptime := 93784 // seconds since the Textio server started

    	days := uptime / 86400
    	rest := uptime % 86400
    	hours := rest / 3600
    	rest = rest % 3600
    	minutes := rest / 60
    	seconds := rest % 60

    	totalHours := float64(uptime) / 3600

    	fmt.Printf("Uptime: %dd %dh %dm %ds\n", days, hours, minutes, seconds)
    	fmt.Printf("Hours: %.2f\n", totalHours)
    }
  expected_output: |
    Uptime: 1d 2h 3m 4s
    Hours: 26.05
---

Textio's status page shows how long the message server has been running. The
server only knows its uptime as a raw number of **seconds**, which isn't very
friendly to read.

Complete the program so that, for `uptime := 93784`, it prints:

```
Uptime: 1d 2h 3m 4s
Hours: 26.05
```

- The first line splits the uptime into whole **days**, **hours**, **minutes** and
  **seconds**.
- The second line is the total uptime in hours, as a decimal with two digits
  after the point.

## Example

`93784` seconds is 1 day (86,400 s), which leaves 7,384 s. That's 2 hours
(7,200 s), leaving 184 s, which is 3 minutes (180 s) and 4 seconds.

## Constraints

- Keep the two `fmt.Printf` lines as they are, and compute the values with
  `/`, `%` and a type conversion. Don't just type the answers in: change
  `uptime` to `3661` and you should see `Uptime: 0d 1h 1m 1s` and `Hours: 1.02`.
