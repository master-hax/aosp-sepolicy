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

LOCAL_MODULE := plat_property_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_pc := $(TARGET_OUT)/etc/selinux/plat_property_contexts

plat_property_contexts_test := $(intermediates)/plat_property_contexts_test
$(plat_property_contexts_test): PRIVATE_FC := $(local_pc)
$(plat_property_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(plat_property_contexts_test): $(HOST_OUT_EXECUTABLES)/property_info_checker $(local_pc) $(built_sepolicy)
	$(hide) $< $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_pc :=

##################################
include $(CLEAR_VARS)

LOCAL_MODULE := product_property_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_PRODUCT_MODULE := true
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_pc := $(TARGET_OUT_PRODUCT)/etc/selinux/product_property_contexts

product_property_contexts_test := $(intermediates)/product_property_contexts_test
$(product_property_contexts_test): PRIVATE_FC := $(local_pc)
$(product_property_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(product_property_contexts_test): $(HOST_OUT_EXECUTABLES)/property_info_checker $(local_pc) $(built_sepolicy)
	$(hide) $< $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_pc :=

##################################
include $(CLEAR_VARS)

LOCAL_MODULE := vendor_property_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_VENDOR_MODULE := true
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_pc := $(TARGET_OUT_VENDOR)/etc/selinux/vendor_property_contexts

vendor_property_contexts_test := $(intermediates)/vendor_property_contexts_test
$(vendor_property_contexts_test): PRIVATE_FC := $(local_pc)
$(vendor_property_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(vendor_property_contexts_test): $(HOST_OUT_EXECUTABLES)/property_info_checker $(local_pc) $(built_sepolicy)
	$(hide) $< $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_pc :=

##################################
include $(CLEAR_VARS)

LOCAL_MODULE := odm_property_contexts_test
LOCAL_MODULE_CLASS := ETC
LOCAL_ODM_MODULE := true
LOCAL_MODULE_TAGS := tests

include $(BUILD_SYSTEM)/base_rules.mk

local_pc := $(TARGET_OUT_ODM)/etc/selinux/odm_property_contexts

odm_property_contexts_test := $(intermediates)/odm_property_contexts_test
$(odm_property_contexts_test): PRIVATE_FC := $(local_pc)
$(odm_property_contexts_test): PRIVATE_SEPOLICY := $(built_sepolicy)
$(odm_property_contexts_test): $(HOST_OUT_EXECUTABLES)/property_info_checker $(local_pc) $(built_sepolicy)
	$(hide) $< $(PRIVATE_SEPOLICY) $(PRIVATE_FC)
	@mkdir -p $(dir $@)
	$(hide) touch $@

local_pc :=
