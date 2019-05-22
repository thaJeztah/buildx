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
package commands

import (
	"fmt"
	"os"

	"github.com/containerd/containerd/images"
	"github.com/docker/buildx/util/imagetools"
	"github.com/docker/cli/cli"
	"github.com/docker/cli/cli/command"
	"github.com/moby/buildkit/util/appcontext"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
)

type inspectOptions struct {
	raw bool
}

func runInspect(dockerCli command.Cli, in inspectOptions, name string) error {
	ctx := appcontext.Context()

	r := imagetools.New(imagetools.Opt{
		Auth: dockerCli.ConfigFile(),
	})

	dt, desc, err := r.Get(ctx, name)
	if err != nil {
		return err
	}

	if in.raw {
		fmt.Printf("%s\n", dt)
		return nil
	}

	switch desc.MediaType {
	// case images.MediaTypeDockerSchema2Manifest, specs.MediaTypeImageManifest:
	// TODO: handle distribution manifest and schema1
	case images.MediaTypeDockerSchema2ManifestList, ocispec.MediaTypeImageIndex:
		imagetools.PrintManifestList(dt, desc, name, os.Stdout)
	default:
		fmt.Printf("%s\n", dt)
	}

	return nil
}

func inspectCmd(dockerCli command.Cli) *cobra.Command {
	var options inspectOptions

	cmd := &cobra.Command{
		Use:   "inspect [OPTIONS] NAME",
		Short: "Show details of image in the registry",
		Args:  cli.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInspect(dockerCli, options, args[0])
		},
	}

	flags := cmd.Flags()

	flags.BoolVar(&options.raw, "raw", false, "Show original JSON manifest")

	_ = flags

	return cmd
}
