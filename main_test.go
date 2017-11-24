package main

import "testing"

func TestParseLine(t *testing.T) {
	record := "testlabel 2012-11-01T22:08:41+00:00 2012-11-01T22:20:41+00:00 2012-12-02T22:08:41+00:00"

	got, err := parseLine(record)
	if err != nil {
		t.Error(err)
	}

	if got.label != "testlabel" {
		t.Error("expected", "testlabel", "got", got.label)
	}

	if len(got.entries) != 2 {
		t.Error("expected 2 entries got:", len(got.entries))
	}

	if !got.entries[1].end.IsZero() {
		t.Error("expected zero time for second entry")
	}

	if got.entries[0].start.IsZero() {
		t.Error("expected the time to be non zero")
	}
}

func TestParseFile(t *testing.T) {
	record := "testlabel1 2012-11-01T22:08:41+00:00 2012-11-01T22:20:41+00:00 2012-11-02T22:08:41+00:00\n"
	record += "testlabel2 2012-11-01T22:08:41+00:00 2012-11-01T22:20:41+00:00 2012-11-02T22:08:41+00:00\n"
	record += "testlabel3 2012-11-01T22:08:41+00:00 2012-11-01T22:20:41+00:00 2012-11-02T22:08:41+00:00\n"

	got := parseFile(record)

	if len(got) != 3 {
		t.Error("epected 3 trackers got", len(got))
	}
}
