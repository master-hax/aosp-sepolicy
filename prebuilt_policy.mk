# Copyright (C) 2020 The Android Open Source Project
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

# prebuilt_policy.mk generates policy files from prebuilts of BOARD_SEPOLICY_VERS.
# The policy files will only be used to compile vendor and odm policies.
#
# Specifically, the following prebuilts are used...
# - system/sepolicy/prebuilts/api/{BOARD_SEPOLICY_VERS}
# - BOARD_PLAT_VENDOR_POLICY (copy of system/sepolicy/vendor)
# - BOARD_REQD_MASK_POLICY (copy of system/sepolicy/reqd_mask)
# - BOARD_SYSTEM_EXT_PUBLIC_PREBUILT_DIRS
# - BOARD_SYSTEM_EXT_PRIVATE_PREBUILT_DIRS
# - BOARD_PRODUCT_PUBLIC_PREBUILT_DIRS
# - BOARD_PRODUCT_PRIVATE_PREBUILT_DIRS
#
# ... to generate following policy files.
#
# - reqd policy mask
# - plat, system_ext, product public policy
# - plat, system_ext, product policy
# - plat, system_ext, product versioned policy
#
# These generated policy files will be used only when building vendor policies.
# They are not installed to system, system_ext, or product partition.
ver := $(BOARD_SEPOLICY_VERS)
prebuilt_dir := $(LOCAL_PATH)/prebuilts/api/$(ver)
plat_public_policy_$(ver) := $(prebuilt_dir)/public
plat_private_policy_$(ver) := $(prebuilt_dir)/private
system_ext_public_policy_$(ver) := $(BOARD_SYSTEM_EXT_PUBLIC_PREBUILT_DIRS)
system_ext_private_policy_$(ver) := $(BOARD_SYSTEM_EXT_PRIVATE_PREBUILT_DIRS)
product_public_policy_$(ver) := $(BOARD_PRODUCT_PUBLIC_PREBUILT_DIRS)
product_private_policy_$(ver) := $(BOARD_PRODUCT_PRIVATE_PREBUILT_DIRS)

##################################
# reqd_policy_mask_$(ver).cil
#
policy_files := $(call build_policy, $(sepolicy_build_files), $(BOARD_REQD_MASK_POLICY))
reqd_policy_mask_$(ver).conf := $(sepolicy_intermediates)/reqd_policy_mask_$(ver).conf
$(reqd_policy_mask_$(ver).conf): PRIVATE_MLS_SENS := $(MLS_SENS)
$(reqd_policy_mask_$(ver).conf): PRIVATE_MLS_CATS := $(MLS_CATS)
$(reqd_policy_mask_$(ver).conf): PRIVATE_TARGET_BUILD_VARIANT := $(TARGET_BUILD_VARIANT)
$(reqd_policy_mask_$(ver).conf): PRIVATE_TGT_ARCH := $(my_target_arch)
$(reqd_policy_mask_$(ver).conf): PRIVATE_TGT_WITH_ASAN := $(with_asan)
$(reqd_policy_mask_$(ver).conf): PRIVATE_TGT_WITH_NATIVE_COVERAGE := $(with_native_coverage)
$(reqd_policy_mask_$(ver).conf): PRIVATE_ADDITIONAL_M4DEFS := $(LOCAL_ADDITIONAL_M4DEFS)
$(reqd_policy_mask_$(ver).conf): PRIVATE_SEPOLICY_SPLIT := $(PRODUCT_SEPOLICY_SPLIT)
$(reqd_policy_mask_$(ver).conf): PRIVATE_COMPATIBLE_PROPERTY := $(PRODUCT_COMPATIBLE_PROPERTY)
$(reqd_policy_mask_$(ver).conf): PRIVATE_TREBLE_SYSPROP_NEVERALLOW := $(treble_sysprop_neverallow)
$(reqd_policy_mask_$(ver).conf): PRIVATE_ENFORCE_SYSPROP_OWNER := $(enforce_sysprop_owner)
$(reqd_policy_mask_$(ver).conf): PRIVATE_POLICY_FILES := $(policy_files)
$(reqd_policy_mask_$(ver).conf): $(policy_files) $(M4)
	$(transform-policy-to-conf)
# b/37755687
CHECKPOLICY_ASAN_OPTIONS := ASAN_OPTIONS=detect_leaks=0

reqd_policy_mask_$(ver).cil := $(sepolicy_intermediates)/reqd_policy_mask_$(ver).cil
$(reqd_policy_mask_$(ver).cil): $(reqd_policy_mask_$(ver).conf) $(HOST_OUT_EXECUTABLES)/checkpolicy
	@mkdir -p $(dir $@)
	$(hide) $(CHECKPOLICY_ASAN_OPTIONS) $(HOST_OUT_EXECUTABLES)/checkpolicy -C -M -c \
		$(POLICYVERS) -o $@ $<

reqd_policy_mask_$(ver).conf :=

reqd_policy_$(ver) := $(BOARD_REQD_MASK_POLICY)

##################################
# plat_pub_policy_$(ver).cil: exported plat policies
#
policy_files := $(call build_policy, $(sepolicy_build_files), \
  $(plat_public_policy_$(ver)) $(reqd_policy_$(ver)))
plat_pub_policy_$(ver).conf := $(sepolicy_intermediates)/plat_pub_policy_$(ver).conf
$(plat_pub_policy_$(ver).conf): PRIVATE_MLS_SENS := $(MLS_SENS)
$(plat_pub_policy_$(ver).conf): PRIVATE_MLS_CATS := $(MLS_CATS)
$(plat_pub_policy_$(ver).conf): PRIVATE_TARGET_BUILD_VARIANT := $(TARGET_BUILD_VARIANT)
$(plat_pub_policy_$(ver).conf): PRIVATE_TGT_ARCH := $(my_target_arch)
$(plat_pub_policy_$(ver).conf): PRIVATE_TGT_WITH_ASAN := $(with_asan)
$(plat_pub_policy_$(ver).conf): PRIVATE_TGT_WITH_NATIVE_COVERAGE := $(with_native_coverage)
$(plat_pub_policy_$(ver).conf): PRIVATE_ADDITIONAL_M4DEFS := $(LOCAL_ADDITIONAL_M4DEFS)
$(plat_pub_policy_$(ver).conf): PRIVATE_SEPOLICY_SPLIT := $(PRODUCT_SEPOLICY_SPLIT)
$(plat_pub_policy_$(ver).conf): PRIVATE_COMPATIBLE_PROPERTY := $(PRODUCT_COMPATIBLE_PROPERTY)
$(plat_pub_policy_$(ver).conf): PRIVATE_TREBLE_SYSPROP_NEVERALLOW := $(treble_sysprop_neverallow)
$(plat_pub_policy_$(ver).conf): PRIVATE_ENFORCE_SYSPROP_OWNER := $(enforce_sysprop_owner)
$(plat_pub_policy_$(ver).conf): PRIVATE_POLICY_FILES := $(policy_files)
$(plat_pub_policy_$(ver).conf): $(policy_files) $(M4)
	$(transform-policy-to-conf)

plat_pub_policy_$(ver).cil := $(sepolicy_intermediates)/plat_pub_policy_$(ver).cil
$(plat_pub_policy_$(ver).cil): PRIVATE_POL_CONF := $(plat_pub_policy_$(ver).conf)
$(plat_pub_policy_$(ver).cil): PRIVATE_REQD_MASK := $(reqd_policy_mask_$(ver).cil)
$(plat_pub_policy_$(ver).cil): $(HOST_OUT_EXECUTABLES)/checkpolicy \
$(HOST_OUT_EXECUTABLES)/build_sepolicy $(plat_pub_policy_$(ver).conf) $(reqd_policy_mask_$(ver).cil)
	@mkdir -p $(dir $@)
	$(hide) $(CHECKPOLICY_ASAN_OPTIONS) $< -C -M -c $(POLICYVERS) -o $@ $(PRIVATE_POL_CONF)
	$(hide) $(HOST_OUT_EXECUTABLES)/build_sepolicy -a $(HOST_OUT_EXECUTABLES) filter_out \
		-f $(PRIVATE_REQD_MASK) -t $@

plat_pub_policy_$(ver).conf :=

##################################
# plat_mapping_cil_$(ver).cil: versioned exported system policy
#
plat_mapping_cil_$(ver) := $(sepolicy_intermediates)/plat_mapping_$(ver).cil
$(plat_mapping_cil_$(ver)) : PRIVATE_VERS := $(ver)
$(plat_mapping_cil_$(ver)) : $(plat_pub_policy_$(ver).cil) $(HOST_OUT_EXECUTABLES)/version_policy
	@mkdir -p $(dir $@)
	$(hide) $(HOST_OUT_EXECUTABLES)/version_policy -b $< -m -n $(PRIVATE_VERS) -o $@
built_plat_mapping_cil_$(ver) := $(plat_mapping_cil_$(ver))

##################################
# plat_policy_$(ver).cil: system policy
#
policy_files := $(call build_policy, $(sepolicy_build_files), \
  $(plat_public_policy_$(ver)) $(plat_private_policy_$(ver)) )
plat_policy_$(ver).conf := $(sepolicy_intermediates)/plat_policy_$(ver).conf
$(plat_policy_$(ver).conf): PRIVATE_MLS_SENS := $(MLS_SENS)
$(plat_policy_$(ver).conf): PRIVATE_MLS_CATS := $(MLS_CATS)
$(plat_policy_$(ver).conf): PRIVATE_TARGET_BUILD_VARIANT := $(TARGET_BUILD_VARIANT)
$(plat_policy_$(ver).conf): PRIVATE_TGT_ARCH := $(my_target_arch)
$(plat_policy_$(ver).conf): PRIVATE_TGT_WITH_ASAN := $(with_asan)
$(plat_policy_$(ver).conf): PRIVATE_TGT_WITH_NATIVE_COVERAGE := $(with_native_coverage)
$(plat_policy_$(ver).conf): PRIVATE_ADDITIONAL_M4DEFS := $(LOCAL_ADDITIONAL_M4DEFS)
$(plat_policy_$(ver).conf): PRIVATE_SEPOLICY_SPLIT := $(PRODUCT_SEPOLICY_SPLIT)
$(plat_policy_$(ver).conf): PRIVATE_COMPATIBLE_PROPERTY := $(PRODUCT_COMPATIBLE_PROPERTY)
$(plat_policy_$(ver).conf): PRIVATE_TREBLE_SYSPROP_NEVERALLOW := $(treble_sysprop_neverallow)
$(plat_policy_$(ver).conf): PRIVATE_ENFORCE_SYSPROP_OWNER := $(enforce_sysprop_owner)
$(plat_policy_$(ver).conf): PRIVATE_POLICY_FILES := $(policy_files)
$(plat_policy_$(ver).conf): $(policy_files) $(M4)
	$(transform-policy-to-conf)
	$(hide) sed '/^\s*dontaudit.*;/d' $@ | sed '/^\s*dontaudit/,/;/d' > $@.dontaudit

plat_policy_$(ver).cil := $(sepolicy_intermediates)/plat_policy_$(ver).cil
$(plat_policy_$(ver).cil): PRIVATE_ADDITIONAL_CIL_FILES := \
  $(call build_policy, $(sepolicy_build_cil_workaround_files), $(plat_private_policy_$(ver)))
$(plat_policy_$(ver).cil): PRIVATE_NEVERALLOW_ARG := $(NEVERALLOW_ARG)
$(plat_policy_$(ver).cil): $(plat_policy_$(ver).conf) $(HOST_OUT_EXECUTABLES)/checkpolicy \
  $(HOST_OUT_EXECUTABLES)/secilc \
  $(call build_policy, $(sepolicy_build_cil_workaround_files), $(plat_private_policy_$(ver)))
	@mkdir -p $(dir $@)
	$(hide) $(CHECKPOLICY_ASAN_OPTIONS) $(HOST_OUT_EXECUTABLES)/checkpolicy -M -C -c \
		$(POLICYVERS) -o $@.tmp $<
	$(hide) cat $(PRIVATE_ADDITIONAL_CIL_FILES) >> $@.tmp
	$(hide) $(HOST_OUT_EXECUTABLES)/secilc -m -M true -G -c $(POLICYVERS) $(PRIVATE_NEVERALLOW_ARG) $@.tmp -o /dev/null -f /dev/null
	$(hide) mv $@.tmp $@

plat_policy_$(ver).conf :=

built_plat_cil_$(ver) := $(plat_policy_$(ver).cil)

ifdef HAS_SYSTEM_EXT_SEPOLICY_DIR

##################################
# system_ext_pub_policy_$(ver).cil: exported system and system_ext policy
#
policy_files := $(call build_policy, $(sepolicy_build_files), \
  $(plat_public_policy_$(ver)) $(system_ext_public_policy_$(ver)) $(reqd_policy_$(ver)))
system_ext_pub_policy_$(ver).conf := $(sepolicy_intermediates)/system_ext_pub_policy_$(ver).conf
$(system_ext_pub_policy_$(ver).conf): PRIVATE_MLS_SENS := $(MLS_SENS)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_MLS_CATS := $(MLS_CATS)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_TARGET_BUILD_VARIANT := $(TARGET_BUILD_VARIANT)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_TGT_ARCH := $(my_target_arch)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_TGT_WITH_ASAN := $(with_asan)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_TGT_WITH_NATIVE_COVERAGE := $(with_native_coverage)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_ADDITIONAL_M4DEFS := $(LOCAL_ADDITIONAL_M4DEFS)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_SEPOLICY_SPLIT := $(PRODUCT_SEPOLICY_SPLIT)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_COMPATIBLE_PROPERTY := $(PRODUCT_COMPATIBLE_PROPERTY)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_TREBLE_SYSPROP_NEVERALLOW := $(treble_sysprop_neverallow)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_ENFORCE_SYSPROP_OWNER := $(enforce_sysprop_owner)
$(system_ext_pub_policy_$(ver).conf): PRIVATE_POLICY_FILES := $(policy_files)
$(system_ext_pub_policy_$(ver).conf): $(policy_files) $(M4)
	$(transform-policy-to-conf)

system_ext_pub_policy_$(ver).cil := $(sepolicy_intermediates)/system_ext_pub_policy_$(ver).cil
$(system_ext_pub_policy_$(ver).cil): PRIVATE_POL_CONF := $(system_ext_pub_policy_$(ver).conf)
$(system_ext_pub_policy_$(ver).cil): PRIVATE_REQD_MASK := $(reqd_policy_mask_$(ver).cil)
$(system_ext_pub_policy_$(ver).cil): $(HOST_OUT_EXECUTABLES)/checkpolicy \
$(HOST_OUT_EXECUTABLES)/build_sepolicy $(system_ext_pub_policy_$(ver).conf) $(reqd_policy_mask_$(ver).cil)
	@mkdir -p $(dir $@)
	$(hide) $(CHECKPOLICY_ASAN_OPTIONS) $< -C -M -c $(POLICYVERS) -o $@ $(PRIVATE_POL_CONF)
	$(hide) $(HOST_OUT_EXECUTABLES)/build_sepolicy -a $(HOST_OUT_EXECUTABLES) filter_out \
		-f $(PRIVATE_REQD_MASK) -t $@

system_ext_pub_policy_$(ver).conf :=

##################################
# system_ext_policy_$(ver).cil: system_ext policy
#
policy_files := $(call build_policy, $(sepolicy_build_files), \
  $(plat_public_policy_$(ver)) $(plat_private_policy_$(ver)) \
  $(system_ext_public_policy_$(ver)) $(system_ext_private_policy_$(ver)) )
system_ext_policy_$(ver).conf := $(sepolicy_intermediates)/system_ext_policy_$(ver).conf
$(system_ext_policy_$(ver).conf): PRIVATE_MLS_SENS := $(MLS_SENS)
$(system_ext_policy_$(ver).conf): PRIVATE_MLS_CATS := $(MLS_CATS)
$(system_ext_policy_$(ver).conf): PRIVATE_TARGET_BUILD_VARIANT := $(TARGET_BUILD_VARIANT)
$(system_ext_policy_$(ver).conf): PRIVATE_TGT_ARCH := $(my_target_arch)
$(system_ext_policy_$(ver).conf): PRIVATE_TGT_WITH_ASAN := $(with_asan)
$(system_ext_policy_$(ver).conf): PRIVATE_TGT_WITH_NATIVE_COVERAGE := $(with_native_coverage)
$(system_ext_policy_$(ver).conf): PRIVATE_ADDITIONAL_M4DEFS := $(LOCAL_ADDITIONAL_M4DEFS)
$(system_ext_policy_$(ver).conf): PRIVATE_SEPOLICY_SPLIT := $(PRODUCT_SEPOLICY_SPLIT)
$(system_ext_policy_$(ver).conf): PRIVATE_COMPATIBLE_PROPERTY := $(PRODUCT_COMPATIBLE_PROPERTY)
$(system_ext_policy_$(ver).conf): PRIVATE_TREBLE_SYSPROP_NEVERALLOW := $(treble_sysprop_neverallow)
$(system_ext_policy_$(ver).conf): PRIVATE_ENFORCE_SYSPROP_OWNER := $(enforce_sysprop_owner)
$(system_ext_policy_$(ver).conf): PRIVATE_POLICY_FILES := $(policy_files)
$(system_ext_policy_$(ver).conf): $(policy_files) $(M4)
	$(transform-policy-to-conf)
	$(hide) sed '/dontaudit/d' $@ > $@.dontaudit

system_ext_policy_$(ver).cil := $(sepolicy_intermediates)/system_ext_policy_$(ver).cil
$(system_ext_policy_$(ver).cil): PRIVATE_NEVERALLOW_ARG := $(NEVERALLOW_ARG)
$(system_ext_policy_$(ver).cil): PRIVATE_PLAT_CIL := $(built_plat_cil_$(ver))
$(system_ext_policy_$(ver).cil): $(system_ext_policy_$(ver).conf) $(HOST_OUT_EXECUTABLES)/checkpolicy \
$(HOST_OUT_EXECUTABLES)/build_sepolicy $(HOST_OUT_EXECUTABLES)/secilc $(built_plat_cil_$(ver))
	@mkdir -p $(dir $@)
	$(hide) $(CHECKPOLICY_ASAN_OPTIONS) $(HOST_OUT_EXECUTABLES)/checkpolicy -M -C -c \
	$(POLICYVERS) -o $@ $<
	$(hide) $(HOST_OUT_EXECUTABLES)/build_sepolicy -a $(HOST_OUT_EXECUTABLES) filter_out \
		-f $(PRIVATE_PLAT_CIL) -t $@
	# Line markers (denoted by ;;) are malformed after above cmd. They are only
	# used for debugging, so we remove them.
	$(hide) grep -v ';;' $@ > $@.tmp
	$(hide) mv $@.tmp $@
	# Combine plat_sepolicy.cil and system_ext_sepolicy.cil to make sure that the
	# latter doesn't accidentally depend on vendor/odm policies.
	$(hide) $(HOST_OUT_EXECUTABLES)/secilc -m -M true -G -c $(POLICYVERS) \
		$(PRIVATE_NEVERALLOW_ARG) $(PRIVATE_PLAT_CIL) $@ -o /dev/null -f /dev/null

system_ext_policy_$(ver).conf :=

built_system_ext_cil_$(ver) := $(system_ext_policy_$(ver).cil)

##################################
# system_ext_mapping_cil_$(ver).cil: versioned exported system_ext policy
#
system_ext_mapping_cil_$(ver) := $(sepolicy_intermediates)/system_ext_mapping_$(ver).cil
$(system_ext_mapping_cil_$(ver)) : PRIVATE_VERS := $(ver)
$(system_ext_mapping_cil_$(ver)) : PRIVATE_PLAT_MAPPING_CIL := $(built_plat_mapping_cil_$(ver))
$(system_ext_mapping_cil_$(ver)) : $(system_ext_pub_policy_$(ver).cil) $(HOST_OUT_EXECUTABLES)/version_policy \
$(built_plat_mapping_cil_$(ver))
	@mkdir -p $(dir $@)
	# Generate system_ext mapping file as mapping file of 'system' (plat) and 'system_ext'
	# sepolicy minus plat_mapping_file.
	$(hide) $(HOST_OUT_EXECUTABLES)/version_policy -b $< -m -n $(PRIVATE_VERS) -o $@
	$(hide) $(HOST_OUT_EXECUTABLES)/build_sepolicy -a $(HOST_OUT_EXECUTABLES) filter_out \
		-f $(PRIVATE_PLAT_MAPPING_CIL) -t $@

built_system_ext_mapping_cil_$(ver) := $(system_ext_mapping_cil_$(ver))

endif # ifdef HAS_SYSTEM_EXT_SEPOLICY_DIR

ifdef HAS_PRODUCT_SEPOLICY_DIR

##################################
# product_policy_$(ver).cil: product policy
#
policy_files := $(call build_policy, $(sepolicy_build_files), \
  $(plat_public_policy_$(ver)) $(plat_private_policy_$(ver)) \
  $(system_ext_public_policy_$(ver)) $(system_ext_private_policy_$(ver)) \
  $(product_public_policy_$(ver)) $(product_private_policy_$(ver)) )
product_policy_$(ver).conf := $(sepolicy_intermediates)/product_policy_$(ver).conf
$(product_policy_$(ver).conf): PRIVATE_MLS_SENS := $(MLS_SENS)
$(product_policy_$(ver).conf): PRIVATE_MLS_CATS := $(MLS_CATS)
$(product_policy_$(ver).conf): PRIVATE_TARGET_BUILD_VARIANT := $(TARGET_BUILD_VARIANT)
$(product_policy_$(ver).conf): PRIVATE_TGT_ARCH := $(my_target_arch)
$(product_policy_$(ver).conf): PRIVATE_TGT_WITH_ASAN := $(with_asan)
$(product_policy_$(ver).conf): PRIVATE_TGT_WITH_NATIVE_COVERAGE := $(with_native_coverage)
$(product_policy_$(ver).conf): PRIVATE_ADDITIONAL_M4DEFS := $(LOCAL_ADDITIONAL_M4DEFS)
$(product_policy_$(ver).conf): PRIVATE_SEPOLICY_SPLIT := $(PRODUCT_SEPOLICY_SPLIT)
$(product_policy_$(ver).conf): PRIVATE_COMPATIBLE_PROPERTY := $(PRODUCT_COMPATIBLE_PROPERTY)
$(product_policy_$(ver).conf): PRIVATE_TREBLE_SYSPROP_NEVERALLOW := $(treble_sysprop_neverallow)
$(product_policy_$(ver).conf): PRIVATE_ENFORCE_SYSPROP_OWNER := $(enforce_sysprop_owner)
$(product_policy_$(ver).conf): PRIVATE_POLICY_FILES := $(policy_files)
$(product_policy_$(ver).conf): $(policy_files) $(M4)
	$(transform-policy-to-conf)
	$(hide) sed '/dontaudit/d' $@ > $@.dontaudit

product_policy_$(ver).cil := $(sepolicy_intermediates)/product_policy_$(ver).cil
$(product_policy_$(ver).cil): PRIVATE_NEVERALLOW_ARG := $(NEVERALLOW_ARG)
$(product_policy_$(ver).cil): PRIVATE_PLAT_CIL_FILES := $(built_plat_cil_$(ver)) $(built_system_ext_cil_$(ver))
$(product_policy_$(ver).cil): $(product_policy_$(ver).conf) $(HOST_OUT_EXECUTABLES)/checkpolicy \
$(HOST_OUT_EXECUTABLES)/build_sepolicy $(HOST_OUT_EXECUTABLES)/secilc \
$(built_plat_cil_$(ver)) $(built_system_ext_cil_$(ver))
	@mkdir -p $(dir $@)
	$(hide) $(CHECKPOLICY_ASAN_OPTIONS) $(HOST_OUT_EXECUTABLES)/checkpolicy -M -C -c \
	$(POLICYVERS) -o $@ $<
	$(hide) $(HOST_OUT_EXECUTABLES)/build_sepolicy -a $(HOST_OUT_EXECUTABLES) filter_out \
		-f $(PRIVATE_PLAT_CIL) -t $@
	# Line markers (denoted by ;;) are malformed after above cmd. They are only
	# used for debugging, so we remove them.
	$(hide) grep -v ';;' $@ > $@.tmp
	$(hide) mv $@.tmp $@
	# Combine plat_sepolicy.cil, system_ext_sepolicy.cil and product_sepolicy.cil to
	# make sure that the latter doesn't accidentally depend on vendor/odm policies.
	$(hide) $(HOST_OUT_EXECUTABLES)/secilc -m -M true -G -c $(POLICYVERS) \
		$(PRIVATE_NEVERALLOW_ARG) $(PRIVATE_PLAT_CIL_FILES) $@ -o /dev/null -f /dev/null

product_policy_$(ver).conf :=

built_product_cil_$(ver) := $(product_policy_$(ver).cil)

endif # ifdef HAS_PRODUCT_SEPOLICY_DIR

##################################
# pub_policy_$(ver).cil: exported plat, system_ext, and product policies
#
policy_files := $(call build_policy, $(sepolicy_build_files), \
  $(plat_public_policy_$(ver)) $(system_ext_public_policy_$(ver)) \
  $(product_public_policy_$(ver)) $(reqd_policy_$(ver)) )
pub_policy_$(ver).conf := $(sepolicy_intermediates)/pub_policy_$(ver).conf
$(pub_policy_$(ver).conf): PRIVATE_MLS_SENS := $(MLS_SENS)
$(pub_policy_$(ver).conf): PRIVATE_MLS_CATS := $(MLS_CATS)
$(pub_policy_$(ver).conf): PRIVATE_TARGET_BUILD_VARIANT := $(TARGET_BUILD_VARIANT)
$(pub_policy_$(ver).conf): PRIVATE_TGT_ARCH := $(my_target_arch)
$(pub_policy_$(ver).conf): PRIVATE_TGT_WITH_ASAN := $(with_asan)
$(pub_policy_$(ver).conf): PRIVATE_TGT_WITH_NATIVE_COVERAGE := $(with_native_coverage)
$(pub_policy_$(ver).conf): PRIVATE_ADDITIONAL_M4DEFS := $(LOCAL_ADDITIONAL_M4DEFS)
$(pub_policy_$(ver).conf): PRIVATE_SEPOLICY_SPLIT := $(PRODUCT_SEPOLICY_SPLIT)
$(pub_policy_$(ver).conf): PRIVATE_COMPATIBLE_PROPERTY := $(PRODUCT_COMPATIBLE_PROPERTY)
$(pub_policy_$(ver).conf): PRIVATE_TREBLE_SYSPROP_NEVERALLOW := $(treble_sysprop_neverallow)
$(pub_policy_$(ver).conf): PRIVATE_ENFORCE_SYSPROP_OWNER := $(enforce_sysprop_owner)
$(pub_policy_$(ver).conf): PRIVATE_POLICY_FILES := $(policy_files)
$(pub_policy_$(ver).conf): $(policy_files) $(M4)
	$(transform-policy-to-conf)

pub_policy_$(ver).cil := $(sepolicy_intermediates)/pub_policy_$(ver).cil
$(pub_policy_$(ver).cil): PRIVATE_POL_CONF := $(pub_policy_$(ver).conf)
$(pub_policy_$(ver).cil): PRIVATE_REQD_MASK := $(reqd_policy_mask_$(ver).cil)
$(pub_policy_$(ver).cil): $(HOST_OUT_EXECUTABLES)/checkpolicy \
$(HOST_OUT_EXECUTABLES)/build_sepolicy $(pub_policy_$(ver).conf) $(reqd_policy_mask_$(ver).cil)
	@mkdir -p $(dir $@)
	$(hide) $(CHECKPOLICY_ASAN_OPTIONS) $< -C -M -c $(POLICYVERS) -o $@ $(PRIVATE_POL_CONF)
	$(hide) $(HOST_OUT_EXECUTABLES)/build_sepolicy -a $(HOST_OUT_EXECUTABLES) filter_out \
		-f $(PRIVATE_REQD_MASK) -t $@

pub_policy_$(ver).conf :=

ifdef HAS_PRODUCT_SEPOLICY_DIR

##################################
# product_mapping_cil_$(ver).cil: versioned exported product policy
#
product_mapping_cil_$(ver) := $(sepolicy_intermediates)/product_mapping_cil_$(ver).cil
$(product_mapping_cil_$(ver)) : PRIVATE_VERS := $(ver)
$(product_mapping_cil_$(ver)) : PRIVATE_FILTER_CIL_FILES := $(built_plat_mapping_cil_$(ver)) $(built_system_ext_mapping_cil_$(ver))
$(product_mapping_cil_$(ver)) : $(pub_policy_$(ver).cil) $(HOST_OUT_EXECUTABLES)/version_policy \
$(built_plat_mapping_cil_$(ver)) $(built_system_ext_mapping_cil_$(ver))
	@mkdir -p $(dir $@)
	# Generate product mapping file as mapping file of all public sepolicy minus
	# plat_mapping_file and system_ext_mapping_file.
	$(hide) $(HOST_OUT_EXECUTABLES)/version_policy -b $< -m -n $(PRIVATE_VERS) -o $@
	$(hide) $(HOST_OUT_EXECUTABLES)/build_sepolicy -a $(HOST_OUT_EXECUTABLES) filter_out \
		-f $(PRIVATE_FILTER_CIL_FILES) -t $@

built_product_mapping_cil_$(ver) := $(product_mapping_cil_$(ver))

endif # ifdef HAS_PRODUCT_SEPOLICY_DIR

##################################
# plat_pub_versioned_$(ver).cil - the exported platform policy
#
plat_pub_versioned_$(ver).cil := $(sepolicy_intermediates)/plat_pub_versioned_$(ver).cil
$(plat_pub_versioned_$(ver).cil) : PRIVATE_VERS := $(ver)
$(plat_pub_versioned_$(ver).cil) : PRIVATE_TGT_POL := $(pub_policy_$(ver).cil)
$(plat_pub_versioned_$(ver).cil) : PRIVATE_DEP_CIL_FILES := $(built_plat_cil_$(ver)) $(built_system_ext_cil_$(ver)) \
$(built_product_cil_$(ver)) $(built_plat_mapping_cil_$(ver)) $(built_system_ext_mapping_cil_$(ver)) \
$(built_product_mapping_cil_$(ver))
$(plat_pub_versioned_$(ver).cil) : $(pub_policy_$(ver).cil) $(HOST_OUT_EXECUTABLES)/version_policy \
  $(HOST_OUT_EXECUTABLES)/secilc $(built_plat_cil_$(ver)) $(built_system_ext_cil_$(ver)) $(built_product_cil_$(ver)) \
  $(built_plat_mapping_cil_$(ver)) $(built_system_ext_mapping_cil_$(ver)) $(built_product_mapping_cil_$(ver))
	@mkdir -p $(dir $@)
	$(HOST_OUT_EXECUTABLES)/version_policy -b $< -t $(PRIVATE_TGT_POL) -n $(PRIVATE_VERS) -o $@
	$(hide) $(HOST_OUT_EXECUTABLES)/secilc -m -M true -G -N -c $(POLICYVERS) \
		$(PRIVATE_DEP_CIL_FILES) $@ -o /dev/null -f /dev/null

built_pub_vers_cil_$(ver) := $(plat_pub_versioned_$(ver).cil)
