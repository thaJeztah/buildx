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
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadTargets(t *testing.T) {
	t.Parallel()
	tmpdir, err := ioutil.TempDir("", "bake")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	fp := filepath.Join(tmpdir, "config.hcl")
	err = ioutil.WriteFile(fp, []byte(`
target "dep" {
}

target "webapp" {
	dockerfile = "Dockerfile.webapp"
	inherits = ["dep"]
}`), 0600)
	require.NoError(t, err)

	ctx := context.TODO()

	m, err := ReadTargets(ctx, []string{fp}, []string{"webapp"}, nil)
	require.NoError(t, err)

	require.Equal(t, "Dockerfile.webapp", *m["webapp"].Dockerfile)
	require.Equal(t, ".", *m["webapp"].Context)
}

func TestReadTargetsCompose(t *testing.T) {
	t.Parallel()
	tmpdir, err := ioutil.TempDir("", "bake")
	require.NoError(t, err)
	defer os.RemoveAll(tmpdir)

	fp := filepath.Join(tmpdir, "docker-compose.yml")
	err = ioutil.WriteFile(fp, []byte(`
version: "3"

services:
  db:
    build: .
    command: ./entrypoint.sh
    image: docker.io/tonistiigi/db
  webapp:
    build:
      dockerfile: Dockerfile.webapp
      args:
        buildno: 1
`), 0600)
	require.NoError(t, err)

	ctx := context.TODO()

	m, err := ReadTargets(ctx, []string{fp}, []string{"default"}, nil)
	require.NoError(t, err)

	require.Equal(t, "Dockerfile.webapp", *m["webapp"].Dockerfile)
	require.Equal(t, ".", *m["webapp"].Context)
}
