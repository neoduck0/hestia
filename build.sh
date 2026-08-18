#!/usr/bin/env bash
set -euo pipefail

project_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
bin_dir=${BIN_DIR:-/usr/local/bin}
build_dir=$(mktemp -d)
trap 'rm -rf "$build_dir"' EXIT

cd "$project_dir"
go build -trimpath -o "$build_dir/hst" ./src

if [[ -d $bin_dir && -w $bin_dir ]]; then
	install -m 0755 "$build_dir/hst" "$bin_dir/hst"
elif [[ ! -e $bin_dir && -w $(dirname -- "$bin_dir") ]]; then
	install -d "$bin_dir"
	install -m 0755 "$build_dir/hst" "$bin_dir/hst"
else
	sudo install -d "$bin_dir"
	sudo install -m 0755 "$build_dir/hst" "$bin_dir/hst"
fi

printf 'installed hst to %s\n' "$bin_dir/hst"
