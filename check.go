// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"time"
)

// Check represent a single health check point.
type Check struct {
	ComponentID       string      `json:"componentId,omitempty"`
	ComponentType     string      `json:"componentType,omitempty"`
	ObservedValue     interface{} `json:"observedValue,omitempty"`
	ObservedUnit      string      `json:"observedUnit,omitempty"`
	Status            Status      `json:"status"`
	AffectedEndpoints []string    `json:"affectedEndpoints,omitempty"`
	Time              *time.Time  `json:"time,omitempty"`
	Output            string      `json:"output,omitempty"`
	Links             []string    `json:"links,omitempty"`
}

// Good returns true if the Check is good, i.e. its status is ‘pass’ or
// ‘warn’.
func (check *Check) Good() bool {
	return check.Status == StatusPass || check.Status == StatusWarn
}

// SetObservedTime sets the observedValue field to a time duration (and the
// observedUnit field to the correct unit).
func (check *Check) SetObservedTime(duration time.Duration) {
	check.ObservedValue = duration.Nanoseconds()
	check.ObservedUnit = "ns"
}
