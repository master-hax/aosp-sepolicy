// Copyright (C) 2023 The Android Open Source Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package selinux

import (
	"android/soong/android"
)

type flagsProperties struct {
	// List of flags to be passed to M4 macro.
	Flags []string
}

type flaggableModule interface {
	android.Module
	flagModuleBase() *flaggableModuleBase
	m4FlagMacroDefinitions(ctx android.ModuleContext) []string
}

type flaggableModuleBase struct {
	properties flagsProperties
}

func initFlaggableModule(m flaggableModule) {
	base := m.flagModuleBase()
	m.AddProperties(&base.properties)
}

func (f *flaggableModuleBase) flagModuleBase() *flaggableModuleBase {
	return f
}

// m4FlagMacroDefinitions returns a list of M4's -D parameters to guard te files and contexts files.
func (f *flaggableModuleBase) m4FlagMacroDefinitions(ctx android.ModuleContext) []string {
	flagMacros := []string{}
	for _, flag := range android.SortedUniqueStrings(f.properties.Flags) {
		if val, ok := ctx.Config().GetBuildFlag(flag); ok {
			flagMacros = append(flagMacros, "-D target_flag_"+flag+"="+val)
		}
	}
	return flagMacros
}
