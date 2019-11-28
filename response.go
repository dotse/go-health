// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"bytes"
	"encoding/json"
	"io"
)

// Response represents a health check response, containing any number of Checks.
type Response struct {
	Status      Status             `json:"status"`
	Version     string             `json:"version,omitempty"`
	ReleaseID   string             `json:"releaseId,omitempty"`
	Notes       []string           `json:"notes,omitempty"`
	Output      string             `json:"output,omitempty"`
	Checks      map[string][]Check `json:"checks,omitempty"`
	Links       []string           `json:"links,omitempty"`
	ServiceID   string             `json:"serviceID,omitempty"`
	Description string             `json:"description,omitempty"`
}

// ReadResponse reads a JSON Response from an io.Reader.
func ReadResponse(r io.Reader) (*Response, error) {
	var resp Response

	if err := json.NewDecoder(r).
		Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// AddChecks adds Checks to a Response and sets the status of the Response to
// the ‘worst’ status.
func (resp *Response) AddChecks(name string, checks ...Check) {
	for _, check := range checks {
		resp.Status = WorstStatus(resp.Status, check.Status)
	}

	if resp.Checks == nil {
		resp.Checks = make(map[string][]Check)
	} else if old, ok := resp.Checks[name]; ok {
		checks = append(old, checks...)
	}

	resp.Checks[name] = checks
}

// Good returns true if the Response is good, i.e. its status is ‘pass’ or
// ‘warn’.
func (resp *Response) Good() bool {
	return resp.Status == StatusPass || resp.Status == StatusWarn
}

// Write writes a JSON Response to an io.Writer.
func (resp *Response) Write(w io.Writer) (int64, error) {
	bajts, err := json.Marshal(resp)
	if err != nil {
		return 0, err
	}

	return bytes.NewBuffer(bajts).
		WriteTo(w)
}
