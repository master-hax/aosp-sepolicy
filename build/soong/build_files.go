// Copyright 2021 The Android Open Source Project
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
	"path/filepath"
	"strings"
)

func init() {
	android.RegisterModuleType("se_build_files", buildFilesFactory)
}

func buildFilesFactory() android.Module {
	module := &buildFiles{}
	module.AddProperties(&module.properties)
	android.InitAndroidModule(module)
	return module
}

type buildFilesProperties struct {
	// list of source file suffixes used to collect selinux policy files.
	// Source files will be looked up in the following local directories:
	// system/sepolicy/{public, private, vendor, reqd_mask}
	// and directories specified by following config variables:
	// BOARD_SEPOLICY_DIRS, BOARD_ODM_SEPOLICY_DIRS
	// BOARD_PLAT_PUBLIC_SEPOLICY_DIR, BOARD_PLAT_PRIVATE_SEPOLICY_DIR
	Srcs []string
}

type buildFiles struct {
	android.ModuleBase
	properties buildFilesProperties

	srcs map[string]android.Paths
}

var _ android.OutputFileProducer = (*buildFiles)(nil)

func (b *buildFiles) findSrcsInDirs(ctx android.ModuleContext, dirs ...string) android.Paths {
	result := android.Paths{}
	for _, file := range b.properties.Srcs {
		for _, dir := range dirs {
			path := filepath.Join(dir, file)
			files, err := ctx.GlobWithDeps(path, nil)
			if err != nil {
				ctx.ModuleErrorf("glob: %s", err.Error())
			}
			for _, f := range files {
				result = append(result, android.PathForSource(ctx, f))
			}
		}
	}
	return result
}

func (b *buildFiles) DepsMutator(ctx android.BottomUpMutatorContext) {
	// do nothing
}

func (b *buildFiles) OutputFiles(tag string) (android.Paths, error) {
	if paths, ok := b.srcs[tag]; ok {
		return paths, nil
	}

	return nil, fmt.Errorf("unknown tag %q. Supported tags are: %q", tag, strings.Join(android.SortedStringKeys(b.srcs), " "))
}

func (b *buildFiles) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	systemPublicDir := filepath.Join(ctx.ModuleDir(), "public")
	systemPrivateDir := filepath.Join(ctx.ModuleDir(), "private")
	systemExtPublicDirs := ctx.DeviceConfig().SystemExtPublicSepolicyDirs()
	systemExtPrivateDirs := ctx.DeviceConfig().SystemExtPrivateSepolicyDirs()
	productPublicDirs := ctx.Config().ProductPublicSepolicyDirs()
	productPrivateDirs := ctx.Config().ProductPrivateSepolicyDirs()
	reqdMaskDir := filepath.Join(ctx.ModuleDir(), "reqd_mask")

	b.srcs = make(map[string]android.Paths)
	b.srcs[".reqd_mask"] = b.findSrcsInDirs(ctx, reqdMaskDir)

	systemBuildDirs := []string{systemPublicDir, systemPrivateDir}
	systemPublicBuildDirs := []string{systemPublicDir}
	b.srcs[".plat"] = b.findSrcsInDirs(ctx, systemBuildDirs...)
	b.srcs[".plat_public"] = b.findSrcsInDirs(ctx, append(systemPublicBuildDirs, reqdMaskDir)...)

	systemExtBuildDirs := append(systemBuildDirs, append(systemExtPublicDirs, systemExtPrivateDirs...)...)
	systemExtPublicBuildDirs := append(systemPublicBuildDirs, systemExtPublicDirs...)
	b.srcs[".system_ext"] = b.findSrcsInDirs(ctx, systemExtBuildDirs...)
	b.srcs[".system_ext_public"] = b.findSrcsInDirs(ctx, append(systemExtPublicBuildDirs, reqdMaskDir)...)

	productBuildDirs := append(systemExtBuildDirs, append(productPublicDirs, productPrivateDirs...)...)
	productPublicBuildDirs := append(systemExtPublicBuildDirs, productPublicDirs...)
	b.srcs[".product"] = b.findSrcsInDirs(ctx, productBuildDirs...)
	b.srcs[".product_public"] = b.findSrcsInDirs(ctx, append(productPublicBuildDirs, reqdMaskDir)...)
}
