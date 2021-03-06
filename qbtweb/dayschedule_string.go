// Code generated by "stringer -type DaySchedule"; DO NOT EDIT.

package qbtweb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EveryDay-0]
	_ = x[EveryWeekday-1]
	_ = x[EveryWeekend-2]
	_ = x[EveryMonday-3]
	_ = x[EveryTuesday-4]
	_ = x[EveryWednesday-5]
	_ = x[EveryThursday-6]
	_ = x[EveryFriday-7]
	_ = x[EverySaturday-8]
	_ = x[EverySunday-9]
}

const _DaySchedule_name = "EveryDayEveryWeekdayEveryWeekendEveryMondayEveryTuesdayEveryWednesdayEveryThursdayEveryFridayEverySaturdayEverySunday"

var _DaySchedule_index = [...]uint8{0, 8, 20, 32, 43, 55, 69, 82, 93, 106, 117}

func (i DaySchedule) String() string {
	if i < 0 || i >= DaySchedule(len(_DaySchedule_index)-1) {
		return "DaySchedule(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _DaySchedule_name[_DaySchedule_index[i]:_DaySchedule_index[i+1]]
}
