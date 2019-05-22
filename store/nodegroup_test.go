/*
   Copyright The buildx Authors.

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
package store

import (
	"testing"

	"github.com/docker/buildx/util/platformutil"
	"github.com/stretchr/testify/require"
)

func TestNodeGroupUpdate(t *testing.T) {
	t.Parallel()

	ng := &NodeGroup{}
	err := ng.Update("foo", "foo0", []string{"linux/amd64"}, true, false)
	require.NoError(t, err)

	err = ng.Update("foo1", "foo1", []string{"linux/arm64", "linux/arm/v7"}, true, true)
	require.NoError(t, err)

	require.Equal(t, len(ng.Nodes), 2)

	// update
	err = ng.Update("foo", "foo2", []string{"linux/amd64", "linux/arm"}, true, false)
	require.NoError(t, err)

	require.Equal(t, len(ng.Nodes), 2)
	require.Equal(t, []string{"linux/amd64", "linux/arm/v7"}, platformutil.Format(ng.Nodes[0].Platforms))
	require.Equal(t, []string{"linux/arm64"}, platformutil.Format(ng.Nodes[1].Platforms))

	require.Equal(t, "foo2", ng.Nodes[0].Endpoint)

	// duplicate endpoint
	err = ng.Update("foo1", "foo2", nil, true, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "duplicate endpoint")

	err = ng.Leave("foo")
	require.NoError(t, err)

	require.Equal(t, len(ng.Nodes), 1)
	require.Equal(t, []string{"linux/arm64"}, platformutil.Format(ng.Nodes[0].Platforms))
}
