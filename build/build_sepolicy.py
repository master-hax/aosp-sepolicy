#!/usr/bin/env python
#
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

"""Command-line tool to build SEPolicy files."""

import argparse
import os
import subprocess
import sys
import tempfile

import file_utils

# All supported commands in this module.
# For each command, need to add two functions. Take 'build_cil' for example:
#   - setup_build_cil()
#     - Sets up command parsers and sets default function to do_build_cil().
#   - do_build_cil()
_SUPPORTED_COMMANDS = ('build_cil',)


def run_host_command(args, **kwargs):
    """Runs a host command and prints output."""
    if kwargs.get('shell'):
        command_log = args
    else:
        command_log = ' '.join(args)
    sys.stdout.write('sepolicy - running command: {}\n'.format(command_log))
    subprocess.check_call(args, **kwargs)


def do_build_cil(args):
    """Builds a sepolicy CIL (Common Intermediate Language) file.

    This functions invokes some host utils (e.g., secilc, checkpolicy,
    version_sepolicy) to generate a .cil file.

    Args:
        args: the parsed command arguments.
    """
    # Builds the policy.conf via 'm4'.
    output_policy_conf = args.output_policy_conf
    file_utils.make_parent_dirs(output_policy_conf)
    m4_cmd = ['m4']
    # Prepares m4 definitions, note that the order matters.
    # https://www.gnu.org/software/m4/manual/m4.html#Preprocessor-features
    if args.m4_additional_defs:
        m4_cmd += args.m4_additional_defs
    m4_cmd += ['-D', 'mls_num_sens={}'.format(args.m4_mls_num_sens)]
    m4_cmd += ['-D', 'mls_num_cats={}'.format(args.m4_mls_num_cats)]
    m4_cmd += ['-D', 'target_build_variant={}'.format(
        args.m4_target_build_variant)]
    m4_cmd += ['-D', 'target_with_dexpreopt={}'.format(
        args.m4_target_with_dexpreopt)]
    m4_cmd += ['-D', 'target_arch={}'.format(args.m4_target_arch)]
    m4_cmd += ['-D', 'target_with_asan={}'.format(args.m4_target_with_asan)]
    m4_cmd += ['-D', 'target_full_treble={}'.format(args.m4_target_full_treble)]

    # The caller might pass no argument to m4_target_compatible_property.
    compatible_property = args.m4_target_compatible_property
    if not compatible_property:
        compatible_property = ''
    m4_cmd += ['-D', 'target_compatible_property={}'.format(
        compatible_property)]

    if args.m4_target_recovery_defs:
        m4_cmd += args.m4_target_recovery_defs
    m4_cmd += ['-s']
    m4_cmd += args.source_files
    m4_cmd += ['>', output_policy_conf]
    # Using shell=True because of '>' above.
    run_host_command(' '.join(m4_cmd), shell=True)

    # Filters out dontaudit lines to a .dontaudit file.
    file_utils.filter_out_re('dontaudit', output_policy_conf,
                             output_policy_conf + '.dontaudit')

    # Builds the raw CIL from output_policy_conf.
    raw_cil_file = tempfile.NamedTemporaryFile(prefix='raw_policy_',
                                               suffix='.cil')

    checkpolicy_cmd = [args.checkpolicy_env]
    checkpolicy_cmd += [os.path.join(args.android_host_path, 'checkpolicy'),
                        '-C', '-M', '-c', args.policy_vers,
                        '-o', raw_cil_file.name, output_policy_conf]
    # Using shell=True to setup args.checkpolicy_env variables.
    run_host_command(' '.join(checkpolicy_cmd), shell=True)
    file_utils.filter_out_lines([args.reqd_mask], raw_cil_file.name)

    # Builds the output CIL by versioning the above raw CIL.
    output_file = args.output_cil
    file_utils.make_parent_dirs(output_file)

    run_host_command([os.path.join(args.android_host_path, 'version_policy'),
                      '-b', args.base_policy, '-t', raw_cil_file.name,
                      '-n', args.treble_sepolicy_vers, '-o', output_file])
    if args.filter_out_files:
        file_utils.filter_out_lines(args.filter_out_files, output_file)

    # Tests that the output file can be merged with the given CILs.
    if args.dependent_cils:
        merge_cmd = [os.path.join(args.android_host_path, 'secilc'),
                     '-m', '-M', 'true', '-G', '-N', '-c', args.policy_vers]
        merge_cmd += args.dependent_cils      # the give CILs to merge
        merge_cmd += [output_file, '-o', '/dev/null', '-f', '/dev/null']
        run_host_command(merge_cmd)


def setup_build_cil(subparsers):
    """Sets up command args for 'build_cil' command."""

    # Required arguments.
    parser = subparsers.add_parser('build_cil', help='build CIL files')
    parser.add_argument('-s', '--source_files', nargs='+', required=True,
                        help='source policy.conf')
    parser.add_argument('-m', '--reqd_mask', required=True,
                        help='the bare minimum policy.conf to use checkpolicy')
    parser.add_argument('-b', '--base_policy', required=True,
                        help='base policy for versioning')
    parser.add_argument('-t', '--treble_sepolicy_vers', required=True,
                        help='the version number to use for Treble-OTA')
    parser.add_argument('-v', '--policy_vers', required=True,
                        help='SELinux policy version')
    parser.add_argument('-p', '--output_policy_conf', required=True,
                        help='the output policy conf in source format')
    parser.add_argument('-o', '--output_cil', required=True,
                        help='the output cil file')

    # 'm4' required arguments.
    parser.add_argument('--m4_mls_num_sens', required=True,
                        help='the value of macro mls_num_sens')
    parser.add_argument('--m4_mls_num_cats', required=True,
                        help='the value of macro mls_num_cats')
    parser.add_argument('--m4_target_build_variant', required=True,
                        help='the value of macro target_build_variant')
    parser.add_argument('--m4_target_with_dexpreopt', required=True,
                        help='the value of macro target_with_dexpreopt')
    parser.add_argument('--m4_target_arch', required=True,
                        help='the value of macro target_arch')
    parser.add_argument('--m4_target_with_asan', required=True,
                        help='the value of macro target_with_asan')
    parser.add_argument('--m4_target_full_treble', required=True,
                        help='the value of macro target_full_treble')
    parser.add_argument('--m4_target_compatible_property', nargs='?',
                        help='the value of macro target_compatible_property')
    # 'm4' optional arguments.
    parser.add_argument('--m4_additional_defs', nargs='*',
                        help='board-specific macro definitions')
    parser.add_argument('--m4_target_recovery_defs', nargs='*',
                        help='the value of macro target_recovery_defs')

    # Optional arguments.
    parser.add_argument('-c', '--checkpolicy_env',
                        help='environment variables passed to checkpolicy')
    parser.add_argument('-f', '--filter_out_files', nargs='+',
                        help='the pattern files to filter out the output cil')
    parser.add_argument('-d', '--dependent_cils', nargs='+',
                        help=('check the output file can be merged with '
                              'the dependent cil files'))

    # The function that performs the actual works.
    parser.set_defaults(func=do_build_cil)


def run(argv):
    """Sets up command parser and execuates sub-command."""
    parser = argparse.ArgumentParser()

    # Adds top-level arguments.
    parser.add_argument('-a', '--android_host_path', default='',
                        help='a path to host out executables')

    # Adds subparsers for each COMMAND.
    subparsers = parser.add_subparsers(title='COMMAND')
    for command in _SUPPORTED_COMMANDS:
        globals()['setup_' + command](subparsers)

    args = parser.parse_args(argv[1:])
    args.func(args)


if __name__ == '__main__':
    run(sys.argv)
