/*
Copyright 2021 CodeNotary, Inc. All rights reserved.

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

package immudb

import "errors"

var (
	ErrConstraintsNotImplemented = errors.New("constraints not implemented")
	ErrNotImplemented            = errors.New("not implemented")
	ErrCorruptedData             = errors.New("corrupted data")
	ErrTimeTravelNotAvailable    = errors.New("time travel is not available if verify flag is provided. This will change soon")
)
