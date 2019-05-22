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
	"github.com/docker/cli/cli/compose/loader"
	composetypes "github.com/docker/cli/cli/compose/types"
)

func parseCompose(dt []byte) (*composetypes.Config, error) {
	parsed, err := loader.ParseYAML([]byte(dt))
	if err != nil {
		return nil, err
	}
	return loader.Load(composetypes.ConfigDetails{
		ConfigFiles: []composetypes.ConfigFile{
			{
				Config: parsed,
			},
		},
	})
}

func ParseCompose(dt []byte) (*Config, error) {
	cfg, err := parseCompose(dt)
	if err != nil {
		return nil, err
	}

	var c Config
	if len(cfg.Services) > 0 {
		c.Group = map[string]Group{}
		c.Target = map[string]Target{}

		var g Group

		for _, s := range cfg.Services {
			g.Targets = append(g.Targets, s.Name)
			var contextPathP *string
			if s.Build.Context != "" {
				contextPath := s.Build.Context
				contextPathP = &contextPath
			}
			var dockerfilePathP *string
			if s.Build.Dockerfile != "" {
				dockerfilePath := s.Build.Dockerfile
				dockerfilePathP = &dockerfilePath
			}
			t := Target{
				Context:    contextPathP,
				Dockerfile: dockerfilePathP,
				Labels:     s.Build.Labels,
				Args:       toMap(s.Build.Args),
				CacheFrom:  s.Build.CacheFrom,
				// TODO: add platforms
			}
			if s.Build.Target != "" {
				target := s.Build.Target
				t.Target = &target
			}
			if s.Image != "" {
				t.Tags = []string{s.Image}
			}
			c.Target[s.Name] = t
		}
		c.Group["default"] = g

	}

	return &c, nil
}

func toMap(in composetypes.MappingWithEquals) map[string]string {
	m := map[string]string{}
	for k, v := range in {
		if v != nil {
			m[k] = *v
		}
	}
	return m
}
