// Copyright 2024 The Android Open Source Project
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
	"os"
	"reflect"
	"testing"

	"android/soong/android"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

var prepareForTest = android.GroupFixturePreparers(
	android.FixtureModifyProductVariables(func(variables android.FixtureProductVariables) {
		buildFlags := make(map[string]string)
		buildFlags["RELEASE_FLAGS_BAR"] = "true"
		buildFlags["RELEASE_FLAGS_FOO1"] = "false"
		// "RELEASE_FLAGS_FOO2" is missing
		buildFlags["RELEASE_AVF_ENABLE_DEVICE_ASSIGNMENT"] = "true"
		variables.BuildFlags = buildFlags
	}),
	android.FixtureRegisterWithContext(func(ctx android.RegistrationContext) {
		ctx.RegisterModuleType("se_flags", flagsFactory)
		ctx.RegisterParallelSingletonModuleType("se_flags_collector", flagsCollectorFactory)
	}),
)

func TestFlagCollector(t *testing.T) {
	t.Parallel()

	ctx := android.GroupFixturePreparers(
		prepareForTest,
		android.FixtureAddTextFile("package_bar/Android.bp", `
			se_flags {
				name: "se_flags_bar",
				flags: ["RELEASE_FLAGS_BAR"],
			}
			`),
		android.FixtureAddTextFile("package_foo/Android.bp", `
			se_flags {
				name: "se_flags_foo",
				flags: ["RELEASE_FLAGS_FOO1", "RELEASE_FLAGS_FOO2"],
			}
			`),
		android.FixtureAddTextFile("system/sepolicy/Android.bp", `
			se_flags {
				name: "se_flags",
				flags: ["RELEASE_AVF_ENABLE_DEVICE_ASSIGNMENT"],
			}
			se_flags_collector {
				name: "se_flags_collector",
			}
			`),
	).RunTest(t).TestContext

	collector := ctx.ModuleForTests("se_flags_collector", "").Module().(buildFlagsModule)
	actual := flagsToM4Macros(collector.getBuildFlags())
	expected := []string{
		"-D target_flag_RELEASE_AVF_ENABLE_DEVICE_ASSIGNMENT=true",
		"-D target_flag_RELEASE_FLAGS_BAR=true",
		"-D target_flag_RELEASE_FLAGS_FOO1=false",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("M4 macros were not exported correctly"+
			"\nactual:   %v"+
			"\nexpected: %v",
			actual,
			expected,
		)
	}
}
