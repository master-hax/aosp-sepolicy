#!/bin/bash

# Copyright (C) 2023 The Android Open Source Project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

if [ $# -ne 2 ]; then
    echo "Usage: $0 <top> <ver>"
    exit 1
fi

top=$1
ver=$2

mkdir -p "$top/system/sepolicy/prebuilts/api/${ver}/"
cp -r "$top/system/sepolicy/public/" "$top/system/sepolicy/prebuilts/api/${ver}/"
cp -r "$top/system/sepolicy/private/" "$top/system/sepolicy/prebuilts/api/${ver}/"

cat > "$top/system/sepolicy/prebuilts/api/${ver}/Android.bp" <<EOF
// Automatically generated file, do not edit!
se_policy_conf {
    name: "{ver}_reqd_policy_mask.conf",
    defaults: ["se_policy_conf_flags_defaults"],
    srcs: reqd_mask_policy,
    installable: false,
    build_variant: "user",
    board_api_level: "{ver}",
}

se_policy_cil {
    name: "{ver}_reqd_policy_mask.cil",
    src: ":{ver}_reqd_policy_mask.conf",
    secilc_check: false,
    installable: false,
}

se_policy_conf {
    name: "{ver}_plat_pub_policy.conf",
    defaults: ["se_policy_conf_flags_defaults"],
    srcs: [
        ":se_build_files{.plat_public_{ver}}",
        ":se_build_files{.reqd_mask}",
    ],
    installable: false,
    build_variant: "user",
    board_api_level: "{ver}",
}

se_policy_cil {
    name: "{ver}_plat_pub_policy.cil",
    src: ":{ver}_plat_pub_policy.conf",
    filter_out: [":{ver}_reqd_policy_mask.cil"],
    secilc_check: false,
    installable: false,
}

se_policy_conf {
    name: "{ver}_product_pub_policy.conf",
    defaults: ["se_policy_conf_flags_defaults"],
    srcs: [
        ":se_build_files{.plat_public_{ver}}",
        ":se_build_files{.system_ext_public_{ver}}",
        ":se_build_files{.product_public_{ver}}",
        ":se_build_files{.reqd_mask}",
    ],
    installable: false,
    build_variant: "user",
    board_api_level: "{ver}",
}

se_policy_cil {
    name: "{ver}_product_pub_policy.cil",
    src: ":{ver}_product_pub_policy.conf",
    filter_out: [":{ver}_reqd_policy_mask.cil"],
    secilc_check: false,
    installable: false,
}

se_versioned_policy {
    name: "{ver}_plat_pub_versioned.cil",
    base: ":{ver}_product_pub_policy.cil",
    target_policy: ":{ver}_product_pub_policy.cil",
    version: "{ver}",
    installable: false,
}

se_policy_conf {
    name: "{ver}_plat_policy.conf",
    defaults: ["se_policy_conf_flags_defaults"],
    srcs: [
        ":se_build_files{.plat_public_{ver}}",
        ":se_build_files{.plat_private_{ver}}",
        ":se_build_files{.system_ext_public_{ver}}",
        ":se_build_files{.system_ext_private_{ver}}",
        ":se_build_files{.product_public_{ver}}",
        ":se_build_files{.product_private_{ver}}",
    ],
    installable: false,
    build_variant: "user",
    board_api_level: "{ver}",
}

se_policy_cil {
    name: "{ver}_plat_policy.cil",
    src: ":{ver}_plat_policy.conf",
    additional_cil_files: [":sepolicy_technical_debt{.plat_private_{ver}}"],
    installable: false,
}

se_policy_binary {
    name: "{ver}_plat_policy",
    srcs: [":{ver}_plat_policy.cil"],
    installable: false,
    dist: {
        targets: ["base-sepolicy-files-for-mapping"],
    },
}

filegroup {
    name: "{ver}_sepolicy_cts_data",
    srcs: [
        "{ver}_general_sepolicy.conf",
        "{ver}_plat_sepolicy.cil",
        "{ver}_mapping.cil",
    ],
}
EOF
