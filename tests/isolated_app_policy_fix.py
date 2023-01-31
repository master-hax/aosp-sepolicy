import os
import re
import mmap

def replace_in_file(file_path):
    with open(file_path, 'r') as file:
        content = file.read()
        if "isolated_app_all" in content: return

        # Replaces cases with allow/macro {generic_domain -isolated_app}
        #
        # We don't have to replace existing non-negate allow rules since those
        # will never grant unexpected access to {isolated_app_all -isolated_app}
        content = re.sub(r'\s*-\s*isolated_app', '- isolated_app_all', content)

    with open(file_path, 'w') as file:
        file.write(content)

def replace_in_directory(directory):
    for root, dirs, files in os.walk(directory):
        for file_name in files:
            file_path = os.path.join(root, file_name)
            replace_in_file(file_path)


if __name__ == "__main__":
    import sys
    if len(sys.argv) != 2:
        print("Usage: python replace_isolated_app.py <directory>")
        sys.exit(1)

    directory = sys.argv[1]
    replace_in_directory(directory)
