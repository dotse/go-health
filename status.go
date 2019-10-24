// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"encoding/json"
)

const (
	// StatusPass is "pass".
	StatusPass Status = iota
	// StatusWarn is "warn".
	StatusWarn Status = iota
	// StatusFail is "fail".
	StatusFail Status = iota
)

// Status is the status part of a Response or Check.
type Status uint8

// WorstStatus returns the worst of a number of statuses, where "warn" is worse
// than "pass" but "fail" is worse than "warn".
func WorstStatus(status Status, statuses ...Status) (worst Status) {
	worst = status
	for _, other := range statuses {
		if other > worst {
			worst = other
		}
	}

	return
}

// MarshalJSON encodes a status as a JSON string.
func (status Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(status.String())
}

// MarshalText encodes a status as a string.
func (status Status) MarshalText() ([]byte, error) {
	return []byte(status.String()), nil
}

// String turns a status into a string.
func (status Status) String() string {
	return statusStringMap()[status]
}

// UnmarshalJSON decodes a status from a JSON string.
func (status *Status) UnmarshalJSON(data []byte) error {
	var tmp string
	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	for k, v := range statusStringMap() {
		if tmp == v {
			*status = k
			return nil
		}
	}

	return &json.UnsupportedValueError{
		Str: tmp,
	}
}

func statusStringMap() map[Status]string {
	return map[Status]string{
		StatusPass: "pass",
		StatusFail: "fail",
		StatusWarn: "warn",
	}
}
