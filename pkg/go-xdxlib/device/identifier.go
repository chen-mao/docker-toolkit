/*
 * Copyright (c) XDXCT CORPORATION.  All rights reserved.
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
	"strconv"

	"github.com/google/uuid"
)

// Identifier can be used to refer to a GPU.
// This includes a device index or UUID.
type Identifier string

// IsGpuIndex checks if an identifier is a full GPU index
func (i Identifier) IsGpuIndex() bool {
	if _, err := strconv.ParseUint(string(i), 10, 0); err != nil {
		return false
	}
	return true
}

// IsUUID checks if an identifier is a UUID
func (i Identifier) IsUUID() bool {
	return i.IsGpuUUID()
}

// IsGpuUUID checks if an identifier is a GPU UUID
func (i Identifier) IsGpuUUID() bool {
	_, err := uuid.Parse(string(i))
	return err == nil
}
