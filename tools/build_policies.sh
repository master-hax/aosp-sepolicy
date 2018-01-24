#!/bin/bash

# Ensure that GNU parallel is installed.
# We use this to build multiple targets at the same time.
if [[ -z $(command -v parallel) ]]; then
  echo "Please install GNU Parallel."
  exit
fi

# Calculate the number of targets to build at once.
# TODO: Do a better job of this.  We'd like to use many cores for the initial
# part of the build that runs ckati and ninja, but fewer for mmma, which uses
# multiple cores itself.  We could probably pass -j to mmma to control this.
cores=$(nproc --all)
jobs=8
if [[ $cores -lt $jobs ]]; then
   jobs=2
fi

if [[ $# -lt 2 ]]; then
  echo "Usage: $0 <Android root directory> <output directory> [specific targets to build]"
  exit
fi

android_root_dir=$1
export out_dir=$2
shift 2
targets="$@"

echo "Android tree: $android_root_dir"
echo "Output directory: $out_dir"

mkdir -p $out_dir

cd $android_root_dir
source build/envsetup.sh > /dev/null

# Collect the list of targets by parsing the output of lunch.
# TODO: This misses some targets.
if [[ "$targets" = "" ]]; then
  targets=`lunch 2>/dev/null <<< _ | grep "[0-9]" | sed 's/^.* //'`
fi

echo "Targets: $(echo $targets | paste -sd' ')"

compile_target () {
  target=$1
  source build/envsetup.sh > /dev/null
  lunch $target &> /dev/null
  # Some targets can't lunch properly.
  if [ $? -ne 0 ]; then
    echo "$target cannot be lunched"
    return 1
  fi
  my_out_file="$out_dir/log.$target"
  rm -f $my_out_file
  # Build the policy.
  OUT_DIR=$out_dir/out.$target mmma system/sepolicy &>> $my_out_file
  if [ $? -ne 0 ]; then
    echo "$target failed to build"
    return 2
  fi
  return 0
}
export -f compile_target

parallel --no-notice -j $jobs --bar --joblog $out_dir/joblog compile_target ::: $targets

echo "Failed to lunch: $(grep "\s1\s0\scompile_target" $out_dir/joblog | sed 's/^.* //' | paste -sd' ')"
echo "Failed to build: $(grep "\s2\s0\scompile_target" $out_dir/joblog | sed 's/^.* //' | paste -sd' ')"
