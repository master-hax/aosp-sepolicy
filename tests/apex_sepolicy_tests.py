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
import re
import shutil
import sys
import tempfile
from dataclasses import dataclass
from typing import List

import policy


SHARED_LIB_EXTENSION = '.dylib' if sys.platform == 'darwin' else '.so'
LIBSEPOLWRAP = "libsepolwrap" + SHARED_LIB_EXTENSION


@dataclass
class Is:
    """Exact matcher for a path."""
    path: str


@dataclass
class Glob:
    """Path matcher with pathlib.PurePath.match"""
    pattern: str


@dataclass
class Regex:
    """Path matcher with re.match"""
    pattern: str


Matcher = Is | Glob | Regex

@dataclass
class AllowRead:
    """Rule checking if scontext can read the entity"""
    tclass: str
    scontext: set[str]


Rule = AllowRead


def match_path(path: str, matcher: Matcher) -> bool:
    """True if path matches with the given matcher"""
    match matcher:
        case Is(target):
            return path == target
        case Glob(pattern):
            return pathlib.PurePath(path).match(pattern)
        case Regex(pattern):
            return re.match(pattern, path)


def check_rule(pol, path: str, tcontext: str, rule: Rule) -> List[str]:
    """Returns error message if scontext can't read the target"""
    match rule:
        case AllowRead(tclass, scontext):
            te_rules = list(pol.QueryTERule(scontext=scontext,
                                            tcontext={tcontext},
                                            tclass={tclass},
                                            perms={'read'}))
            if len(te_rules) > 0:
                return []  # no errors

            return [f"Error: {path}: {scontext} can't read. (tcontext={tcontext})"]


rules = [
    # permissions
    (Is('./etc/permissions/'), AllowRead('dir', {'system_server'})),
    (Glob('./etc/permissions/*.xml'), AllowRead('file', {'system_server'})),
    # init scripts
    (Regex('\./etc/.*\.\d*rc'), AllowRead('file', {'init'})),
    # vintf fragments
    (Is('./etc/vintf/'), AllowRead('dir', {'servicemanager', 'apexd'})),
    (Glob('./etc/vintf/*.xml'), AllowRead('file', {'servicemanager', 'apexd'})),
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

    errors = []
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
                if match_path(path, matcher):
                    errors.extend(check_rule(pol, path, tcontext, rule))
    if len(errors) > 0:
        sys.exit('\n'.join(errors))


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
