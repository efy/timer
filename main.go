package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"strings"
	"text/tabwriter"
	"time"
)

const (
	timeformat = time.RFC3339
)

var (
	// Overwritten with ldflag at build time
	version = "master"

	// Commands
	start  = flag.Bool("start", false, "Start a timer Requires label to be passed")
	stop   = flag.Bool("stop", false, "Stop the timer with the providied label")
	delete = flag.Bool("delete", false, "Delete the timer with the specified label")
	new    = flag.Bool("new", false, "Start a new timer with the provided label")
	list   = flag.Bool("list", false, "List the active timers")
	v      = flag.Bool("version", false, "Print the version")

	// Options
	label = flag.String("label", "", "Label of the timer to operate on when using a command")
	file  = flag.String("file", "", "File to store the timer information (Defaults to $HOME/.timers)")
)

func main() {
	flag.Parse()

	var f *os.File

	if *v {
		fmt.Println("version:", version)
		os.Exit(0)
	}

	if *file == "" {
		usr, err := user.Current()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		*file = usr.HomeDir + "/.timers"
	}

	f, err := os.Open(*file)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	trackers := parseFile(string(contents))
	f.Close()

	// Create a new entry
	if *new {
		if *label == "" {
			fmt.Println("must provide label using the -label flag")
			os.Exit(1)
		}

		_, err := getTracker(trackers, *label)
		if err == nil {
			fmt.Printf("Tracker %q already exists\n", *label)
		}

		e := entry{
			time.Now(),
			time.Time{},
		}
		tr := tracker{
			label: *label,
		}
		tr.entries = append(tr.entries, e)
		trackers = append(trackers, tr)

		err = save(trackers, *file)
		if err != nil {
			fmt.Println("could not save trackers:", err)
			os.Exit(1)
		}

		fmt.Printf("created tracker %q\n", *label)
		os.Exit(0)
	}

	if *stop {
		if *label == "" {
			fmt.Println("must provide label using the -label flag")
			os.Exit(1)
		}
		tr, err := getTracker(trackers, *label)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if tr.entries[len(tr.entries)-1].end.IsZero() {
			tr.entries[len(tr.entries)-1].end = time.Now()
			err := save(trackers, *file)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("stopped tracker %q\n", tr.label)
		} else {
			fmt.Printf("tracker %q not running\n", tr.label)
			os.Exit(1)
		}

		os.Exit(0)
	}

	if *start {
		if *label == "" {
			fmt.Println("must provide label using the -label flag")
			os.Exit(1)
		}

		tr, err := getTracker(trackers, *label)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if tr.entries[len(tr.entries)-1].end.IsZero() {
			fmt.Printf("tracker %q already running\n", tr.label)
			os.Exit(1)
		} else {
			e := entry{
				time.Now(),
				time.Time{},
			}

			tr.entries = append(tr.entries, e)
			updateTracker(trackers, tr)

			err := save(trackers, *file)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("started tracker %q\n", tr.label)
			os.Exit(0)
		}

		os.Exit(0)
	}

	if *delete {
		if *label == "" {
			fmt.Println("must provide label using the -label flag")
			os.Exit(1)
		}

		tr, err := getTracker(trackers, *label)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		trackers = deleteTracker(trackers, tr)
		err = save(trackers, *file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("deleted %q\n", tr.label)
		os.Exit(0)
	}

	// display
	if *list {
		if *label == "" {
			renderTable(trackers)
		} else {
			tr, err := getTracker(trackers, *label)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			renderItemTable(tr)
			os.Exit(0)
		}
	}
}

type tracker struct {
	label   string
	entries []entry
}

type entry struct {
	start time.Time
	end   time.Time
}

func save(tr []tracker, path string) error {
	sr := serializeTrackers(tr)
	err := ioutil.WriteFile(path, []byte(sr), 0777)
	if err != nil {
		return err
	}
	return nil
}

func deleteTracker(trackers []tracker, rem tracker) []tracker {
	for i, tr := range trackers {
		if tr.label == rem.label {
			trackers = append(trackers[:i], trackers[i+1:]...)
			return trackers
		}
	}
	return trackers
}

func updateTracker(trackers []tracker, new tracker) {
	for i, tr := range trackers {
		if tr.label == new.label {
			trackers[i] = new
			return
		}
	}
}

func getTracker(trackers []tracker, label string) (tracker, error) {
	for _, tr := range trackers {
		if tr.label == label {
			return tr, nil
		}
	}

	return tracker{}, fmt.Errorf("tracker %q does not exist", label)
}

func entryDuration(ent entry) time.Duration {
	if ent.end.IsZero() {
		//timer still running
		return time.Now().Sub(ent.start)
	}
	return ent.end.Sub(ent.start)
}

func calculateDuration(entries []entry) time.Duration {
	var dur time.Duration
	for _, v := range entries {
		dur += entryDuration(v)
	}
	return dur
}

func renderItemTable(tr tracker) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Start\tEnd\tDuration\n")
	for _, entry := range tr.entries {
		dur := entryDuration(entry)
		if entry.end.IsZero() {
			fmt.Fprintf(w, "%v\t%v\t%v\n", entry.start, "Running", dur)
			continue
		}
		fmt.Fprintf(w, "%v\t%v\t%v\n", entry.start, entry.end, dur)
	}
	w.Flush()
}

func renderTable(ts []tracker) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Label\tDuration\n")
	for _, tr := range ts {
		dur := calculateDuration(tr.entries)
		fmt.Fprintf(w, "%s\t%v\n", tr.label, dur)
	}
	w.Flush()
}

func serializeTrackers(ts []tracker) string {
	str := ""
	for _, t := range ts {
		str += fmt.Sprintf("%v ", t.label)
		for _, e := range t.entries {
			if e.start.IsZero() {
				continue
			} else {
				str += fmt.Sprintf("%s ", e.start.Format(timeformat))
			}
			if e.end.IsZero() {
				continue
			} else {
				str += fmt.Sprintf("%s ", e.end.Format(timeformat))
			}
		}
		str = strings.Trim(str, " \n\t")
		str += "\n"
	}

	return str
}

func parseFile(contents string) []tracker {
	var trackers []tracker

	lines := strings.Split(contents, "\n")

	for _, v := range lines {
		if v == "" {
			continue
		}
		t, err := parseLine(v)
		if err != nil {
			fmt.Println(err)
		}
		trackers = append(trackers, t)
	}

	return trackers
}

func parseLine(rec string) (tracker, error) {
	var label string
	var entries []entry

	parts := strings.Split(rec, " ")
	if len(parts) > 0 && parts[0] != "" {
		label = parts[0]
	}

	rest := parts[1:]

	for i := 0; i < len(rest); i += 2 {
		var s time.Time
		var f time.Time

		s, err := time.Parse(timeformat, rest[i])
		if err != nil {
			return tracker{}, err
		}

		if i+1 < len(rest) {
			f, err = time.Parse(timeformat, rest[i+1])
		}

		entries = append(entries, entry{s, f})
	}

	return tracker{
		label:   label,
		entries: entries,
	}, nil
}
