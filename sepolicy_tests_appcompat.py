#!/usr/bin/python3

#
# Copyright (C) 2024 The Android Open Source Project
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

import os
import subprocess
import sys
import pkgutil
import tempfile
import stat

def get_resource_path(resource_name, mode=0o644):
    """Retrieves a resource from the package and writes it to a temporary file.

    Args:
        resource_name: The name of the resource.
        mode: File permissions mode (default is 0o644).
    """
    data = pkgutil.get_data(__name__, resource_name)
    if not data:
        raise FileNotFoundError(f"Resource not found: {resource_name}")
    with tempfile.NamedTemporaryFile(delete=False) as temp_file:
        temp_file.write(data)
        resource_path = temp_file.name
    os.chmod(resource_path, mode)  # Set permissions
    return resource_path

def main():
    print("NOTE: appcompat is still under development. It can report")
    print("API uses that do not execute at runtime, and reflection uses")
    print("that do not exist. It can also miss on reflection uses.")

    sepolicy_tests_path = get_resource_path("sepolicy_tests", 0o755)
    plat_file_contexts_path = get_resource_path("plat_file_contexts")
    vendor_file_contexts_path = get_resource_path("vendor_file_contexts")
    system_ext_file_contexts_path = get_resource_path("system_ext_file_contexts")
    product_file_contexts_path = get_resource_path("product_file_contexts")
    odm_file_contexts_path = get_resource_path("odm_file_contexts")
    precompiled_sepolicy_path = get_resource_path("precompiled_sepolicy")

    args = [
        sepolicy_tests_path,
        f"-f {plat_file_contexts_path} ",
        f"-f {vendor_file_contexts_path} ",
        f"-f {system_ext_file_contexts_path} ",
        f"-f {product_file_contexts_path} ",
        f"-f {odm_file_contexts_path} ",
        f"-p precompiled_sepolicy_path} ",
        f" && touch sepolicy_test",
    ] + sys.argv[1:]

    subprocess.run(args)

if __name__ == "__main__":
    main()
