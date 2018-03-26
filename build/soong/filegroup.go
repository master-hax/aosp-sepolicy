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
	Srcs    []string
	SrcDirs []string
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
	srcDirsMap := map[string][]string{
		"BOARD_PLAT_PUBLIC_SEPOLICY_DIR": []string{
			ctx.DeviceConfig().PlatPrivateSepolicyDir()},
		"BOARD_PLAT_PRIVATE_SEPOLICY_DIR": []string{
			ctx.DeviceConfig().PlatPrivateSepolicyDir()},
		"BOARD_VENDOR_SEPOLICY_DIRS": ctx.DeviceConfig().VendorSepolicyDirs(),
		"BOARD_ODM_SEPOLICY_DIRS":    ctx.DeviceConfig().OdmSepolicyDirs(),
	}

	var allSrcDirs []string
	for _, d := range fg.properties.SrcDirs {
		boardDirs, ok := srcDirsMap[d]
		if !ok {
			// Look up in module's local source directory.
			boardDirs = []string{filepath.Join(ctx.ModuleDir(), d)}
		}
		allSrcDirs = append(allSrcDirs, boardDirs...)
	}

	for _, f := range fg.properties.Srcs {
		for _, d := range allSrcDirs {
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
