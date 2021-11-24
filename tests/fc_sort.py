#!/usr/bin/env python

import argparse
import os
import sys


META_CHARS = frozenset(['.', '^', '$', '?', '*', '+', '|', '[', '(', '{'])
ESCAPED_META_CHARS = frozenset(['\.', '\^', '\$', '\?', '\*', '\+', '\|', '\[', '\(', '\{'])


def get_stem_len(path):
    """Returns the length of the stem."""
    stem_len = 0
    i = 0
    while i < len(path):
        if path[i] == "\\":
            i += 1
        elif path[i] in META_CHARS:
            break
        stem_len += 1
        i += 1
    return stem_len


def is_meta(path):
    """Indicates if a path contains any metacharacter."""
    meta_char_count = 0
    escaped_meta_char_count = 0
    for c in META_CHARS:
        if c in path:
            meta_char_count += 1
    for c in ESCAPED_META_CHARS:
        if c in path:
            escaped_meta_char_count += 1
    return meta_char_count > escaped_meta_char_count


class FileContextsNode(object):
    """An entry in a file_context file."""

    def __init__(self, path, file_type, context, meta, stem_len, str_len, line):
        self.path = path
        self.file_type = file_type
        self.context = context
        self.meta = meta
        self.stem_len = stem_len
        self.str_len = str_len
        self.type = context.split(":")[2]
        self.line = line

    @classmethod
    def create(cls, line):
        if (len(line) == 0) or (line[0] == '#'):
            return None

        split = line.split()
        path = split[0].strip()
        context = split[-1].strip()
        file_type = None
        if len(split) == 3:
            file_type = split[1].strip()
        meta = is_meta(path)
        stem_len = get_stem_len(path)
        str_len = len(path.replace("\\", ""))

        return cls(path, file_type, context, meta, stem_len, str_len, line)


def read_file_contexts(file_descriptor):
    file_contexts = []
    for line in file_descriptor:
        node = FileContextsNode.create(line.strip())
        if node != None:
            file_contexts.append(node)
    return file_contexts


def read_multiple_file_contexts(files):
    file_contexts = []
    for filename in files:
        with open(filename) as fd:
            file_contexts.extend(read_file_contexts(fd))
    return file_contexts


# Comparator function for list.sort() based off of fc_sort.c
# Compares two FileContextsNodes a and b and returns 1 if a is more
# specific or -1 if b is more specific.
def compare(a, b):
    # The regex without metachars is more specific
    if a.meta and not b.meta:
        return -1
    if b.meta and not a.meta:
        return 1

    # The regex with longer stem_len (regex before any meta characters) is more specific.
    if a.stem_len < b.stem_len:
        return -1
    if b.stem_len < a.stem_len:
        return 1

    # The regex with longer string length is more specific
    if a.str_len < b.str_len:
        return -1
    if b.str_len < a.str_len:
        return 1

    # A regex with a file_type defined (e.g. file, dir) is more specific.
    if a.file_type is None and b.file_type is not None:
        return -1
    if b.file_type is None and a.file_type is not None:
        return 1

    # Regexes are equally specific.
    return 0


def sort(files):
    for f in files:
        if not os.path.exists(f):
            sys.exit("Error: File_contexts file " + f + " does not exist\n")
    file_contexts = read_multiple_file_contexts(files)
    file_contexts.sort(cmp=compare)
    return file_contexts


def print_fc(Fc, out):
    if not out:
        f = sys.stdout
    else:
        f = open(out, "w")
    for node in Fc:
        f.write(node.line + "\n")


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description="SELinux file_contexts sorting tool.")
    parser.add_argument("-i", dest="input",
            help="Path to the file_contexts file(s).", nargs="?", action='append')
    parser.add_argument("-o", dest="output",
            help="Path to the output file.", nargs=1)
    args = parser.parse_args()
    if not args.input:
        parser.error("Must include path to policy")
    if not not args.output:
        args.output = args.output[0]

    print_fc(sort(args.input), args.output)
