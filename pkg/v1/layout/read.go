// Copyright 2019 The original author or authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package layout

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/v1/match"
	"os"
	"path/filepath"

	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// FromPath reads an OCI image layout at path and constructs a layout.Path.
func FromPath(path string) (Path, error) {
	// TODO: check oci-layout exists

	_, err := os.Stat(filepath.Join(path, "index.json"))
	if err != nil {
		return "", err
	}

	return Path(path), nil
}

var maxDepth = 10

func walk(idx v1.ImageIndex, matcher match.Matcher, depth int) (v1.Image, error) {
	if depth >= maxDepth {
		return nil, fmt.Errorf("max depth exceeded: %d", depth)
	}

	manifest, err := idx.IndexManifest()
	if err != nil {
		return nil, fmt.Errorf("reading manifest %s: %w", idx, err)
	}

	for _, m := range manifest.Manifests {
		if m.MediaType.IsIndex() {
			subIdx, err := idx.ImageIndex(m.Digest)
			if err != nil {
				return nil, fmt.Errorf("reading index %s: %w", m.Digest, err)
			}
			return walk(subIdx, matcher, depth+1)
		} else if m.MediaType.IsImage() {
			if matcher(m) {
				img, err := idx.Image(m.Digest)
				if err != nil {
					return nil, fmt.Errorf("reading image %s: %w", m.Digest, err)
				}
				return img, nil
			}
		}
	}
	return nil, nil
}

func FindImage(path Path, matcher match.Matcher) (v1.Image, error) {
	idx, err := path.ImageIndex()
	if err != nil {
		return nil, fmt.Errorf("reading image %s: %w", path, err)
	}

	depth := 0
	img, err := walk(idx, matcher, depth)
	if err != nil {
		return nil, fmt.Errorf("reading image %s: %w", path, err)
	}

	return img, nil
}
