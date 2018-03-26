// Copyright 2018 Google Inc. All rights reserved.
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
	"path/filepath"
)

func init() {
	android.RegisterModuleType("se_filegroup", FileGroupFactory)
}

func FileGroupFactory() android.Module {
	module := &fileGroup{}
	module.AddProperties(&module.properties)
	android.InitAndroidModule(module)
	return module
}

type fileGroupProperties struct {
	// list of source file suffixes used to collect selinux policy files.
	// Source files will be looked up in the following local directories:
	// system/sepolicy/{public, private, vendor, reqd_mask}
	// and directories specified by following config variables:
	// BOARD_SEPOLICY_DIRS, BOARD_ODM_SEPOLICY_DIRS
	// BOARD_PLAT_PUBLIC_SEPOLICY_DIR, BOARD_PLAT_PRIVATE_SEPOLICY_DIR
	Srcs    []string
}

type fileGroup struct {
	android.ModuleBase
	properties fileGroupProperties
	srcs       android.Paths
}

var _ android.SourceFileProducer = (*fileGroup)(nil)

func (fg *fileGroup) Srcs() android.Paths {
	return fg.srcs
}

func (fg *fileGroup) DepsMutator(ctx android.BottomUpMutatorContext) {}

func (fg *fileGroup) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	sepolicyDirs := []string {
		filepath.Join(ctx.ModuleDir(), "public"),
		filepath.Join(ctx.ModuleDir(), "private"),
		filepath.Join(ctx.ModuleDir(), "vendor"),
		filepath.Join(ctx.ModuleDir(), "redq_mask"),
		ctx.DeviceConfig().PlatPublicSepolicyDir(),
		ctx.DeviceConfig().PlatPrivateSepolicyDir(),
	}
	sepolicyDirs = append(sepolicyDirs, ctx.DeviceConfig().VendorSepolicyDirs()...)
	sepolicyDirs = append(sepolicyDirs, ctx.DeviceConfig().OdmSepolicyDirs()...)

	for _, f := range fg.properties.Srcs {
		for _, d := range sepolicyDirs {
			path := filepath.Join(d, f)
			files, err := ctx.GlobWithDeps(path, nil)
			if err != nil {
				ctx.ModuleErrorf("glob: %s", err.Error())
			}
			for _, f := range files {
				fg.srcs = append(fg.srcs, android.PathForSource(ctx, f))
			}
		}
	}
}
