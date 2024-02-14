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
	"maps"

	"android/soong/android"
)

var (
	flagsDepTag      = dependencyTag{name: "flags"}
	buildFlagsDepTag = dependencyTag{name: "build_flags"}
)

const (
	flagsCollectorModuleName = "se_flags_collector"
)

func init() {
	ctx := android.InitRegistrationContext
	ctx.RegisterModuleType("se_flags", flagsFactory)
	ctx.RegisterParallelSingletonModuleType("se_flags_collector", flagsCollectorFactory)
}

type flagsProperties struct {
	// List of build time flags for flag-guarding.
	Flags []string
}

type flagsModule struct {
	android.ModuleBase
	properties flagsProperties
}

// se_flags contains a list of build time flags for sepolicy.  Build time flags are defined under
// .scl files (e.g. build/release/build_flags.scl). By importing flags with se_flags modules,
// sepolicy rules can be guarded by `is_flag_enabled` / `is_flag_disabled` macro.
//
// For example, an Android.bp file could have:
//
//	se_flags {
//	    name: "se_avf_flags",
//	    flags: ["RELEASE_AVF_ENABLE_DEVICE_ASSIGNMENT"],
//	}
//
// And then one could flag-guard .te file rules:
//
//	is_flag_enabled(RELEASE_AVF_ENABLE_DEVICE_ASSIGNMENT, `
//	    type vfio_handler, domain, coredomain;
//	    binder_use(vfio_handler)
//	')
//
// or contexts entries:
//
//	is_flag_enabled(RELEASE_AVF_ENABLE_DEVICE_ASSIGNMENT, `
//	    android.system.virtualizationservice_internal.IVfioHandler u:object_r:vfio_handler_service:s0
//	')
func flagsFactory() android.Module {
	module := &flagsModule{}
	module.AddProperties(&module.properties)
	android.InitAndroidModule(module)
	return module
}

func (f *flagsModule) DepsMutator(ctx android.BottomUpMutatorContext) {
	// dep se_build_flag_collector -> se_flags
	ctx.AddReverseDependency(ctx.Module(), flagsDepTag, flagsCollectorModuleName)
}

func (f *flagsModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	// does nothing
}

type flagsCollectorModule struct {
	android.SingletonModuleBase
	buildFlags map[string]string
}

// se_flags_collector module collects all flags from all se_flags modules, and then converts them
// into build-time flags.  It will be used to generate M4 macros to flag-guard sepolicy.
func flagsCollectorFactory() android.SingletonModule {
	module := &flagsCollectorModule{}
	android.InitAndroidModule(module)
	android.AddLoadHook(module, func(ctx android.LoadHookContext) {
		if ctx.ModuleName() != flagsCollectorModuleName {
			ctx.PropertyErrorf("name", "module name must be %s", flagsCollectorModuleName)
		}
	})
	return module
}

func (f *flagsCollectorModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	var flags []string
	ctx.VisitDirectDepsWithTag(flagsDepTag, func(module android.Module) {
		if m, ok := module.(*flagsModule); ok {
			flags = append(flags, m.properties.Flags...)
		}
	})
	f.buildFlags = make(map[string]string)
	for _, flag := range android.SortedUniqueStrings(flags) {
		if val, ok := ctx.Config().GetBuildFlag(flag); ok {
			f.buildFlags[flag] = val
		}
	}
}

func (f *flagsCollectorModule) getBuildFlags() map[string]string {
	return f.buildFlags
}

func (f *flagsCollectorModule) GenerateSingletonBuildActions(ctx android.SingletonContext) {
	// does nothing
}

type buildFlagsModule interface {
	getBuildFlags() map[string]string
}

var _ buildFlagsModule = (*flagsCollectorModule)(nil)

type flaggableModuleProperties struct {
	// List of se_build_flag_collector modules to be passed to M4 macro.
	Build_flags []string
}

type flaggableModule interface {
	android.Module
	flagModuleBase() *flaggableModuleBase
	flagDeps(ctx android.BottomUpMutatorContext)
	getBuildFlags(ctx android.ModuleContext) map[string]string
}

type flaggableModuleBase struct {
	properties flaggableModuleProperties
}

func initFlaggableModule(m flaggableModule) {
	base := m.flagModuleBase()
	m.AddProperties(&base.properties)
}

func (f *flaggableModuleBase) flagModuleBase() *flaggableModuleBase {
	return f
}

func (f *flaggableModuleBase) flagDeps(ctx android.BottomUpMutatorContext) {
	ctx.AddDependency(ctx.Module(), buildFlagsDepTag, f.properties.Build_flags...)
}

// getBuildFlags returns a map from flag names to flag values.
func (f *flaggableModuleBase) getBuildFlags(ctx android.ModuleContext) map[string]string {
	ret := make(map[string]string)
	ctx.VisitDirectDepsWithTag(buildFlagsDepTag, func(m android.Module) {
		if i, ok := m.(buildFlagsModule); ok {
			maps.Copy(ret, i.getBuildFlags())
		}
	})
	return ret
}
