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
	"io"
)

func init() {
	android.RegisterModuleType("se_cil_compat_map_ext", cilCompatMapExtFactory)
}

// This file contains "se_cil_compat_map_ext" module type used to build and
// install device-specific extensions of sepolicy backwards compatibility mapping
// files. The core mapping files are built by "se_cil_compat_map" module type.

func expandProductSources(ctx android.ModuleContext, srcFiles []string) android.Paths {
	expandedSrcFiles := make(android.Paths, 0, len(srcFiles))
	for _, s := range srcFiles {
		if m := android.SrcIsModule(s); m != "" {
			module := ctx.GetDirectDepWithTag(m, android.SourceDepTag)
			if module == nil {
				// Error will have been handled by ExtractSourcesDeps
				continue
			}
			if fg, ok := module.(*fileGroup); ok {
				// Partner extensions to the compatibility mapping in must be located in
				// BOARD_PLAT_PRIVATE_SEPOLICY_DIR
				expandedSrcFiles = append(expandedSrcFiles, fg.SystemExtPrivateSrcs()...)
			} else {
				ctx.ModuleErrorf("srcs dependency %q is not an selinux filegroup", m)
			}
		}
	}
	return expandedSrcFiles
}

func cilCompatMapExtFactory() android.Module {
	c := &cilCompatMapExt{}
	c.AddProperties(&c.properties)
	android.InitAndroidModule(c)
	return c
}

type cilCompatMapExtProperties struct {
	// list of source (.cil) files used to build device-specific extension of
	// compatibility mapping file. srcs may reference the outputs of other
	// modules that produce source files like genrule or filegroup using the
	// syntax ":module". srcs has to be non-empty.
	Srcs []string
	// Name of the output file.
	Stem *string
}

type cilCompatMapExt struct {
	android.ModuleBase
	properties cilCompatMapExtProperties
	// (.intermediate) module output path as installation source.
	installSource android.Path
}

func (c *cilCompatMapExt) DepsMutator(ctx android.BottomUpMutatorContext) {
	android.ExtractSourcesDeps(ctx, c.properties.Srcs)
}

func (c *cilCompatMapExt) GenerateAndroidBuildActions(ctx android.ModuleContext) {
	srcFiles := expandProductSources(ctx, c.properties.Srcs)
	for _, src := range srcFiles {
		if src.Ext() != ".cil" {
			ctx.PropertyErrorf("%s has to be a .cil file.", src.String())
		}
	}

	out := android.PathForModuleGen(ctx, c.Name())
	if len(srcFiles) == 0 {
		ctx.Build(pctx, android.BuildParams{
			Rule:   android.Touch,
			Output: out,
		})
	} else {
		ctx.Build(pctx, android.BuildParams{
			Rule:   android.Cat,
			Output: out,
			Inputs: srcFiles,
		})
	}
	c.installSource = out
}

func (c *cilCompatMapExt) AndroidMk() android.AndroidMkData {
	ret := android.AndroidMkData{
		OutputFile: android.OptionalPathForPath(c.installSource),
		Class:      "ETC",
	}
	ret.Extra = append(ret.Extra, func(w io.Writer, outputFile android.Path) {
		fmt.Fprintln(w, "LOCAL_MODULE_PATH := $(TARGET_OUT_PRODUCT)/etc/selinux/mapping")
		if c.properties.Stem != nil {
			fmt.Fprintln(w, "LOCAL_MODULE_STEM := " + *c.properties.Stem)
		}
	})
	return ret
}
