#!/usr/bin/env python3

# Copyright 2021 The Android Open Source Project
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
import sys

# Parses a line in property_contexts and return a (prop, ctx) tuple.
# Raises an error for any malformed entries.
def parse_line(line):
    # A line should look like:
    #
    # {prop_name} u:object_r:{context_name}:s0
    tokens = line.split()
    if len(tokens) <= 1:
        raise ValueError('malformed entry "' + line + '" in property_contexts')

    context_tokens = tokens[1].split(':')
    if len(context_tokens) != 4:
        raise ValueError('malformed entry "' + line + '" in property_contexts')

    return tokens[0], context_tokens[2]

def parse_args():
    parser = argparse.ArgumentParser(
        description="Finds any violations in property_contexts, with given allowed prefixes. "
        "If any violations are found, return a nonzero (failure) exit code.")
    parser.add_argument("--property-contexts", help="Path to property_contexts file.")
    parser.add_argument("--allowed-property-prefix", action="extend", nargs="*",
        help="Allowed property prefixes. If empty, any properties are allowed.")
    parser.add_argument("--allowed-context-prefix", action="extend", nargs="*",
        help="Allowed context prefixes. If empty, any contexts are allowed.")

    return parser.parse_args()

args = parse_args()

violations = []

with open(args.property_contexts, 'r') as f:
    lines = f.read().split('\n')

for line in lines:
    tokens = line.strip()
    # if this line empty or a comment, skip
    if tokens == '' or tokens[0] == '#':
        continue

    prop, context = parse_line(line)

    violated = False

    if args.allowed_property_prefix and not prop.startswith(tuple(args.allowed_property_prefix)):
        violated = True

    if args.allowed_context_prefix and not context.startswith(tuple(args.allowed_context_prefix)):
        violated = True

    if violated:
        violations.append(line)

if len(violations) > 0:
    print('******************************')
    print('%d violations found:' % len(violations))
    print('\n'.join(violations))
    print('******************************')
    sys.exit(1)
