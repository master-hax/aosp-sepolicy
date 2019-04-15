// Copyright (C) 2019 The Android Open Source Project
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
	"fmt"
	"github.com/google/blueprint"
	"github.com/google/blueprint/proptools"
	"io"
	"strings"
)

const (
	coreMode     = "core"
	recoveryMode = "recovery"
)

type selinuxContextsProperties struct {
	// Filenames under sepolicy directories, which will be used to generate contexts file.
	Srcs []string

	Product_variables struct {
		// debuggable is true for eng and userdebug builds, and can be used to turn on additional
		// debugging features that don't significantly impact runtime behavior.  userdebug builds
		// are used for dogfooding and performance testing, and should be as similar to user builds
		// as possible.
		Debuggable struct {
			Srcs []string
		}

		// address_sanitize can be used to turn on additional features for AddressSanitizer.
		Address_sanitize struct {
			Srcs []string
		}
	}

	// Whether reqd_mask directory is included to sepolicy directories or not.
	Reqd_mask *bool

	// Whether the comments in generated contexts file will be removed or not.
	Remove_comment *bool

	// Whether the result context file is sorted with fc_sort or not.
	Fc_sort *bool

	// Make this module available when building for recovery
	Recovery_available *bool

	InRecovery bool `blueprint:"mutated"`
}

type fileContextsProperties struct {
	Product_variables struct {
		// flatten_apex can be used to specify additional sources of file_contexts.
		// Apex paths, /system/apex/{apex_name}, will be amended to the paths of file_contexts
		// entries.
		Flatten_apex struct {
			Srcs []string
		}
	}
}

type selinuxContextsModule struct {
	android.ModuleBase

	properties             selinuxContextsProperties
	fileContextsProperties fileContextsProperties
	build                  func(ctx android.ModuleContext, inputs android.Paths)
	moduleType             string
	outputPath             android.ModuleGenPath
	installPath            android.OutputPath
}

var (
	reuseContextsDepTag = dependencyTag{name: "reuseContexts"}

	m4 = pctx.AndroidStaticRule("m4",
		blueprint.RuleParams{
			Command: "m4 --fatal-warnings -s $m4defs $in > $out",
		}, "m4defs")

	remove_comment = pctx.AndroidStaticRule("remove_comment",
		blueprint.RuleParams{
			Command: "sed -e 's/#.*$$//' -e '/^$$/d' $in > $out",
		})

	fc_sort = pctx.AndroidStaticRule("fc_sort",
		blueprint.RuleParams{
			Command:     "$fc_sort -i $in -o $out",
			CommandDeps: []string{"$fc_sort"},
		})

	build_flattened_apex_file_contexts = pctx.AndroidStaticRule("apex_flattened_file_contexts",
		blueprint.RuleParams{
			Command: "awk '/object_r/{printf(\"${apex_path}%s\\n\",$$0)}' $in > $out",
		}, "apex_path")
)

func init() {
	pctx.HostBinToolVariable("fc_sort", "fc_sort")

	android.RegisterModuleType("file_contexts", fileFactory)
	android.RegisterModuleType("hwservice_contexts", hwServiceFactory)
	android.RegisterModuleType("property_contexts", propertyFactory)
	android.RegisterModuleType("service_contexts", serviceFactory)

	android.PreDepsMutators(func(ctx android.RegisterMutatorsContext) {
		ctx.BottomUp("selinux_contexts", selinuxContextsMutator).Parallel()
	})
}

func (m *selinuxContextsModule) inRecovery() bool {
	return m.properties.InRecovery || m.ModuleBase.InstallInRecovery()
}

func (m *selinuxContextsModule) onlyInRecovery() bool {
	return m.ModuleBase.InstallInRecovery()
}

func (m *selinuxContextsModule) InstallInRecovery() bool {
	return m.inRecovery()
}

func (m *selinuxContextsModule) policySrcs(config android.Config) []string {
	srcs := m.properties.Srcs

	if config.Debuggable() {
		srcs = append(srcs, m.properties.Product_variables.Debuggable.Srcs...)
	}

	for _, sanitize := range config.SanitizeDevice() {
		if sanitize == "address" {
			srcs = append(srcs, m.properties.Product_variables.Address_sanitize.Srcs...)
			break
		}
	}

	return srcs
}

func (m *selinuxContextsModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	if m.InstallInRecovery() {
		// Workaround for installing context files at the root of the recovery partition
		m.installPath = android.PathForOutput(ctx,
			"target", "product", ctx.Config().DeviceName(), "recovery", "root")
	} else {
		m.installPath = android.PathForModuleInstall(ctx, "etc", "selinux")
	}

	if m.inRecovery() && !m.onlyInRecovery() {
		reused := false

		ctx.VisitDirectDepsWithTag(reuseContextsDepTag, func(dep android.Module) {
			if reused {
				return
			}

			reuseDeps, ok := dep.(*selinuxContextsModule)
			if !ok {
				ctx.ModuleErrorf("unknown dependency %q", ctx.OtherModuleName(dep))
			}

			m.outputPath = reuseDeps.outputPath
			ctx.InstallFile(m.installPath, ctx.ModuleName(), m.outputPath)
			reused = true
		})

		if reused {
			return
		}
	}

	var inputs android.Paths

	ctx.VisitDirectDepsWithTag(android.SourceDepTag, func(dep android.Module) {
		segroup, ok := dep.(*fileGroup)
		if !ok {
			ctx.ModuleErrorf("srcs dependency %q is not an selinux filegroup",
				ctx.OtherModuleName(dep))
			return
		}

		if ctx.ProductSpecific() {
			inputs = append(inputs, segroup.ProductPrivateSrcs()...)
		} else if ctx.SocSpecific() {
			inputs = append(inputs, segroup.SystemVendorSrcs()...)
			inputs = append(inputs, segroup.VendorSrcs()...)
		} else if ctx.DeviceSpecific() {
			inputs = append(inputs, segroup.OdmSrcs()...)
		} else {
			inputs = append(inputs, segroup.SystemPrivateSrcs()...)
			inputs = append(inputs, segroup.SystemExtPrivateSrcs()...)

			if ctx.Config().ProductCompatibleProperty() {
				inputs = append(inputs, segroup.SystemPublicSrcs()...)
			}
		}

		if proptools.Bool(m.properties.Reqd_mask) {
			inputs = append(inputs, segroup.SystemReqdMaskSrcs()...)
		}
	})

	for _, src := range m.policySrcs(ctx.Config()) {
		if android.SrcIsModule(src) == "" {
			inputs = append(inputs, android.PathsForModuleSrcExcludes(ctx, []string{src}, nil)...)
		}
	}

	m.build(ctx, inputs)
}

func (m *selinuxContextsModule) DepsMutator(ctx android.BottomUpMutatorContext) {
	if m.inRecovery() && !m.onlyInRecovery() {
		ctx.AddFarVariationDependencies([]blueprint.Variation{
			{Mutator: "selinux_contexts", Variation: "core"},
		}, reuseContextsDepTag, m.Name())
	} else {
		android.ExtractSourcesDeps(ctx, m.policySrcs(ctx.Config()))
	}
}

func newModule() *selinuxContextsModule {
	m := &selinuxContextsModule{}
	m.AddProperties(
		&m.properties,
	)
	android.InitAndroidArchModule(m, android.DeviceSupported, android.MultilibCommon)
	return m
}

func (m *selinuxContextsModule) AndroidMk() android.AndroidMkData {
	return android.AndroidMkData{
		Custom: func(w io.Writer, name, prefix, moduleDir string, data android.AndroidMkData) {
			nameSuffix := ""
			if m.inRecovery() && !m.onlyInRecovery() {
				nameSuffix = ".recovery"
			}
			fmt.Fprintln(w, "\ninclude $(CLEAR_VARS)")
			fmt.Fprintln(w, "LOCAL_PATH :=", moduleDir)
			fmt.Fprintln(w, "LOCAL_MODULE :=", name+nameSuffix)
			fmt.Fprintln(w, "LOCAL_MODULE_CLASS := ETC")
			if m.Owner() != "" {
				fmt.Fprintln(w, "LOCAL_MODULE_OWNER :=", m.Owner())
			}
			fmt.Fprintln(w, "LOCAL_MODULE_TAGS := optional")
			fmt.Fprintln(w, "LOCAL_PREBUILT_MODULE_FILE :=", m.outputPath.String())
			fmt.Fprintln(w, "LOCAL_MODULE_PATH :=", "$(OUT_DIR)/"+m.installPath.RelPathString())
			fmt.Fprintln(w, "LOCAL_INSTALLED_MODULE_STEM :=", name)
			fmt.Fprintln(w, "include $(BUILD_PREBUILT)")
		},
	}
}

func selinuxContextsMutator(ctx android.BottomUpMutatorContext) {
	m, ok := ctx.Module().(*selinuxContextsModule)
	if !ok {
		return
	}

	var coreVariantNeeded bool = true
	var recoveryVariantNeeded bool = false
	if proptools.Bool(m.properties.Recovery_available) {
		recoveryVariantNeeded = true
	}

	if m.ModuleBase.InstallInRecovery() {
		recoveryVariantNeeded = true
		coreVariantNeeded = false
	}

	var variants []string
	if coreVariantNeeded {
		variants = append(variants, coreMode)
	}
	if recoveryVariantNeeded {
		variants = append(variants, recoveryMode)
	}
	mod := ctx.CreateVariations(variants...)
	for i, v := range variants {
		if v == recoveryMode {
			m := mod[i].(*selinuxContextsModule)
			m.properties.InRecovery = true
		}
	}
}

func (m *selinuxContextsModule) buildGeneralContexts(ctx android.ModuleContext, inputs android.Paths) {
	m.outputPath = android.PathForModuleGen(ctx, ctx.ModuleName()+"_m4out")

	ctx.Build(pctx, android.BuildParams{
		Rule:        m4,
		Inputs:      inputs,
		Output:      m.outputPath,
		Description: "generate " + m.moduleType + " to " + m.outputPath.String(),
		Args: map[string]string{
			"m4defs": android.JoinWithPrefix(ctx.DeviceConfig().SepolicyM4Defs(), "-D"),
		},
	})

	if proptools.Bool(m.properties.Remove_comment) {
		remove_comment_output := android.PathForModuleGen(ctx, ctx.ModuleName()+"_remove_comment")

		ctx.Build(pctx, android.BuildParams{
			Rule:        remove_comment,
			Inputs:      android.Paths{m.outputPath},
			Output:      remove_comment_output,
			Description: "remove_comment of " + m.outputPath.String(),
		})

		m.outputPath = remove_comment_output
	}

	if proptools.Bool(m.properties.Fc_sort) {
		sorted_output := android.PathForModuleGen(ctx, ctx.ModuleName()+"_sorted")

		ctx.Build(pctx, android.BuildParams{
			Rule:        fc_sort,
			Inputs:      android.Paths{m.outputPath},
			Output:      sorted_output,
			Description: "fc_sort of " + m.outputPath.String(),
		})

		m.outputPath = sorted_output
	}

	ctx.InstallFile(m.installPath, ctx.ModuleName(), m.outputPath)
}

func (m *selinuxContextsModule) buildFileContexts(ctx android.ModuleContext, inputs android.Paths) {
	if m.properties.Fc_sort == nil {
		m.properties.Fc_sort = proptools.BoolPtr(true)
	}

	if ctx.Config().FlattenApex() {
		for _, src := range m.fileContextsProperties.Product_variables.Flatten_apex.Srcs {
			if m := android.SrcIsModule(src); m != "" {
				ctx.ModuleErrorf(
					"Module srcs dependency %q is not supported for flatten_apex.srcs", m)
				return
			}
			for _, path := range android.PathsForModuleSrcExcludes(ctx, []string{src}, nil) {
				out := android.PathForModuleGen(ctx, "flattened_apex", path.Rel())
				apex_path := "/system/apex/" + strings.Replace(
					strings.TrimSuffix(path.Base(), "-file_contexts"),
					".", "\\\\.", -1)

				ctx.Build(pctx, android.BuildParams{
					Rule:        build_flattened_apex_file_contexts,
					Inputs:      android.Paths{path},
					Output:      out,
					Description: "build flattened apex file contexts of " + path.String(),
					Args: map[string]string{
						"apex_path": apex_path,
					},
				})

				inputs = append(inputs, out)
			}
		}
	}

	m.buildGeneralContexts(ctx, inputs)
}

func fileFactory() android.Module {
	m := newModule()
	m.AddProperties(&m.fileContextsProperties)
	m.build = m.buildFileContexts
	m.moduleType = "file_contexts"
	return m
}

func (m *selinuxContextsModule) buildHwServiceContexts(ctx android.ModuleContext, inputs android.Paths) {
	if m.properties.Remove_comment == nil {
		m.properties.Remove_comment = proptools.BoolPtr(true)
	}

	m.buildGeneralContexts(ctx, inputs)
}

func hwServiceFactory() android.Module {
	m := newModule()
	m.build = m.buildHwServiceContexts
	m.moduleType = "hwservice_contexts"
	return m
}

func propertyFactory() android.Module {
	m := newModule()
	m.build = m.buildGeneralContexts
	m.moduleType = "property_contexts"
	return m
}

func serviceFactory() android.Module {
	m := newModule()
	m.build = m.buildGeneralContexts
	m.moduleType = "service_contexts"
	return m
}
