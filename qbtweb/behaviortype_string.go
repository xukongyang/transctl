// Code generated by "stringer -type BehaviorType -trimprefix Behavior"; DO NOT EDIT.

package qbtweb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[BehaviorPause-0]
	_ = x[BehaviorRemove-1]
}

const _BehaviorType_name = "PauseRemove"

var _BehaviorType_index = [...]uint8{0, 5, 11}

func (i BehaviorType) String() string {
	if i < 0 || i >= BehaviorType(len(_BehaviorType_index)-1) {
		return "BehaviorType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _BehaviorType_name[_BehaviorType_index[i]:_BehaviorType_index[i+1]]
}
