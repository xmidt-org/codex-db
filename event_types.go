// SPDX-FileCopyrightText: 2025 Comcast Cable Communications Management, LLC
// SPDX-License-Identifier: Apache-2.0

package db

//go:generate stringer -type=EventType

// EventType is an enum for specifying the type of event being stored.
type EventType int

const (
	// default event type
	Default EventType = iota

	// event type for online and offline events
	State
)

var (
	eventUnmarshal = map[string]EventType{
		"Default": Default,
		"State":   State,
	}
)

// ParseEventType returns the enum when given a string.
func ParseEventType(event string) EventType {
	if value, ok := eventUnmarshal[event]; ok {
		return value
	}
	return Default
}
