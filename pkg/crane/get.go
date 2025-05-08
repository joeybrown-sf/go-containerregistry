// Copyright 2018 Google LLC All Rights Reserved.
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

package crane

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/layout"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

func getImage(r string, opt ...Option) (v1.Image, name.Reference, error) {
	o := makeOptions(opt...)
	ref, err := name.ParseReference(r, o.Name...)
	if err != nil {
		return nil, nil, fmt.Errorf("parsing reference %q: %w", r, err)
	}
	img, err := remote.Image(ref, o.Remote...)
	if err != nil {
		return nil, nil, fmt.Errorf("reading image %q: %w", ref, err)
	}
	return img, ref, nil
}

func getManifest(r string, opt ...Option) (*remote.Descriptor, error) {
	o := makeOptions(opt...)
	ref, err := name.ParseReference(r, o.Name...)
	if err != nil {
		return nil, fmt.Errorf("parsing reference %q: %w", r, err)
	}
	return remote.Get(ref, o.Remote...)
}

// Get calls remote.Get and returns an uninterpreted response.
func Get(r string, opt ...Option) (*remote.Descriptor, error) {
	return getManifest(r, opt...)
}

// Head performs a HEAD request for a manifest and returns a content descriptor
// based on the registry's response.
func Head(r string, opt ...Option) (*v1.Descriptor, error) {
	o := makeOptions(opt...)
	ref, err := name.ParseReference(r, o.Name...)
	if err != nil {
		return nil, err
	}
	return remote.Head(ref, o.Remote...)
}

func find(idx v1.ImageIndex, o Options) (v1.Image, error) {
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
			return find(subIdx, o)
		} else if m.MediaType.IsImage() {
			if (*m.Platform).Equals(*o.Platform) {
				img, err := idx.Image(m.Digest)
				if err != nil {
					return nil, fmt.Errorf("reading image %s: %w", m.Digest, err)
				}
				return img, nil
			}
		}
	}

	return nil, fmt.Errorf("cannot find image for platform %s", o.Platform)
}

func Read(p string, opt ...Option) (v1.Image, error) {
	o := makeOptions(opt...)
	if o.Platform == nil {
		platform := v1.Platform{
			Architecture: "amd64",
			OS:           "linux",
		}
		o.Platform = &platform
	}

	path, err := layout.FromPath(p)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", p, err)
	}

	idx, err := path.ImageIndex()
	if err != nil {
		return nil, fmt.Errorf("reading image %s: %w", idx, err)
	}

	img, err := find(idx, o)
	if err != nil {
		return nil, err
	}

	return img, nil
	//
	//for _, m := range idxManifest.Manifests {
	//
	//	if m.MediaType.IsIndex() {
	//		subIdx, err := idx.ImageIndex(m.Digest)
	//		_ = subIdx
	//		_ = err
	//	}
	//	if m.MediaType.IsImage() {
	//		x := 1
	//		_ = x
	//	}
	//
	//	platformImg, err := idx.Image(m.Digest)
	//	m, err := platformImg.Manifest()
	//
	//	_ = platformImg
	//	_ = err
	//	_ = m

	//if platformImg.Manifest().Platform == o.Platform {
	//
	//	if err != nil {
	//		return nil, fmt.Errorf("reading image at digest %s: %w", m.Digest, err)
	//	}
	//	return platformImg, nil
	//}
}
