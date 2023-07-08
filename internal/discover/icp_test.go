/**
# Copyright (c) NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package discover

import (
	"testing"

	"github.com/NVIDIA/xdxct-container-toolkit/internal/lookup"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestIPCMounts(t *testing.T) {
	l := ipcMounts(
		mounts{
			logger: logrus.New(),
			lookup: &lookup.LocatorMock{
				LocateFunc: func(path string) ([]string, error) {
					return []string{"/host/path"}, nil
				},
			},
			required: []string{"target"},
		},
	)

	mounts, err := l.Mounts()
	require.NoError(t, err)

	require.EqualValues(
		t,
		[]Mount{
			{
				HostPath: "/host/path",
				Path:     "/host/path",
				Options: []string{
					"ro",
					"nosuid",
					"nodev",
					"bind",
					"noexec",
				},
			},
		},
		mounts,
	)
}
