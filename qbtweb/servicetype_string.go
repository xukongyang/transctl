// Code generated by "stringer -type ServiceType -trimprefix Service"; DO NOT EDIT.

package qbtweb

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ServiceUseDyDNS-0]
	_ = x[ServiceUseNOIP-1]
}

const _ServiceType_name = "UseDyDNSUseNOIP"

var _ServiceType_index = [...]uint8{0, 8, 15}

func (i ServiceType) String() string {
	if i < 0 || i >= ServiceType(len(_ServiceType_index)-1) {
		return "ServiceType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ServiceType_name[_ServiceType_index[i]:_ServiceType_index[i+1]]
}
