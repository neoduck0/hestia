#!/usr/bin/env bash
set -euo pipefail

project_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
bin_dir=${BIN_DIR:-/usr/local/bin}
bin_name=${BIN_NAME:-hst}
build_dir=$(mktemp -d)
trap 'rm -rf "$build_dir"' EXIT

cd "$project_dir"
go build -trimpath -o "$build_dir/$bin_name" ./src

if [[ -d $bin_dir && -w $bin_dir ]]; then
	install -m 0755 "$build_dir/$bin_name" "$bin_dir/$bin_name"
elif [[ ! -e $bin_dir && -w $(dirname -- "$bin_dir") ]]; then
	install -d "$bin_dir"
	install -m 0755 "$build_dir/$bin_name" "$bin_dir/$bin_name"
else
	sudo install -d "$bin_dir"
	sudo install -m 0755 "$build_dir/$bin_name" "$bin_dir/$bin_name"
fi

printf 'installed %s to %s\n' "$bin_name" "$bin_dir/$bin_name"
