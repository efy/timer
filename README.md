[![Go Report Card](https://goreportcard.com/badge/github.com/efy/timer)](https://goreportcard.com/report/github.com/efy/timer)

Timer is a small CLI tool for keeping track off time in the terminal.

## Usage

```
# start a new timer
timer -new -label=<label>

# stop a running timer
timer -stop -label=<label>

# start a stopped timer
timer -start -label=<label>

# delete a timer
timer -delete -label=<label>

# print timers as a table
timer -list
```

By default times are stored in `~/.timers` an alternate file can be passed using
the file flag or via the `TIMERS_FILE` environment variable.

## License

The code in this project is licensed under MIT license. See the [LICENSE](./LICENSE) file for
more details.
