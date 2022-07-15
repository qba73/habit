package habit_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/habit"
)

func TestHabit_StartsNewActivityWithNameAndInitialDate(t *testing.T) {
	t.Parallel()

	fakeTerminal := &bytes.Buffer{}
	habitName := "jog"
	h, err := habit.New(habitName, habit.WithOutput(fakeTerminal))
	if err != nil {
		t.Fatal(err)
	}
	h.Output = fakeTerminal

	want := fmt.Sprintf(messageStartNewHabit, habitName)
	got := fakeTerminal.String()

	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}

	wantTime := time.Now().UTC().Truncate(24 * time.Hour)
	gotTime := h.Dates[0]

	if wantTime != gotTime {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestHabit_DoesNotLogDuplicatedActivityOnTheSameDay(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog", habit.WithOutput(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	h.LogActivity()

	want := 1
	got := len(h.Dates)
	if want != got {
		t.Fatalf("want %d, got %d", want, got)
	}
}

func TestHabit_RecordActivityOnNextDay(t *testing.T) {
	t.Parallel()

	h, err := habit.New("jog", habit.WithOutput(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	h.Dates = []time.Time{
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2),
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1),
	}

	h.LogActivity()

	want := 3
	got := len(h.Dates)
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestHabit_PrintsMessageOnThreeDayStreak(t *testing.T) {
	t.Parallel()
	fakeTerminal := &bytes.Buffer{}
	habitName := "jog"
	h, err := habit.New(habitName, habit.WithOutput(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	h.Dates = []time.Time{
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2),
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1),
	}

	h.Output = fakeTerminal
	h.LogActivity()

	want := fmt.Sprintf(messageOnContinousStreak, h.Name, 3)
	got := fakeTerminal.String()
	if want != got {
		t.Errorf("want %s, got %s", want, got)
	}
}

func TestHabits_PrintsMessageOnStartingNewStreakAfterBreak(t *testing.T) {
	t.Parallel()
	fakeTerminal := &bytes.Buffer{}
	habitName := "jog"
	h, err := habit.New(habitName, habit.WithOutput(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	h.Dates = []time.Time{
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -5),
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -4),
	}

	h.Output = fakeTerminal
	h.LogActivity()

	wantMessage := fmt.Sprintf(messageOnStaringNewStreakAfterBreak, h.Name, 4)
	gotMessage := fakeTerminal.String()
	if wantMessage != gotMessage {
		t.Errorf("want %s, got %s", wantMessage, gotMessage)
	}

	wantDates := []time.Time{
		time.Now().UTC().Truncate(24 * time.Hour),
	}
	gotDates := h.Dates

	if !cmp.Equal(wantDates, gotDates) {
		t.Error(cmp.Diff(wantDates, gotDates))
	}

}

func TestHabit_PrintsNumberOfDaysOnNotBrokenCurrentStreak(t *testing.T) {
	t.Parallel()
	fakeTerminal := &bytes.Buffer{}
	habitName := "jog"
	h, err := habit.New(habitName, habit.WithOutput(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	h.Dates = []time.Time{
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -4),
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -3),
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -2),
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1),
	}
	h.Output = fakeTerminal
	want := 4
	got := h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	wantMessage := fmt.Sprintf(messageOnNotBrokenStreakCheck, 4, habitName)
	gotMessage := fakeTerminal.String()
	if wantMessage != gotMessage {
		t.Errorf("want %s, got %s", wantMessage, gotMessage)
	}

}

func TestHabit_PrintsNumberOfDaysSinceBrokenStreak(t *testing.T) {
	t.Parallel()
	fakeTerminal := &bytes.Buffer{}
	habitName := "jog"
	h, err := habit.New(habitName, habit.WithOutput(io.Discard))
	if err != nil {
		t.Fatal(err)
	}

	h.Dates = []time.Time{
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -4),
		time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -3),
	}
	h.Output = fakeTerminal

	want := 3
	got := h.Check()

	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	wantMessage := fmt.Sprintf(messageOnBrokenStreakCheck, want, habitName)
	gotMessage := fakeTerminal.String()
	if wantMessage != gotMessage {
		t.Errorf("want %s, got %s", wantMessage, gotMessage)
	}
}

var (
	messageStartNewHabit                = "Good luck with your new habit '%s'! Don't forget to do it again tomorrow."
	messageOnContinousStreak            = "Nice work: you've done the habit '%s' for %d days in a row now. Keep it up!"
	messageOnStaringNewStreakAfterBreak = "You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!"
	messageOnBrokenStreakCheck          = "It's been %d days since you did '%s'. It's okay, life happens. Get back on that horse today!"
	messageOnNotBrokenStreakCheck       = "You're currently on a %d-day streak for '%s'. Stick to it!"
)
