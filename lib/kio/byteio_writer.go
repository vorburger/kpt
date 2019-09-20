// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kio

import (
	"io"

	"lib.kpt.dev/yaml"
)

// Writer writes ResourceNodes to bytes.
type ByteWriter struct {
	// Writer is where ResourceNodes are encoded.
	Writer io.Writer

	// KeepReaderAnnotations if set will keep the Reader specific annotations when writing
	// the Resources, otherwise they will be cleared.
	KeepReaderAnnotations bool

	// ClearAnnotations is a list of annotations to clear when writing the Resources.
	ClearAnnotations []string

	// Style is a style that is set on the Resource Node Document.
	Style yaml.Style
}

var _ Writer = ByteWriter{}

func (w ByteWriter) Write(nodes []*yaml.RNode) error {
	if err := sortNodes(nodes); err != nil {
		return err
	}

	encoder := yaml.NewEncoder(w.Writer)
	defer encoder.Close()
	for i := range nodes {

		// clean resources by removing annotations set by the Reader
		if !w.KeepReaderAnnotations {
			_, err := nodes[i].Pipe(yaml.ClearAnnotation(IndexAnnotation))
			if err != nil {
				return err
			}
		}
		for _, a := range w.ClearAnnotations {
			_, err := nodes[i].Pipe(yaml.ClearAnnotation(a))
			if err != nil {
				return err
			}
		}

		// TODO(pwittrock): factor this into a a common module for pruning empty values
		_, err := nodes[i].Pipe(yaml.Lookup("metadata"), yaml.FieldClearer{
			Name: "annotations", IfEmpty: true})
		if err != nil {
			return err
		}
		_, err = nodes[i].Pipe(yaml.FieldClearer{Name: "metadata", IfEmpty: true})
		if err != nil {
			return err
		}

		if w.Style != 0 {
			nodes[i].YNode().Style = w.Style
		}
		err = encoder.Encode(nodes[i].Document())
		if err != nil {
			return err
		}
	}
	return nil
}