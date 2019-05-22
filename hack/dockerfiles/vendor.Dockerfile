#   Copyright The Buildx Authors.

#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at

#       http://www.apache.org/licenses/LICENSE-2.0

#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
# syntax = docker/dockerfile:1.0-experimental
FROM golang:1.12-alpine AS vendored
RUN  apk add --no-cache git rsync
WORKDIR /src
RUN --mount=target=/context \
  --mount=target=.,type=tmpfs,readwrite  \
  --mount=target=/go/pkg/mod,type=cache \
  rsync -a /context/. . && \
  go mod tidy && go mod vendor && \
  mkdir /out && cp -r go.mod go.sum vendor /out

FROM scratch AS update
COPY --from=vendored /out /out

FROM vendored AS validate
RUN --mount=target=/context \
  --mount=target=.,type=tmpfs,readwrite  \
  rsync -a /context/. . && \
  git add -A && \
  rm -rf vendor && \
  cp -rf /out/* . && \
  ./hack/validate-vendor check
