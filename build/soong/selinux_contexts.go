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
	"path/filepath"
	"strconv"
	"strings"
)

const (
	platPrivatePolicyDir = "system/sepolicy/private"
	platPublicPolicyDir  = "system/sepolicy/public"
	platVendorPolicyDir  = "system/sepolicy/vendor"

	reqdMaskPolicyDir = "system/sepolicy/reqd_mask"

	coreMode     = "core"
	recoveryMode = "recovery"
)

type commonProperties struct {
	Owner *string
}

type selinuxContextsProperties struct {
	// Filenames under sepolicy directories, which will be used to generate contexts file.
	Filenames []string

	// debuggable is true for eng and userdebug builds, and can be used to turn on additional
	// debugging features that don't significantly impact runtime behavior.  userdebug builds
	// are used for dogfooding and performance testing, and should be as similar to user builds
	// as possible.
	Debuggable struct {
		Filenames []string
	}

	// sanitize can be used to turn on additional features for AddressSanitizer,
	// ThreadSanitizer, or UndefinedBehaviorSanitizer.
	Sanitize struct {
		Address struct {
			Filenames []string
		}
	}

	// Whether reqd_mask directory is included to sepolicy directories or not.
	Reqd_mask *bool

	// Whether the comments in generated contexts file will be removed or not.
	Remove_comment *bool

	// Make this module available when building for recovery
	Recovery_available *bool

	InRecovery bool `blueprint:"mutated"`
}

type fileContextsProperties struct {
	// apex.flatten can be used to specify additional sources of file_contexts.
	// Apex paths, /system/apex/{apex_name}, will be amended to the paths of file_contexts
	// entries.
	Apex struct {
		Flatten struct {
			Srcs []string
		}
	}
}

type selinuxContextsModule struct {
	android.ModuleBase

	commonProperties       commonProperties
	properties             selinuxContextsProperties
	fileContextsProperties fileContextsProperties
	build                  func(ctx android.ModuleContext, inputs android.Paths)
	outputPath             android.ModuleOutPath
	installPath            android.OutputPath
}

var (
	reuseContextsDep = dependencyTag{name: "reuseContexts"}

	file_contexts = pctx.AndroidStaticRule("file_contexts",
		blueprint.RuleParams{
			Command: "m4 --fatal-warnings -s $m4defs $in > $out.tmp && " +
				"if [ $remove_comment = true ]; then " +
				"sed -e 's/#.*$$$$//' -e '/^$$$$/d' $out.tmp > $out.tmp2 && " +
				"mv -f $out.tmp2 $out.tmp; " +
				"fi && $fc_sort -i $out.tmp -o $out && rm $out.tmp",
			CommandDeps: []string{"$fc_sort"},
		}, "m4defs", "remove_comment")

	build_flattened_apex_file_contexts = pctx.AndroidStaticRule("apex_flattened_file_contexts",
		blueprint.RuleParams{
			Command: "awk '/object_r/{printf(\"${apex_path}%s\\n\",$$0)}' $in > $out",
		}, "apex_path")

	hwservice_contexts = pctx.AndroidStaticRule("hwservice_contexts",
		blueprint.RuleParams{
			Command: "m4 --fatal-warnings -s $m4defs $in > $out.tmp && " +
				"if [ $remove_comment = true ]; then " +
				"sed -e 's/#.*$$$$//' -e '/^$$$$/d' $out.tmp > $out.tmp2 && " +
				"mv -f $out.tmp2 $out.tmp; " +
				"fi && mv -f $out.tmp $out",
		}, "m4defs", "remove_comment")

	property_contexts = pctx.AndroidStaticRule("property_contexts",
		blueprint.RuleParams{
			Command: "m4 --fatal-warnings -s $m4defs $in > $out.tmp && " +
				"if [ $remove_comment = true ]; then " +
				"sed -e 's/#.*$$$$//' -e '/^$$$$/d' $out.tmp > $out.tmp2 && " +
				"mv -f $out.tmp2 $out.tmp; " +
				"fi && mv -f $out.tmp $out",
		}, "m4defs", "remove_comment")

	service_contexts = pctx.AndroidStaticRule("service_contexts",
		blueprint.RuleParams{
			Command: "m4 --fatal-warnings -s $m4defs $in > $out.tmp && " +
				"if [ $remove_comment = true ]; then " +
				"sed -e 's/#.*$$$$//' -e '/^$$$$/d' $out.tmp > $out.tmp2 && " +
				"mv -f $out.tmp2 $out.tmp; " +
				"fi && mv -f $out.tmp $out",
		}, "m4defs", "remove_comment")
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

func (m *selinuxContextsModule) systemPolicyDirs(ctx android.ModuleContext) []string {
	ret := ctx.DeviceConfig().PlatPrivateSepolicyDirs()
	ret = append(ret, platPrivatePolicyDir)

	if ctx.Config().ProductCompatibleProperty() {
		ret = append(ret, platPublicPolicyDir)
	}

	return ret
}

func (m *selinuxContextsModule) vendorPolicyDirs(ctx android.ModuleContext) []string {
	ret := ctx.DeviceConfig().VendorSepolicyDirs()
	ret = append(ret, platVendorPolicyDir)

	return ret
}

func (m *selinuxContextsModule) productPolicyDirs(ctx android.ModuleContext) []string {
	return ctx.Config().ProductPrivatePolicyDirs()
}

func (m *selinuxContextsModule) odmPolicyDirs(ctx android.ModuleContext) []string {
	return ctx.DeviceConfig().OdmSepolicyDirs()
}

func (m *selinuxContextsModule) policyDirs(ctx android.ModuleContext) []string {
	var ret []string

	if ctx.SocSpecific() {
		ret = m.vendorPolicyDirs(ctx)
	} else if ctx.ProductSpecific() {
		ret = m.productPolicyDirs(ctx)
	} else if ctx.DeviceSpecific() {
		ret = m.odmPolicyDirs(ctx)
	} else {
		ret = m.systemPolicyDirs(ctx)
	}

	if proptools.Bool(m.properties.Reqd_mask) {
		ret = append(ret, reqdMaskPolicyDir)
	}

	return ret
}

func (m *selinuxContextsModule) policyFiles(ctx android.ModuleContext) []string {
	files := m.properties.Filenames

	if ctx.Config().Debuggable() {
		files = append(files, m.properties.Debuggable.Filenames...)
	}

	for _, sanitize := range ctx.Config().SanitizeDevice() {
		if sanitize == "address" {
			files = append(files, m.properties.Sanitize.Address.Filenames...)
			break
		}
	}

	return files
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

func (m *selinuxContextsModule) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	m.outputPath = android.PathForModuleOut(ctx, ctx.ModuleName())
	if m.InstallInRecovery() {
		// Workaround for installing context files at the root of the recovery partition
		m.installPath = android.PathForOutput(ctx,
			"target", "product", ctx.Config().DeviceName(), "recovery", "root")
	} else {
		m.installPath = android.PathForModuleInstall(ctx, "etc", "selinux")
	}

	if m.inRecovery() && !m.onlyInRecovery() {
		reused := false

		ctx.VisitDirectDeps(func(dep android.Module) {
			if !reused && ctx.OtherModuleDependencyTag(dep) == reuseContextsDep {
				ctx.Build(pctx, android.BuildParams{
					Rule:   android.Cp,
					Input:  dep.(*selinuxContextsModule).outputPath,
					Output: m.outputPath,
				})

				ctx.InstallFile(m.installPath, ctx.ModuleName(), m.outputPath)

				reused = true
				return
			}
		})

		if reused {
			return
		}
	}

	var inputs android.Paths

	dirs := m.policyDirs(ctx)
	files := m.policyFiles(ctx)

	for _, dir := range dirs {
		for _, file := range files {
			path := android.ExistentPathForSource(ctx, dir, file)
			if path.Valid() {
				inputs = append(inputs, path.Path())
			}
		}
	}

	m.build(ctx, inputs)
}

func (m *selinuxContextsModule) DepsMutator(ctx android.BottomUpMutatorContext) {
	if m.inRecovery() && !m.onlyInRecovery() {
		ctx.AddFarVariationDependencies([]blueprint.Variation{
			{Mutator: "selinux_contexts", Variation: "core"},
		}, reuseContextsDep, m.Name())
	}
}

func newModule() *selinuxContextsModule {
	m := &selinuxContextsModule{}
	m.AddProperties(
		&m.properties,
		&m.commonProperties,
	)
	android.InitAndroidArchModule(m, android.DeviceSupported, android.MultilibFirst)
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
			if m.commonProperties.Owner != nil {
				fmt.Fprintln(w, "LOCAL_MODULE_OWNER :=", *m.commonProperties.Owner)
			}
			fmt.Fprintln(w, "LOCAL_MODULE_TAGS := optional")
			fmt.Fprintln(w, "LOCAL_PREBUILT_MODULE_FILE :=", m.outputPath.String())
			fmt.Fprintln(w, "LOCAL_MODULE_PATH :=", "$(OUT_DIR)/"+m.installPath.RelPathString())
			fmt.Fprintln(w, "LOCAL_INSTALLED_MODULE_STEM :=", m.outputPath.Base())
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

func (m *selinuxContextsModule) buildFileContexts(ctx android.ModuleContext, inputs android.Paths) {
	if ctx.Config().FlattenApex() {
		for _, src := range m.fileContextsProperties.Apex.Flatten.Srcs {
			for _, path := range ctx.GlobFiles(filepath.Join(ctx.ModuleDir(), src), nil) {
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

	ctx.Build(pctx, android.BuildParams{
		Rule:        file_contexts,
		Inputs:      inputs,
		Output:      m.outputPath,
		Description: "generate file_contexts to " + m.outputPath.String(),
		Args: map[string]string{
			"m4defs":         android.JoinWithPrefix(ctx.DeviceConfig().SepolicyM4Defs(), "-D"),
			"remove_comment": strconv.FormatBool(proptools.Bool(m.properties.Remove_comment)),
		},
	})

	ctx.InstallFile(m.installPath, ctx.ModuleName(), m.outputPath)
}

func fileFactory() android.Module {
	m := newModule()
	m.AddProperties(&m.fileContextsProperties)
	m.build = m.buildFileContexts
	return m
}

func (m *selinuxContextsModule) buildHwServiceContexts(ctx android.ModuleContext, inputs android.Paths) {
	ctx.Build(pctx, android.BuildParams{
		Rule:        hwservice_contexts,
		Inputs:      inputs,
		Output:      m.outputPath,
		Description: "generate hwservice_contexts to " + m.outputPath.String(),
		Args: map[string]string{
			"m4defs":         android.JoinWithPrefix(ctx.DeviceConfig().SepolicyM4Defs(), "-D"),
			"remove_comment": strconv.FormatBool(proptools.Bool(m.properties.Remove_comment)),
		},
	})

	ctx.InstallFile(m.installPath, ctx.ModuleName(), m.outputPath)
}

func hwServiceFactory() android.Module {
	m := newModule()
	m.build = m.buildHwServiceContexts
	return m
}

func (m *selinuxContextsModule) buildPropertyContexts(ctx android.ModuleContext, inputs android.Paths) {
	ctx.Build(pctx, android.BuildParams{
		Rule:        property_contexts,
		Inputs:      inputs,
		Output:      m.outputPath,
		Description: "generate property_contexts to " + m.outputPath.String(),
		Args: map[string]string{
			"m4defs":         android.JoinWithPrefix(ctx.DeviceConfig().SepolicyM4Defs(), "-D"),
			"remove_comment": strconv.FormatBool(proptools.Bool(m.properties.Remove_comment)),
		},
	})

	ctx.InstallFile(m.installPath, ctx.ModuleName(), m.outputPath)
}

func propertyFactory() android.Module {
	m := newModule()
	m.build = m.buildPropertyContexts
	return m
}

func (m *selinuxContextsModule) buildServiceContexts(ctx android.ModuleContext, inputs android.Paths) {
	ctx.Build(pctx, android.BuildParams{
		Rule:        service_contexts,
		Inputs:      inputs,
		Output:      m.outputPath,
		Description: "generate service_contexts to " + m.outputPath.String(),
		Args: map[string]string{
			"m4defs":         android.JoinWithPrefix(ctx.DeviceConfig().SepolicyM4Defs(), "-D"),
			"remove_comment": strconv.FormatBool(proptools.Bool(m.properties.Remove_comment)),
		},
	})

	ctx.InstallFile(m.installPath, ctx.ModuleName(), m.outputPath)
}

func serviceFactory() android.Module {
	m := newModule()
	m.build = m.buildServiceContexts
	return m
}
