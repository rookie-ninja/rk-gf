// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

package rkgfinter

import "testing"

func TestNewNoopGLogger(t *testing.T) {
	log := NewNoopGLogger()
	log.Write([]byte{})
}
