#!/usr/bin/env python3
#
# Copyright 2023 The Android Open Source Project
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
""" A tool to test APEX file_contexts

Usage:
    $ deapexer list --dir -Z foo.apex > /tmp/fc
    $ apex_sepolicy_tests -f /tmp/fc -p policy
"""


import argparse
import os
import pathlib
import pkgutil
import shutil
import sys
import tempfile

import policy


SHARED_LIB_EXTENSION = '.dylib' if sys.platform == 'darwin' else '.so'
LIBSEPOLWRAP = "libsepolwrap" + SHARED_LIB_EXTENSION


class Is: # pylint: disable=too-few-public-methods
    """
        Exact matcher for a path.
    """

    def __init__(self, path):
        self._path = path

    def match(self, path):
        """True if path is equal to the given"""
        return self._path == path


class Match: # pylint: disable=too-few-public-methods
    """
        Path matcher with pathlib.PurePath.match.
    """

    def __init__(self, pattern):
        self._pattern = pattern

    def match(self, path):
        """True if path matches with the given pattern"""
        return pathlib.PurePath(path).match(self._pattern)


class AllowRead: # pylint: disable=too-few-public-methods
    """
        Rule checking if scontext can read the entity
    """

    def __init__(self, tclass: str, scontext: set):
        self._tclass = tclass
        self._scontext = scontext

    def check(self, pol, path, tcontext):
        """Returns error message if scontext can't read the target"""
        te_rules = list(pol.QueryTERule(scontext=self._scontext,
                                        tcontext={tcontext},
                                        tclass={self._tclass},
                                        perms={'read'}))
        if len(te_rules) > 0:
            return ''

        return f"Error: {path}({tcontext},{self._tclass}): not readable by {self._scontext}"


rules = [
    # permissions
    (Is('./etc/permissions'), AllowRead('dir', {'system_server'})),
    (Match('./etc/permissions/*.xml'), AllowRead('file', {'system_server'})),
    # init scripts
    (Match('./etc/*rc'), AllowRead('file', {'init'})),
    # vintf fragments
    (Is('./etc/vintf'), AllowRead('dir', {'servicemanager', 'apexd'})),
    (Match('./etc/vintf/*.xml'), AllowRead('file', {'servicemanager', 'apexd'})),
    # ./ and apex_manifest.pb
    (Is('./apex_manifest.pb'), AllowRead('file', {'linkerconfig', 'apexd'})),
    (Is('./'), AllowRead('dir', {'linkerconfig', 'apexd'})),
]


def do_main():
    """Do testing"""
    parser = argparse.ArgumentParser()
    parser.add_argument('-f', '--file_contexts', help='output of "deapexer list -Z --dir"')
    parser.add_argument('-p', '--policy', help='compiled policy')

    args = parser.parse_args()

    pol = policy.Policy(args.policy, None, lib_path)

    ret = ''
    with open(args.file_contexts, 'rt', encoding='utf-8') as file_contexts:
        for line in file_contexts.readlines():
            # skip empty/comment line
            line = line.strip()
            if line == '' or line[0] == '#':
                continue

            # parse
            split = line.split()
            if len(split) != 2:
                sys.exit(f"Error: invalid file_contexts: {line}\n")
            path, context = split[0], split[1]
            tcontext = context.split(':')[2]

            # check rules
            for matcher, rule in rules:
                if matcher.match(path):
                    ret += rule.check(pol, path, tcontext)
    if len(ret) > 0:
        sys.exit(ret)


if __name__ == '__main__':
    temp_dir = tempfile.mkdtemp()
    try:
        # Extract libsepolwrap from the package
        lib_path = os.path.join(temp_dir, LIBSEPOLWRAP)
        with open(lib_path, 'wb') as f:
            blob = pkgutil.get_data('apex_sepolicy_tests', LIBSEPOLWRAP)
            if not blob:
                sys.exit(f"Error: {LIBSEPOLWRAP} does not exist. Is this binary corrupted?\n")
            f.write(blob)

        do_main()
    finally:
        shutil.rmtree(temp_dir)
