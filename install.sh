#!/bin/sh
# shellcheck shell=dash

set -u

GITHUB_ORG=laciferin2024
GITHUB_REPO=bacalhau

PRE_RELEASE=${PRE_RELEASE:=false}
B_HTTP_REQUEST_CLI=${B_HTTP_REQUEST_CLI:="curl"}



version=${BACALHAU_VERSION:=""} #v0.701.0

getLatestRelease() {

    # /latest ignores pre-releases, see https://docs.github.com/en/rest/releases/releases#get-the-latest-release
    # local tag_regex='v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+)*)' # not cactching -alpha.1
    local tag_regex='v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)*'
    
    if [ "$PRE_RELEASE" = "true" ]; then
        echo "Installing most recent pre-release version..."
        local bacalhauReleaseUrl="https://api.github.com/repos/${GITHUB_ORG}/${GITHUB_REPO}/releases?per_page=1"
    else
        local bacalhauReleaseUrl="https://api.github.com/repos/${GITHUB_ORG}/${GITHUB_REPO}/releases/latest"
    fi
    local latest_release=""

    echo "bacalhau release url $bacalhauReleaseUrl"

    if [ "$B_HTTP_REQUEST_CLI" = "curl" ]; then
        echo "using curl"
        latest_release=$(curl -s $bacalhauReleaseUrl  | grep \"tag_name\" | grep -E -i "\"$tag_regex\"" | awk 'NR==1{print $2}' | sed -n 's/\"\(.*\)\",/\1/p')
    else
        echo "using wget"
        latest_release=$(wget -q --header="Accept: application/json" -O - $bacalhauReleaseUrl | grep \"tag_name\" | grep -E -i "^$tag_regex$" | awk 'NR==1{print $2}' |  sed -n 's/\"\(.*\)\",/\1/p')
    fi
    echo "latest release is $latest_release"
    version=${latest_release}
    echo "$latest_release"
}

detect_os_info() {
  OSARCH=$(uname -m | awk '{if ($0 ~ /arm64|aarch64/) print "arm64"; else if ($0 ~ /x86_64|amd64/) print "amd64"; else print "unsupported_arch"}') && export OSARCH
  OSNAME=$(uname -s | awk '{if ($1 == "Darwin") print "darwin"; else if ($1 == "Linux") print "linux"; else print "unsupported_os"}') && export OSNAME;

  if  [ "$OSNAME" = "unsupported_os" ] || [ "$OSARCH" = "unsupported_arch" ]; then
    echo "Unsupported OS or architecture"
    echo "Checkout if our latest releases support $OSARCH_$OSNAME: https://github.com/${GITHUB_ORG}/${GITHUB_REPO}/releases/latest"
    exit 1
  fi
 
} 

install() {
  LOC=${LOC:-"/tmp/b"}
  INSTALL_LOC=${INSTALL_LOC:-"/usr/local/bin/bacalhau"}

  if [[ "$version" == "" ]]; then
      getLatestRelease
  fi
  echo "installing bacalhau:$version"
  rurl=https://github.com/${GITHUB_ORG}/${GITHUB_REPO}/releases/download/$version/bacalhau-$OSNAME-$OSARCH
  
  echo "release url=$rurl"
  curl -sSL -o $LOC $rurl
  chmod +x $LOC

  cp $LOC $INSTALL_LOC 

  if command -v bacalhau >/dev/null 2>&1; then
    echo "Installed bacalhau successfully!" in "$INSTALL_LOC"
  else
    echo "Bacalhau installation failed or not found in PATH"
  fi

}

main() {
  detect_os_info

  install


  cat << EOF
        _______ _    _          _   _ _  __ __     ______  _    _
        |__   __| |  | |   /\   | \ | | |/ / \ \   / / __ \| |  | |
          | |  | |__| |  /  \  |  \| |   /   \ \_/ / |  | | |  | |
          | |  |  __  | / /\ \ |     |  <     \   /| |  | | |  | |
          | |  | |  | |/ ____ \| |\  |   \     | | | |__| | |__| |
          |_|  |_|  |_/_/    \_\_| \_|_|\_\    |_|  \____/ \____/

      Thanks for installing Bacalhau! We're hoping to unlock an new world of more efficient AI Applications, and would really love to hear from you on how we can improve.

      - ⭐️ Give us a star on GitHub (https://github.com/laciferin2024/bacalhau)
      - 🧑‍💻 Request a feature! (https://github.com/laciferin2024/bacalhau/issues/new)
      - 🐛 File a bug! (https://github.com/laciferin2024/bacalhau/issues/new)
      - ❓ Join our Community! (https://t.me/decenterai)
      - 📰 Checkout our docs! (https://decenter-ai.gitbook.io/)

      Thanks again!
      ~ Team DecenterAI
EOF
}


main "$@" || exit  1
