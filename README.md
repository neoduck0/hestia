<h1 align="center">
  <img src="assets/logo.svg" alt="Hestia logo" width="160" height="160">
  <br>
  Hestia
</h1>

Hestia (`hst`) is a Linux command-line tool for deploying dotfiles and other
files from a project to their destinations using symlinks or copies. Mappings
are organized into named groups.

## Install

With Go installed, run from this repository:

```sh
./build.sh
```

This installs `hst` system-wide to `/usr/local/bin`. To install for your user
only:

```sh
BIN_DIR="$HOME/.local/bin" ./build.sh
```

Ensure the chosen directory is on your `PATH`.

## Usage

Start from a directory containing your dotfiles. You can organize them as files
and subdirectories, for example:

```text
dotfiles/
├── bashrc
├── gitconfig
└── config/
    └── nvim/
        └── init.lua
```

From inside `dotfiles/`, initialize a project and link a file:

```sh
hst init
hst add bashrc '~/.bashrc' --group shell --create
hst link shell
hst status
```

`init` creates `.hestia/mappings.conf` inside the project directory to store
named groups of source-to-destination mappings. Source files stay in your
project; the example links `bashrc` to `~/.bashrc`.

Use `hst --help` or `hst <command> --help` for available commands and flags.

## License

[MIT License](LICENSE).
