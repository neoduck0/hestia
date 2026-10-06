# Hestia

Hestia (`hst`) is a Linux command-line tool for deploying dotfiles and other
files from a project to their destinations using symlinks or copies. Mappings
are organized into named groups.

## Install

With Go installed, run from this repository:

```sh
./build.sh
```

This installs `hst` system-wide to `/usr/local/bin`. To install for your user only:

```sh
BIN_DIR="$HOME/.local/bin" ./build.sh
```

Ensure the chosen directory is on your `PATH`.

## Quick start

From a directory containing your dotfiles (with an existing `bashrc` file):

```sh
hst init
hst add bashrc '~/.bashrc' --group shell --create
hst link shell
hst status
```

`init` creates `.hestia/mappings.conf`. Commands find the project by searching
the current directory and its ancestors. Relative paths are resolved against
the project directory; `~` expands to the home directory.

Mappings can also be edited directly:

```text
[shell]
"bashrc" -> "~/.bashrc"
```

Sources must exist. Directory sources are deployed file by file, and destination
parent directories are created as needed.

## Commands

- `hst link --all`: link every group.
- `hst link --exclude shell`: link every group except `shell`.
- `hst link shell --copy`: copy files instead of symlinking.
- `hst status`: show linked/total mapping counts and each group's status.
- `hst group list|add|delete|rename`: manage groups. Deleting a group does not
  remove deployed files.

Use `hst --help` or `hst <command> --help` for details, and `--verbose` for debug
output.

## License

[MIT License](LICENSE).
