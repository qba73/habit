package habit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/habit"
)

func TestHabit_StartsNewActivityWithNameAndInitialDate(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Now().UTC().Truncate(24 * time.Hour)
	got := h.Date
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

func TestStart_ConfiguresValidHabit(t *testing.T) {
	t.Parallel()
	habitName := "jog"
	h, err := habit.New(habitName)
	if err != nil {
		t.Error(err)
	}
	got := h.Start()
	want := fmt.Sprintf("Good luck with your new habit '%s'. Don't forget to do it tomorrow.", habitName)
	if want != got {
		t.Error(cmp.Diff(want, got))
	}

	gotDate := h.Date
	wantDate := time.Now().UTC().Truncate(24 * time.Hour)
	if wantDate != gotDate {
		t.Error(cmp.Diff(want, got))
	}

}

func TestLog_DoesNotDuplicateActivityOnTheSameDay(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	h.Log()
	wantDate := time.Now().UTC().Truncate(24 * time.Hour)
	gotDate := h.Date
	if wantDate != gotDate {
		t.Fatalf("want %v, got %v", wantDate, gotDate)
	}

	wantStreak := 1
	gotStreak := h.Streak
	if wantStreak != gotStreak {
		t.Errorf("want %d, got %d", wantStreak, gotStreak)
	}
}

func TestCheck_ReportsValidStreakLengthOnNotBrokenStreak(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	want := 0
	got, _ := h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)
	want = 1
	got, msg := h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
	wantMessage := fmt.Sprintf("You're currently on a %d-day streak for '%s'. Stick to it!\n", want, h.Name)
	if wantMessage != msg {
		t.Error(cmp.Diff(wantMessage, msg))
	}
}

func TestCheck_ReportsBrokenStreak(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	// Setup date 5 days ago to simulate that the habit was logged 5 days ago.
	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -5)

	got, msg := h.Check()
	wantDays := 5
	if wantDays != got {
		t.Errorf("want %d, got %d", wantDays, got)
	}
	wantMsg := fmt.Sprintf("It's been %d days since you did '%s'. It's ok, life happens. Get back on that horse today!\n", wantDays, h.Name)
	if wantMsg != msg {
		t.Error(cmp.Diff(wantMsg, msg))
	}
}

func TestLog_RecordsActivityOnNextDayOnNotBrokenStreak(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)

	got, msg := h.Log()
	want := 2
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	wantMsg := fmt.Sprintf("Nice work: you've done the habit '%s' for %d days in a row now. Keep it up!\n", h.Name, want)
	if wantMsg != msg {
		t.Error(cmp.Diff(wantMsg, msg))
	}
}

func TestLog_StartsNewStreakAfterBrokenStreak(t *testing.T) {
	t.Parallel()
	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	// Setup date 5 days ago to simulate that the habit was logged 5 days ago.
	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -5)

	want := 1
	got, msg := h.Log()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	wantDays := 5
	wantMsg := fmt.Sprintf("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, wantDays)
	if wantMsg != msg {
		t.Error(cmp.Diff(wantMsg, msg))
	}
}

func TestLoadHabitFromFile(t *testing.T) {
	t.Parallel()
	habitFilepath := "testdata/habit.json"
	got, err := habit.LoadFromFile(habitFilepath)
	if err != nil {
		t.Fatal(err)
	}
	want := &habit.Habit{
		Name:   "walk",
		Date:   time.Date(2022, 07, 15, 00, 00, 00, 00, time.UTC),
		Streak: 1,
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestSaveHabitToFile(t *testing.T) {
	t.Parallel()
	h, err := habit.New("run")
	if err != nil {
		t.Fatal(err)
	}
	h.Date = time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -1)

	path := t.TempDir() + "/habit.json"
	err = habit.SaveToFile(path, h)
	if err != nil {
		t.Fatal(err)
	}
	h2, err := habit.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(h, h2) {
		t.Errorf("want %+v, got %+v", h, h2)
	}
}

func TestHabit_SavesUpdatedHabitDataToFile(t *testing.T) {
	t.Parallel()
	habitFilepath := "testdata/habit.json"
	h, err := habit.LoadFromFile(habitFilepath)
	if err != nil {
		t.Fatal(err)
	}

	h.Log()

	path := t.TempDir() + "/path.json"

	habit.SaveToFile(path, h)

	h2, err := habit.LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(h2, h) {
		t.Error(cmp.Diff(h2, h))
	}
}

// func TestMain(m *testing.M) {
// 	testscript.RunMain(m, map[string]func() int{
// 		"habit": func() int {
// 			habit.RunCLI()
// 			return 0
// 		},
// 	})
// }
// func TestScript(t *testing.T) {
// 	testscript.Run(t, testscript.Params{Dir: "testdata/script"})
// }
