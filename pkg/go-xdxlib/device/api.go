/*
 * Copyright (c) 2022, XDXCT CORPORATION.  All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

 package device

 import (
	"github.com/XDXCT/xdxct-container-toolkit/pkg/go-xdxlib/xdxml"
 )
 
 // Interface provides the API to the 'device' package
 type Interface interface {
	 GetDevices() ([]Device, error)
	 VisitDevices(func(i int, d Device) error) error
 }
 
 type devicelib struct {
	 xdxml xdxml.Interface
 }
 
 var _ Interface = &devicelib{}
 
 func New(opts ...Option) Interface {
	 d := &devicelib{}
	 for _, opt := range opts {
		 opt(d)
	 }
	 if d.xdxml == nil {
		 d.xdxml = xdxml.New()
	 }
	 return d
 }
 
 func WithXdxml(xdxml xdxml.Interface) Option {
	 return func(d *devicelib) {
		 d.xdxml = xdxml
	 }
 }
 
 // Option defines a function for passing options to the New() call
 type Option func(*devicelib)
 