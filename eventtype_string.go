// Code generated by "stringer -type=EventType"; DO NOT EDIT.

package db

import "strconv"

const _EventType_name = "DefaultState"

var _EventType_index = [...]uint8{0, 7, 12}

func (i EventType) String() string {
	if i < 0 || i >= EventType(len(_EventType_index)-1) {
		return "EventType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EventType_name[_EventType_index[i]:_EventType_index[i+1]]
}
