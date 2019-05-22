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
package progress

import (
	"context"
	"strings"
	"sync"

	"github.com/moby/buildkit/client"
	"golang.org/x/sync/errgroup"
)

type MultiWriter struct {
	w     Writer
	eg    *errgroup.Group
	once  sync.Once
	ready chan struct{}
}

func (mw *MultiWriter) WithPrefix(pfx string, force bool) Writer {
	in := make(chan *client.SolveStatus)
	out := mw.w.Status()
	p := &prefixed{
		main: mw.w,
		in:   in,
	}
	mw.eg.Go(func() error {
		mw.once.Do(func() {
			close(mw.ready)
		})
		for {
			select {
			case v, ok := <-in:
				if ok {
					if force {
						for _, v := range v.Vertexes {
							v.Name = addPrefix(pfx, v.Name)
						}
					}
					out <- v
				} else {
					return nil
				}
			case <-mw.Done():
				return mw.Err()
			}
		}
	})
	return p
}

func (mw *MultiWriter) Done() <-chan struct{} {
	return mw.w.Done()
}

func (mw *MultiWriter) Err() error {
	return mw.w.Err()
}

func (mw *MultiWriter) Status() chan *client.SolveStatus {
	return nil
}

type prefixed struct {
	main Writer
	in   chan *client.SolveStatus
}

func (p *prefixed) Done() <-chan struct{} {
	return p.main.Done()
}

func (p *prefixed) Err() error {
	return p.main.Err()
}

func (p *prefixed) Status() chan *client.SolveStatus {
	return p.in
}

func NewMultiWriter(pw Writer) *MultiWriter {
	if pw == nil {
		return nil
	}
	eg, _ := errgroup.WithContext(context.TODO())

	ready := make(chan struct{})

	go func() {
		<-ready
		eg.Wait()
		close(pw.Status())
	}()

	return &MultiWriter{
		w:     pw,
		eg:    eg,
		ready: ready,
	}
}

func addPrefix(pfx, name string) string {
	if strings.HasPrefix(name, "[") {
		return "[" + pfx + " " + name[1:]
	}
	return "[" + pfx + "] " + name
}
