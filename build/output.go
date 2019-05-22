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
package build

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/containerd/console"
	"github.com/moby/buildkit/client"
	"github.com/pkg/errors"
)

func ParseOutputs(inp []string) ([]client.ExportEntry, error) {
	var outs []client.ExportEntry
	if len(inp) == 0 {
		return nil, nil
	}
	for _, s := range inp {
		csvReader := csv.NewReader(strings.NewReader(s))
		fields, err := csvReader.Read()
		if err != nil {
			return nil, err
		}

		out := client.ExportEntry{
			Attrs: map[string]string{},
		}
		if len(fields) == 1 && fields[0] == s && !strings.HasPrefix(s, "type=") {
			if s != "-" {
				outs = append(outs, client.ExportEntry{
					Type:      client.ExporterLocal,
					OutputDir: s,
				})
				continue
			}
			out = client.ExportEntry{
				Type: client.ExporterTar,
				Attrs: map[string]string{
					"dest": s,
				},
			}
		}

		if out.Type == "" {
			for _, field := range fields {
				parts := strings.SplitN(field, "=", 2)
				if len(parts) != 2 {
					return nil, errors.Errorf("invalid value %s", field)
				}
				key := strings.TrimSpace(strings.ToLower(parts[0]))
				value := parts[1]
				switch key {
				case "type":
					out.Type = value
				default:
					out.Attrs[key] = value
				}
			}
		}
		if out.Type == "" {
			return nil, errors.Errorf("type is required for output")
		}

		// handle client side
		switch out.Type {
		case client.ExporterLocal:
			dest, ok := out.Attrs["dest"]
			if !ok {
				return nil, errors.Errorf("dest is required for local output")
			}
			out.OutputDir = dest
			delete(out.Attrs, "dest")
		case client.ExporterOCI, client.ExporterDocker, client.ExporterTar:
			dest, ok := out.Attrs["dest"]
			if !ok {
				if out.Type != client.ExporterDocker {
					dest = "-"
				}
			}
			if dest == "-" {
				if _, err := console.ConsoleFromFile(os.Stdout); err == nil {
					return nil, errors.Errorf("output file is required for %s exporter. refusing to write to console", out.Type)
				}
				out.Output = os.Stdout
			} else if dest != "" {
				fi, err := os.Stat(dest)
				if err != nil && !os.IsNotExist(err) {
					return nil, errors.Wrapf(err, "invalid destination file: %s", dest)
				}
				if err == nil && fi.IsDir() {
					return nil, errors.Errorf("destination file %s is a directory", dest)
				}
				f, err := os.Create(dest)
				if err != nil {
					return nil, errors.Errorf("failed to open %s", err)
				}
				out.Output = f
			}
			delete(out.Attrs, "dest")
		case "registry":
			out.Type = client.ExporterImage
			out.Attrs["push"] = "true"
		}

		outs = append(outs, out)
	}
	return outs, nil
}
