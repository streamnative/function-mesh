// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package cmdutils

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/ghodss/yaml"
	"github.com/spf13/pflag"
)

// region OutputConfig

// OutputConfig represents an output configuration
type OutputConfig struct {
	// the output format (Table, Plain, Json, Yaml)
	Format string
}

// AddTo registers the output flagset into a group
func (c *OutputConfig) AddTo(group *NamedFlagSetGroup) {
	group.InFlagSet("Output", func(flags *pflag.FlagSet) {
		flags.StringVarP(
			&c.Format,
			"output",
			"o",
			string(TextOutputFormat),
			"The output format (text,json,yaml)")
	})
}

// WriteOutput writes output based on the configured output format and on available content
func (c *OutputConfig) WriteOutput(w io.Writer, f OutputNegotiable) error {
	ow := f.Negotiate(OutputFormat(c.Format))
	if ow == nil {
		return fmt.Errorf("unsupported output format: %s", c.Format)
	}
	err := ow.WriteTo(w)
	if err != nil {
		return fmt.Errorf("output error: %v", err)
	}
	return nil
}

// OutputNegotiable facilitates content negotiation
type OutputNegotiable interface {
	// Negotiate produces an OutputWritable for the given output format, or nil if the format isn't supported
	Negotiate(format OutputFormat) OutputWritable
}

type OutputNegotiableFunc func(format OutputFormat) OutputWritable

func (f OutputNegotiableFunc) Negotiate(format OutputFormat) OutputWritable {
	return f(format)
}

// endregion

// region OutputFormat

// OutputFormat represents a user-configured output format
type OutputFormat string

const (
	TextOutputFormat OutputFormat = "text"
	JSONOutputFormat OutputFormat = "json"
	YAMLOutputFormat OutputFormat = "yaml"
)

func (fmt OutputFormat) String() string {
	return string(fmt)
}

// endregion

// region OutputWritable

// OutputWritable indicates an object that is writable to a given io.Writer
type OutputWritable interface {
	WriteTo(w io.Writer) error
}

// OutputWritableFunc adapts a function as an OutputWritable
type OutputWritableFunc func(w io.Writer) error

func (f OutputWritableFunc) WriteTo(w io.Writer) error {
	return f(w)
}

// byteOutputFunc adapts a function which produces bytes as an OutputWritable
type byteOutputFunc func() ([]byte, error)

func (f byteOutputFunc) WriteTo(w io.Writer) error {
	b, err := f()
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

// endregion

// region OutputContent
var _ OutputNegotiable = &OutputContent{}

// OutputContent adapts various Go types to output format(s)
type OutputContent struct {
	text OutputWritable
	obj  func() interface{}
}

func NewOutputContent() *OutputContent {
	return &OutputContent{}
}

// WithText produces the given formatted string
func (o *OutputContent) WithText(format string, i ...interface{}) *OutputContent {
	o.text = OutputWritableFunc(func(w io.Writer) error { _, err := fmt.Fprintf(w, format, i...); return err })
	return o
}

func (o *OutputContent) WithTextFunc(f func(w io.Writer) error) *OutputContent {
	o.text = OutputWritableFunc(f)
	return o
}

func (o *OutputContent) WithObject(obj interface{}) *OutputContent {
	o.obj = func() interface{} { return obj }
	return o
}

func (o *OutputContent) WithObjectFunc(f func() interface{}) *OutputContent {
	o.obj = f
	return o
}

// Negotiate produces an OutputWritable based on available content
func (o OutputContent) Negotiate(format OutputFormat) OutputWritable {
	switch format {
	case TextOutputFormat:
		if o.text != nil {
			return o.text
		}
		if o.obj != nil {
			// fallback to a JSON representation
			return byteOutputFunc(func() ([]byte, error) {
				return json.MarshalIndent(o.obj(), "", "  ")
			})
		}
	case JSONOutputFormat:
		if o.obj != nil {
			return byteOutputFunc(func() ([]byte, error) {
				return json.MarshalIndent(o.obj(), "", "  ")
			})
		}
	case YAMLOutputFormat:
		if o.obj != nil {
			return byteOutputFunc(func() ([]byte, error) {
				return yaml.Marshal(o.obj())
			})
		}
	default:
		return nil
	}
	return nil
}

// endregion
