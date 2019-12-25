// Code generated by "stringer -type Encryption -trimprefix Encryption"; DO NOT EDIT.

package qbtweb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[EncryptionPreferred-0]
	_ = x[EncryptionForceOn-1]
	_ = x[EncryptionForceOff-2]
}

const _Encryption_name = "PreferredForceOnForceOff"

var _Encryption_index = [...]uint8{0, 9, 16, 24}

func (i Encryption) String() string {
	if i < 0 || i >= Encryption(len(_Encryption_index)-1) {
		return "Encryption(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Encryption_name[_Encryption_index[i]:_Encryption_index[i+1]]
}
