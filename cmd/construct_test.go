/*
Copyright 2023 F. Hoffmann-La Roche AG

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
package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_checkIfVersionSufficient(t *testing.T) {
	assert.True(t, checkIfVersionSufficient("2", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("2", ">", "1"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "2"))
	assert.False(t, checkIfVersionSufficient("1", ">", "2"))
	assert.False(t, checkIfVersionSufficient("2", ">", "2"))
	assert.True(t, checkIfVersionSufficient("2", ">=", "2"))
	assert.True(t, checkIfVersionSufficient("1.2", ">=", "1.2"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">=", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">", "1.2"))
	assert.False(t, checkIfVersionSufficient("1.2", ">=", "1.3"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.3"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "1.2"))
	assert.False(t, checkIfVersionSufficient("1", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.2", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("1.2", ">", "1"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "1.2.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.2.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">=", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2", ">=", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">=", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.3", ">", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "1.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.3"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("1.2.3", ">", "1"))
	assert.True(t, checkIfVersionSufficient("2", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("2", ">", "1.2.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "2"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "2"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.3.5"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.3.5"))
	assert.True(t, checkIfVersionSufficient("1.2.3.5", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.5", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.3"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.4", ">", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2.4"))
	assert.False(t, checkIfVersionSufficient("1.2", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">", "1.2"))
	assert.True(t, checkIfVersionSufficient("1.3", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.3", ">", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "1.3"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "1.3"))
	assert.False(t, checkIfVersionSufficient("1", ">=", "1.2.3.4"))
	assert.False(t, checkIfVersionSufficient("1", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("2", ">=", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("2", ">", "1.2.3.4"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">=", "1"))
	assert.True(t, checkIfVersionSufficient("1.2.3.4", ">", "1"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">=", "2"))
	assert.False(t, checkIfVersionSufficient("1.2.3.4", ">", "2"))
}
