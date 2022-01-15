// Copyright (c) 2021 rookie-ninja
//
// Use of this source code is governed by an Apache-style
// license that can be found in the LICENSE file.

// Package rkgfinter provides common utility functions for middleware of GoFrame framework
package rkgfinter

import (
	"github.com/gogf/gf/v2/os/glog"
)

type noopWriter struct{}

func (w noopWriter) Write([]byte) (n int, err error) {
	return 0, nil
}

func NewNoopGLogger() *glog.Logger {
	return glog.NewWithWriter(noopWriter{})
}
