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
package bake

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseHCL(t *testing.T) {
	var dt = []byte(`
	group "default" {
		targets = ["db", "webapp"]
	}

	target "db" {
		context = "./db"
		tags = ["docker.io/tonistiigi/db"]
	}

	target "webapp" {
		context = "./dir"
		dockerfile = "Dockerfile-alternate"
		args = {
			buildno = "123"
		}
	}

	target "cross" {
		platforms = [
			"linux/amd64",
			"linux/arm64"
		]
	}

	target "webapp-plus" {
		inherits = ["webapp", "cross"]
		args = {
			IAMCROSS = "true"
		}
	}
	`)

	c, err := ParseHCL(dt)
	require.NoError(t, err)

	require.Equal(t, 1, len(c.Group))
	require.Equal(t, []string{"db", "webapp"}, c.Group["default"].Targets)

	require.Equal(t, 4, len(c.Target))
	require.Equal(t, "./db", *c.Target["db"].Context)

	require.Equal(t, 1, len(c.Target["webapp"].Args))
	require.Equal(t, "123", c.Target["webapp"].Args["buildno"])

	require.Equal(t, 2, len(c.Target["cross"].Platforms))
	require.Equal(t, []string{"linux/amd64", "linux/arm64"}, c.Target["cross"].Platforms)
}
