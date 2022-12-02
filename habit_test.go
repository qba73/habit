package habit_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/qba73/habit"
)

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

func TestNew_CreatesNewHabit(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-11-01T02:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	got, err := habit.New("jog")
	if err != nil {
		t.Fatal(err)
	}

	date, err := time.Parse(time.RFC3339, "2022-11-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	want := &habit.Habit{
		Name:   "jog",
		Date:   date,
		Streak: 1,
	}

	if !cmp.Equal(want, got) {
		t.Error(cmp.Diff(want, got))
	}
}

func TestStart_StartsHabit(t *testing.T) {
	testTime, err := time.Parse(time.RFC3339, "2022-10-01T01:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	habit.Now = func() time.Time {
		return testTime
	}

	habitName := "jog"
	gotHabit, err := habit.New(habitName)
	if err != nil {
		t.Fatal(err)
	}
	got := gotHabit.Start()
	want := fmt.Sprintf("Good luck with your new habit '%s'. Don't forget to do it tomorrow.\n", habitName)
	if want != got {
		t.Error(cmp.Diff(want, got))
	}

	date, err := time.Parse(time.RFC3339, "2022-10-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	wantHabit := &habit.Habit{
		Name:   habitName,
		Date:   date,
		Streak: 1,
	}

	if !cmp.Equal(wantHabit, gotHabit) {
		t.Error(cmp.Diff(wantHabit, gotHabit))
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

	gotHabit, err := habit.New("jog")
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

	days, msg := gotHabit.Log()

	wantDays := 1
	if !cmp.Equal(wantDays, days) {
		t.Error(cmp.Diff(wantDays, days))
	}

	wantMsg := ""
	if !cmp.Equal(wantMsg, msg) {
		t.Error(cmp.Diff(wantMsg, msg))
	}

	wantDate, err := time.Parse(time.RFC3339, "2022-09-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}

	wantHabit := &habit.Habit{
		Name:   "jog",
		Date:   wantDate,
		Streak: 1,
	}

	if !cmp.Equal(wantHabit, gotHabit) {
		t.Error(cmp.Diff(wantHabit, gotHabit))
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

	tt := []struct {
		checkTime   string
		wantDayDiff int
	}{
		{checkTime: "2022-10-30T01:00:00Z", wantDayDiff: 3},
		{checkTime: "2022-01-31T05:00:00Z", wantDayDiff: 269},
		{checkTime: "2022-11-28T15:00:00Z", wantDayDiff: 32},
	}

	for _, tc := range tt {
		checkTime, err := time.Parse(time.RFC3339, tc.checkTime)
		if err != nil {
			t.Fatal(err)
		}
		h.Date = checkTime

		got, msg := h.Check()
		wantDays := tc.wantDayDiff
		if wantDays != got {
			t.Errorf("want %d, got %d", tc.wantDayDiff, got)
		}
		wantMsg := fmt.Sprintf("It's been %d days since you did '%s'. It's ok, life happens. Get back on that horse today!\n", tc.wantDayDiff, h.Name)
		if wantMsg != msg {
			t.Error(cmp.Diff(wantMsg, msg))
		}
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

// func TestHabits_AddHabit(t *testing.T) {
// 	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	habit.Now = func() time.Time {
// 		return testTime
// 	}

// 	h, err := habit.New("walk")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	h.Start()

// 	store := habit.Habits{}
// 	store.Add(*h)

// 	got := store.Read("walk")

// 	if !cmp.Equal(want, got) {
// 		t.Error(cmp.Diff(want, got))
// 	}
// }

// func TestHabits_ListAllTrackedHabits(t *testing.T) {

// }

//
//}

// func TestFileStore_LoadsHabit(t *testing.T) {
// 	store, err := habit.NewFileStore(
// 		habit.WithFilePath("testdata/habit.json"),
// 	)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	got, err := store.Load("walk")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	want := &habit.Habit{
// 		Name:   "walk",
// 		Date:   time.Date(2022, 10, 03, 00, 00, 00, 00, time.UTC),
// 		Streak: 1,
// 	}
// 	if !cmp.Equal(want, got) {
// 		t.Error(cmp.Diff(want, got))
// 	}
// }

// func TestFileStore_SavesHabit(t *testing.T) {
// 	testTime, err := time.Parse(time.RFC3339, "2022-09-01T03:00:00Z")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	habit.Now = func() time.Time {
// 		return testTime
// 	}

// 	h, err := habit.New("run")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	h.Start()

// 	path := t.TempDir() + "/habit.json"

// 	store, err := habit.NewFileStore(habit.WithFilePath(path))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	err = store.Save(*h)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	h2, err := store.Load("run")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	if !cmp.Equal(h, h2) {
// 		t.Errorf("want %+v, got %+v", h, h2)
// 	}
// }

// func TestFileStore_ListsAllTrackedHabits(t *testing.T) {
// 	path := "testdata/habits.json"
// 	store, err := habit.NewFileStore(habit.WithFilePath(path))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	got, err := store.List()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	want := []habit.Habit{
// 		{Name: "walk", Date: time.Date(2022, 10, 03, 00, 00, 00, 00, time.UTC), Streak: 1},
// 		{Name: "jog", Date: time.Date(2022, 10, 01, 00, 00, 00, 00, time.UTC), Streak: 2},
// 		{Name: "read", Date: time.Date(2022, 10, 23, 00, 00, 00, 00, time.UTC), Streak: 3},
// 	}

// 	if !cmp.Equal(want, got) {
// 		t.Error(want, got)
// 	}
// }

// func TestFileStore_LoadHabitErrorsOnNotExistingHabit(t *testing.T) {
// 	t.Parallel()

// 	path := "testdata/habits.json"
// 	store, err := habit.NewFileStore(habit.WithFilePath(path))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	_, err = store.Load("skate")
// 	if err == nil {
// 		t.Fatal("want error, got nil")
// 	}
// }

// func TestFileStore_LoadExistingHabit(t *testing.T) {
// 	t.Parallel()

// 	path := "testdata/habits.json"
// 	store, err := habit.NewFileStore(habit.WithFilePath(path))
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	got, err := store.Load("jog")
// 	if err != nil {
// 		t.Fatal()
// 	}
// 	want := &habit.Habit{Name: "jog", Date: time.Date(2022, 10, 01, 00, 00, 00, 00, time.UTC), Streak: 2}
// 	if !cmp.Equal(want, got) {
// 		t.Error(cmp.Diff(want, got))
// 	}
// }

// func TestMain(m *testing.M) {
// 	os.Exit(testscript.RunMain(m, map[string]func() int{
// 		"habctl": habit.Main,
// 	}))
// }

// func TestHabit(t *testing.T) {
// 	testscript.Run(t, testscript.Params{
// 		Dir:  "testdata/script",
// 		Cmds: map[string]func(ts *testscript.TestScript, neg bool, args []string){"date": cmdDate},
// 	})
// }

// func cmdDate(ts *testscript.TestScript, neg bool, args []string) {
// 	if neg {
// 		ts.Fatalf("unsupported: ! date")
// 	}
// 	if len(args) != 2 {
// 		ts.Fatalf("usage: date filepath -1")
// 	}
// 	// Verify 2nd arg
// 	dayShift, err := strconv.Atoi(args[1])
// 	if err != nil {
// 		ts.Fatalf("expected int of max value -1")
// 	}
// 	// Verify 1st arg
// 	filepath := args[0]
// 	data, err := os.ReadFile(filepath)
// 	if err != nil {
// 		ts.Fatalf("reading file: %s, %v", filepath, err)
// 	}

// 	var h habit.Habit
// 	err = json.Unmarshal(data, &h)
// 	if err != nil {
// 		ts.Fatalf("unmarshaling data: %v", err)
// 	}

// 	newDate := habit.RoundDate(h.Date.AddDate(0, 0, dayShift))
// 	h.Date = newDate

// d, err := json.Marshal(h)
//
//	if err != nil {
//		ts.Fatalf("marshaling updated habit date")
//	}
//
//	if err = os.WriteFile(filepath, d, 0644); err != nil {
//		ts.Fatalf("saving updated filestore: %s, %v", filepath, err)
//	}
