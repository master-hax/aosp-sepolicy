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

var (
	flagDepTag = dependencyTag{name: "flags"}
)

func init() {
	android.RegisterModuleType("se_flags", flagsFactory)
}

type flagsProperties struct {
	// List of flags to be passed to M4 macro.
	Flags []string
}

type flagsModule struct {
	android.ModuleBase

	properties flagsProperties
	flagMacros []string
}

// se_flags module defines a list of flags to be passed to M4 macro when compiling sepolicy files.
func flagsFactory() android.Module {
	f := &flagsModule{}
	f.AddProperties(&f.properties)
	android.InitAndroidModule(f)
	return f
}

func (f *flagsModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	f.flagMacros = []string{}
	for _, flag := range android.SortedUniqueStrings(f.properties.Flags) {
		if val, ok := ctx.Config().GetBuildFlag(flag); ok {
			f.flagMacros = append(f.flagMacros, "-D target_flag_"+flag+"="+val)
		}
	}
}

func addFlagsDependency(ctx android.BottomUpMutatorContext) {
	ctx.AddDependency(ctx.Module(), flagDepTag, "se_flags")
}

// m4FlagMacroDefinitions returns a list of M4's -D parameters to guard te files and contexts files.
func m4FlagMacroDefinitions(ctx android.ModuleContext) []string {
	return ctx.GetDirectDepWithTag("se_flags", flagDepTag).(*flagsModule).flagMacros
}
