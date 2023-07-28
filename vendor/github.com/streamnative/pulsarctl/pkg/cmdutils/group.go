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
	"fmt"
	"strings"
	"text/tabwriter"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagGrouping holds a superset of all flagsets for all commands
type FlagGrouping struct {
	groups map[*cobra.Command]*NamedFlagSetGroup
}

type namedFlagSet struct {
	name string
	fs   *pflag.FlagSet
}

// NamedFlagSetGroup holds a single group of flagsets
type NamedFlagSetGroup struct {
	list []namedFlagSet
}

// NewGrouping creates an instance of Grouping
func NewGrouping() *FlagGrouping {
	return &FlagGrouping{
		make(map[*cobra.Command]*NamedFlagSetGroup),
	}
}

// New creates a new group of flagsets for use with a subcommand
func (g *FlagGrouping) New(cmd *cobra.Command) *NamedFlagSetGroup {
	n := &NamedFlagSetGroup{}
	g.groups[cmd] = n
	return n
}

// InFlagSet returns new or existing named FlagSet in a group
func (n *NamedFlagSetGroup) InFlagSet(name string, cb func(*pflag.FlagSet)) {
	for _, nfs := range n.list {
		if nfs.name == name {
			cb(nfs.fs)
			return
		}
	}

	nfs := namedFlagSet{
		name: name,
		fs:   &pflag.FlagSet{},
	}
	cb(nfs.fs)
	n.list = append(n.list, nfs)
}

func (n *NamedFlagSetGroup) Add(name string, fs *pflag.FlagSet) {
	for _, nfs := range n.list {
		if nfs.name == name {
			nfs.fs = fs
			return
		}
	}

	nfs := namedFlagSet{
		name: name,
		fs:   fs,
	}
	n.list = append(n.list, nfs)
}

// AddTo mixes all flagsets in the given group to another flagset
func (n *NamedFlagSetGroup) AddTo(cmd *cobra.Command) {
	for _, nfs := range n.list {
		cmd.Flags().AddFlagSet(nfs.fs)
	}
}

// Usage is for use with (*cobra.Command).SetUsageFunc
func (g *FlagGrouping) Usage(cmd *cobra.Command) error {
	if cmd == nil {
		return fmt.Errorf("nil command")
	}

	group := g.groups[cmd]

	usage := []string{fmt.Sprintf("Usage: %s", cmd.UseLine())}

	if cmd.HasAvailableSubCommands() {
		usage = append(usage, "\nCommands:")
		buf := strings.Builder{}
		w := tabwriter.NewWriter(&buf, 10, 0, 3, ' ', 0)
		for _, subCommand := range cmd.Commands() {
			_, _ = fmt.Fprintf(w, "  %s\t%s\n", subCommand.Name(), subCommand.Short)
		}
		_ = w.Flush()
		usage = append(usage, strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")...)
	}

	if cmd.HasExample() {
		usage = append(usage, "\nExamples:")
		usage = append(usage, cmd.Example)
	}

	if len(cmd.Aliases) > 0 {
		usage = append(usage, "\nAliases: "+cmd.NameAndAliases())
	}

	if group != nil {
		for _, nfs := range group.list {
			usage = append(usage, fmt.Sprintf("\n%s flags:", nfs.name))
			usage = append(usage, strings.TrimRightFunc(nfs.fs.FlagUsages(), unicode.IsSpace))
		}
	}

	usage = append(usage, "\nCommon flags:")
	if len(cmd.PersistentFlags().FlagUsages()) != 0 {
		usage = append(usage, strings.TrimRightFunc(cmd.PersistentFlags().FlagUsages(), unicode.IsSpace))
	}
	if len(cmd.InheritedFlags().FlagUsages()) != 0 {
		usage = append(usage, strings.TrimRightFunc(cmd.InheritedFlags().FlagUsages(), unicode.IsSpace))
	}

	usage = append(usage,
		fmt.Sprintf("\nUse '%s [command] --help' for more information about a command.\n",
			cmd.CommandPath()))

	cmd.Println(strings.Join(usage, "\n"))

	return nil
}
