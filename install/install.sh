#!/bin/sh
#
# Adapted from the pressly/goose installer: Copyright 2021. MIT License.
# Ref: https://github.com/pressly/goose/blob/master/install.sh
#
# Adapted from the Deno installer: Copyright 2019 the Deno authors. All rights reserved. MIT license.
# Ref: https://github.com/denoland/deno_install
#
# TODO(everyone): Keep this script simple and easily auditable.

# Not intended for Windows.

set -e

# source: colors.sh file on render managed envs
# source: https://stackoverflow.com/questions/5947742/how-to-change-the-output-color-of-echo-in-linux
BLUE=$(tput setaf 4)
CYAN=$(tput setaf 6)
BOLD=$(tput bold)
RED=$(tput setaf 1)
RESET=$(tput sgr0)

os=$(uname -s | tr '[:upper:]' '[:lower:]')
arch=$(uname -m)

if [ "$arch" = "aarch64" ]; then
    arch="arm64"
fi

if [ "$arch" = "amd64" ]; then
  arch="x86_64"
fi

if [ "$os" = "darwin" ]; then
    arch="all"
fi

if [ $# -eq 0 ]; then
    splash_uri="https://github.com/joshi4/splash/releases/latest/download/splash_${os}_${arch}"
else
    splash_uri="https://github.com/joshi4/splash/releases/download/${1}/splash_${os}_${arch}"
fi

splash_install="${SPLASH_INSTALL:-$HOME/.splash}"
bin_dir="${splash_install}/bin"
exe="${bin_dir}/splash"

if [ ! -d "${bin_dir}" ]; then
    mkdir -p "${bin_dir}"
fi

curl --silent --show-error --fail --location --output "${exe}" "$splash_uri"
chmod +x "${exe}"

echo
echo "splash was installed successfully to ${exe}"
echo


# defaults to zsh
shell="${SHELL:-'zsh'}"

case :$PATH:
  in *:${bin_dir}*) ;; # do nothing
     *) echo
        echo "${BLUE}${BOLD} Add splash to your PATH:${RESET}"
        echo
        case :$shell:
        in *zsh*) echo "${BLUE}${BOLD} echo 'export PATH=\"$bin_dir:\$PATH\"' >> ~/.zshrc${RESET}";;
          *bash*) echo "${BLUE}${BOLD} echo 'export PATH=\"$bin_dir:\$PATH\"' >> ~/.bashrc${RESET}";;
          *fish*) echo "${BLUE}${BOLD} echo 'fish_add_path -P $bin_dir' >> ~/.config/fish/config.fish${RESET}";;
        esac;;
esac

## for bash check if .bash_profile exists and if the user has sourced bashrc inside it
BASHRC="$HOME/.bashrc"
BASH_PROFILE="$HOME/.bash_profile"
if [ "$shell" = "bash" ]; then
  if [ -f "$BASH_PROFILE" ]; then
    # Look for lines that either use source /path/to/.bashrc or . /path/to/.bashrc, accounting for potential spaces.
    # The command following if is executed, and if its exit status is 0 (which indicates success), the then branch is executed.
    if ! grep -qE "^\s*(source|\.)\s*(.+\.bashrc)" "$BASH_PROFILE"; then
      echo "${BLUE}${BOLD} echo 'source ~/.bashrc' >> ~/.bash_profile${RESET}"
    fi
  fi
fi

case :$shell:
  in  *zsh*) echo "${BLUE}${BOLD} source ~/.zshrc # to pick up the new changes${RESET}";;
      *bash*) echo "${BLUE}${BOLD} source ~/.bashrc # to pick up the new changes${RESET}";;
      *fish*) echo "${BLUE}${BOLD} source ~/.config/fish/config.fish # to pick up the new changes${RESET}";;
esac

echo
echo "Run 'splash --help' or visit https://github.com/joshi4/splash to learn more about splash."
echo
