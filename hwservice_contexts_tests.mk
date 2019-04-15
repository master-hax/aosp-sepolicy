# Copyright (C) 2019 The Android Open Source Project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

include $(CLEAR_VARS)

LOCAL_MODULE := plat_hwservice_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_hsc := $(TARGET_OUT)/etc/selinux/plat_hwservice_contexts

plat_hwservice_contexts_test := $(intermediates)/plat_hwservice_contexts_test
$(plat_hwservice_contexts_test): PRIVATE_FC := $(local_hsc)
$(plat_hwservice_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(plat_hwservice_contexts_test): $(HOST_OUT_EXECUTABLES)/checkfc $(local_hsc) $(built_sepolicy)
	$(hide) $< -e -l $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_hsc :=

##################################
include $(CLEAR_VARS)

LOCAL_MODULE := product_hwservice_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_PRODUCT_MODULE := true
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_hsc := $(TARGET_OUT_PRODUCT)/etc/selinux/product_hwservice_contexts

product_hwservice_contexts_test := $(intermediates)/product_hwservice_contexts_test
$(product_hwservice_contexts_test): PRIVATE_FC := $(local_hsc)
$(product_hwservice_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(product_hwservice_contexts_test): $(HOST_OUT_EXECUTABLES)/checkfc $(local_hsc) $(built_sepolicy)
	$(hide) $< -e -l $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_hsc :=

##################################
include $(CLEAR_VARS)

LOCAL_MODULE := vendor_hwservice_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_VENDOR_MODULE := true
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_hsc := $(TARGET_OUT_VENDOR)/etc/selinux/vendor_hwservice_contexts

vendor_hwservice_contexts_test := $(intermediates)/vendor_hwservice_contexts_test
$(vendor_hwservice_contexts_test): PRIVATE_FC := $(local_hsc)
$(vendor_hwservice_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(vendor_hwservice_contexts_test): $(HOST_OUT_EXECUTABLES)/checkfc $(local_hsc) $(built_sepolicy)
	$(hide) $< -e -l $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_hsc :=

##################################
include $(CLEAR_VARS)

LOCAL_MODULE := odm_hwservice_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_ODM_MODULE := true
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_hsc := $(TARGET_OUT_ODM)/etc/selinux/odm_hwservice_contexts

odm_hwservice_contexts_test := $(intermediates)/odm_hwservice_contexts_test
$(odm_hwservice_contexts_test): PRIVATE_FC := $(local_hsc)
$(odm_hwservice_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(odm_hwservice_contexts_test): $(HOST_OUT_EXECUTABLES)/checkfc $(local_hsc) $(built_sepolicy)
	$(hide) $< -e -l $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_hsc :=
