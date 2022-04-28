/*
Copyright © 2020 Dell Inc. or its subsidiaries. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package node

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNodeLivenessCheck(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	t.Run("OK by default", func(t *testing.T) {
		check := NewLivenessCheckHelper(logger, nil, nil)
		assert.True(t, check.Check())
	})

	t.Run("Marked as failed, ttl not expired", func(t *testing.T) {
		check := NewLivenessCheckHelper(logger, nil, nil)
		check.Fail()
		assert.True(t, check.Check())
	})

	t.Run("Marked as failed, ttl expired", func(t *testing.T) {
		ttl := time.Millisecond * 10
		check := NewLivenessCheckHelper(logger, &ttl, nil)
		check.Fail()
		time.Sleep(ttl)
		assert.False(t, check.Check())
	})

	t.Run("Marked OK, no updates for a long time", func(t *testing.T) {
		ttl := time.Millisecond * 10
		check := NewLivenessCheckHelper(logger, &ttl, nil)
		time.Sleep(ttl)
		assert.True(t, check.Check())
	})
	t.Run("Marked OK, no updates for a long time, timeout expired", func(t *testing.T) {
		ttl := time.Millisecond * 10
		timeout := time.Millisecond * 20
		check := NewLivenessCheckHelper(logger, &ttl, &timeout)
		time.Sleep(timeout)
		assert.False(t, check.Check())
	})

	t.Run("Recover", func(t *testing.T) {
		ttl := time.Millisecond * 10
		check := NewLivenessCheckHelper(logger, &ttl, nil)
		check.Fail()
		time.Sleep(ttl)
		assert.False(t, check.Check())
		check.OK()
		assert.True(t, check.Check())
	})
}
