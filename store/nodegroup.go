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
	"fmt"

	"github.com/containerd/containerd/platforms"
	"github.com/docker/buildx/util/platformutil"
	specs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type NodeGroup struct {
	Name   string
	Driver string
	Nodes  []Node
}

type Node struct {
	Name      string
	Endpoint  string
	Platforms []specs.Platform
}

func (ng *NodeGroup) Leave(name string) error {
	i := ng.findNode(name)
	if i == -1 {
		return errors.Errorf("node %q not found for %s", name, ng.Name)
	}
	if len(ng.Nodes) == 1 {
		return errors.Errorf("can not leave last node, do you want to rm instance instead?")
	}
	ng.Nodes = append(ng.Nodes[:i], ng.Nodes[i+1:]...)
	return nil
}

func (ng *NodeGroup) Update(name, endpoint string, platforms []string, endpointsSet bool, actionAppend bool) error {
	i := ng.findNode(name)
	if i == -1 && !actionAppend {
		if len(ng.Nodes) > 0 {
			return errors.Errorf("node %s not found, did you mean to append?", name)
		}
		ng.Nodes = nil
	}

	pp, err := platformutil.Parse(platforms)
	if err != nil {
		return err
	}

	if i != -1 {
		n := ng.Nodes[i]
		if endpointsSet {
			n.Endpoint = endpoint
		}
		if len(platforms) > 0 {
			n.Platforms = pp
		}
		ng.Nodes[i] = n
		if err := ng.validateDuplicates(endpoint, i); err != nil {
			return err
		}
		return nil
	}

	if name == "" {
		name = ng.nextNodeName()
	}

	name, err = ValidateName(name)
	if err != nil {
		return err
	}

	n := Node{
		Name:      name,
		Endpoint:  endpoint,
		Platforms: pp,
	}
	ng.Nodes = append(ng.Nodes, n)

	if err := ng.validateDuplicates(endpoint, len(ng.Nodes)-1); err != nil {
		return err
	}
	return nil
}

func (ng *NodeGroup) validateDuplicates(ep string, idx int) error {
	i := 0
	for _, n := range ng.Nodes {
		if n.Endpoint == ep {
			i++
		}
	}
	if i > 1 {
		return errors.Errorf("invalid duplicate endpoint %s", ep)
	}

	m := map[string]struct{}{}
	for _, p := range ng.Nodes[idx].Platforms {
		m[platforms.Format(p)] = struct{}{}
	}

	for i := range ng.Nodes {
		if i == idx {
			continue
		}
		ng.Nodes[i].Platforms = filterPlatforms(ng.Nodes[i].Platforms, m)
	}

	return nil
}

func (ng *NodeGroup) findNode(name string) int {
	for i, n := range ng.Nodes {
		if n.Name == name {
			return i
		}
	}
	return -1
}

func (ng *NodeGroup) nextNodeName() string {
	i := 0
	for {
		name := fmt.Sprintf("%s%d", ng.Name, i)
		if ii := ng.findNode(name); ii != -1 {
			i++
			continue
		}
		return name
	}
}

func filterPlatforms(in []specs.Platform, m map[string]struct{}) []specs.Platform {
	out := make([]specs.Platform, 0, len(in))
	for _, p := range in {
		if _, ok := m[platforms.Format(p)]; !ok {
			out = append(out, p)
		}
	}
	return out
}
