// Copyright © 2019 The Swedish Internet Foundation
//
// Distributed under the MIT License. (See accompanying LICENSE file or copy at
// <https://opensource.org/licenses/MIT>.)

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatus_String(t *testing.T) {
	assert.Equal(t, "pass", StatusPass.String())
	assert.Equal(t, "warn", StatusWarn.String())
	assert.Equal(t, "fail", StatusFail.String())
}

func TestWorstStatus(t *testing.T) {
	assert.Equal(t, StatusPass, WorstStatus(StatusPass))
	assert.Equal(t, StatusWarn, WorstStatus(StatusWarn))
	assert.Equal(t, StatusFail, WorstStatus(StatusFail))

	assert.Equal(t, StatusWarn, WorstStatus(StatusPass, StatusWarn))
	assert.Equal(t, StatusFail, WorstStatus(StatusWarn, StatusFail))
	assert.Equal(t, StatusFail, WorstStatus(StatusPass, StatusFail))

	assert.Equal(t, StatusWarn, WorstStatus(StatusWarn, StatusWarn))
	assert.Equal(t, StatusFail, WorstStatus(StatusFail, StatusWarn))
	assert.Equal(t, StatusFail, WorstStatus(StatusPass, StatusFail, StatusPass))
}
