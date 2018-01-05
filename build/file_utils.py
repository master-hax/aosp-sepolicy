# Copyright 2018 - The Android Open Source Project
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

"""File-related utilities."""


import os
import re
import shutil
import tempfile


def make_parent_dirs(file_path):
    """Creates parent directories for the file_path."""
    if os.path.exists(file_path):
        return

    parent_dir = os.path.dirname(file_path)
    if parent_dir and not os.path.exists(parent_dir):
        os.makedirs(parent_dir)


def filter_out_lines(pattern_files, input_file, output_file=None):
    """Outputs input_file lines that do not match any line in pattern_files.

    Args:
        pattern_files: a list of files for filter patterns.
        input_file: a file used for filter input.
        output_file: a file to save the filtered results.  If None, input_file
            will be modified in-place.
    """
    # Prepares patterns.
    patterns = []
    for f in pattern_files:
        patterns.extend(open(f).readlines())

    # Copy lines that are not in the pattern.
    tmp_output = tempfile.NamedTemporaryFile()
    with open(input_file, 'r') as in_file:
        tmp_output.writelines(line for line in in_file.readlines()
                              if line not in patterns)
        tmp_output.flush()

    # Saves the result to the target file.
    copy_to = output_file if output_file else input_file
    shutil.copyfile(tmp_output.name, copy_to)


def filter_out_re(pattern, input_file, output_file=None):
    """Outputs input_file lines that do not match the RE pattern.

    Args:
        pattern: the regular expression pattern.
        input_file: a file used for filter input.
        output_file: a file to save the filtered results.  If None, input_file
            will be modified in-place.
    """
    # Copy lines that are not in the pattern.
    tmp_output = tempfile.NamedTemporaryFile()
    with open(input_file, 'r') as in_file:
        tmp_output.writelines(line for line in in_file.readlines()
                              if not re.search(pattern, line))
        tmp_output.flush()

    # Saves the result to the target file.
    copy_to = output_file if output_file else input_file
    shutil.copyfile(tmp_output.name, copy_to)
