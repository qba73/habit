package habit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/habit"
)

func TestDayDiff_CalculatesDurationInFullDaysBetweenDates(t *testing.T) {
	t.Parallel()

	tt := []struct {
		name  string
		start string
		end   string
		want  int
	}{
		{
			name:  "with twenty day gap",
			start: "2022-10-01T23:00:00Z",
			end:   "2022-10-21T01:00:00Z",
			want:  20,
		},
		{
			name:  "with three day gap",
			start: "2022-10-31T23:00:00Z",
			end:   "2022-11-03T01:00:00Z",
			want:  3,
		},
		{
			name:  "with one day gap",
			start: "2022-10-31T23:00:00Z",
			end:   "2022-11-01T01:00:00Z",
			want:  1,
		},
		{
			name:  "wth no gap between dates",
			start: "2022-10-30T01:00:00Z",
			end:   "2022-10-30T14:00:00Z",
			want:  0,
		},
		{
			name:  "with multiple days gap",
			start: "2022-09-20T23:00:00Z",
			end:   "2022-11-30T01:00:00Z",
			want:  71,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			startTime, err := time.Parse(time.RFC3339, tc.start)
			if err != nil {
				t.Fatal(err)
			}

			stopTime, err := time.Parse(time.RFC3339, tc.end)
			if err != nil {
				t.Fatal(err)
			}

			got := habit.DayDiff(startTime, stopTime)
			if tc.want != got {
				t.Error(cmp.Diff(tc.want, got))
			}
		})
	}

}

func TestRoundDate_RoundsHabitLogTimeToAFullDay(t *testing.T) {
	t.Parallel()

	testTime, err := time.Parse(time.RFC3339, "2022-11-01T23:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	want, err := time.Parse(time.RFC3339, "2022-11-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	got := habit.RoundDate(testTime)
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestNew_CreatesActivityWithNameAndInitialDate(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-11-01T02:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	want, err := time.Parse(time.RFC3339, "2022-11-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	got := h.Date
	if !cmp.Equal(want.UTC(), got) {
		t.Errorf(cmp.Diff(want, got))
	}
}

func TestStart_ConfiguresValidHabit(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	habitName := "jog"
	h, err := habit.New(habitName)
	if err != nil {
		t.Fatal(err)
	}
	got := h.Start()
	want := fmt.Sprintf("Good luck with your new habit '%s'. Don't forget to do it tomorrow.\n", habitName)
	if want != got {
		t.Error(cmp.Diff(want, got))
	}

	gotDate := h.Date
	wantDate, err := time.Parse(time.RFC3339, "2022-10-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(wantDate.UTC(), gotDate.UTC()) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestLog_DoesNotDuplicateActivityOnTheSameDay(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	testTime, err = time.Parse(time.RFC3339, "2022-09-01T15:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	days, msg := h.Log()

	wantDays := 1
	wantMsg := ""
	if !cmp.Equal(wantDays, days) {
		t.Error(cmp.Diff(wantDays, days))
	}
	if !cmp.Equal(wantMsg, msg) {
		t.Error(cmp.Diff(wantMsg, msg))
	}

	wantDate, err := time.Parse(time.RFC3339, "2022-09-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

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
	testTime, err := time.Parse(time.RFC3339, "2022-11-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	want := 0
	got, _ := h.Check()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	checkTime, err := time.Parse(time.RFC3339, "2022-11-02T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return checkTime
	}

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
	testTime, err := time.Parse(time.RFC3339, "2022-10-27T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	checkTime, err := time.Parse(time.RFC3339, "2022-10-30T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	h.Date = checkTime

	got, msg := h.Check()
	wantDays := 3
	if wantDays != got {
		t.Errorf("want %d, got %d", wantDays, got)
	}
	wantMsg := fmt.Sprintf("It's been %d days since you did '%s'. It's ok, life happens. Get back on that horse today!\n", wantDays, h.Name)
	if wantMsg != msg {
		t.Error(cmp.Diff(wantMsg, msg))
	}
}

func TestLog_RecordsActivityOnNextDayOnNotBrokenStreak(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-02T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	habit.Now = func() time.Time {
		habitLogTime, err := time.Parse(time.RFC3339, "2022-10-03T00:00:00Z")
		if err != nil {
			t.Fatal(err)
		}
		return habitLogTime
	}

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
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	checkTime, err := time.Parse(time.RFC3339, "2022-09-05T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return checkTime
	}

	want := 1
	got, msg := h.Log()
	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}

	wantDays := 4
	wantMsg := fmt.Sprintf("You last did the habit '%s' %d days ago, so you're starting a new streak today. Good luck!\n", h.Name, wantDays)
	if wantMsg != msg {
		t.Error(cmp.Diff(wantMsg, msg))
	}
}

func TestNewFileStore_CreatesStoreOnValidPathInput(t *testing.T) {
	habitFilepath := t.TempDir() + "/habit-initial.json"
	store, err := habit.NewFileStore(habit.WithFilePath(habitFilepath))
	if err != nil {
		t.Fatal(err)
	}

	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}
	h.Start()

	err = store.Save(h)
	if err != nil {
		t.Fatal(err)
	}

	h2, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(h2, h) {
		t.Error(cmp.Diff(h2, h))
	}
}

func TestFileStore_LoadsHabit(t *testing.T) {
	habitFilepath := "testdata/habit.json"

	store, err := habit.NewFileStore(habit.WithFilePath(habitFilepath))
	if err != nil {
		t.Fatal(err)
	}

	got, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	want := &habit.Habit{
		Name:   "walk",
		Date:   time.Date(2022, 10, 03, 00, 00, 00, 00, time.UTC),
		Streak: 1,
	}
	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestFileStore_SavesHabit(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	h, err := habit.New("run")
	if err != nil {
		t.Fatal(err)
	}
	h.Start()

	path := t.TempDir() + "/habit.json"

	store, err := habit.NewFileStore(habit.WithFilePath(path))
	if err != nil {
		t.Fatal(err)
	}
	err = store.Save(h)
	if err != nil {
		t.Fatal(err)
	}

	h2, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}

	if !cmp.Equal(h, h2) {
		t.Errorf("want %+v, got %+v", h, h2)
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
