#!/bin/sh

install () {

set -eu

PROJECT="go-ecs"
REPO="colinmutter/go-ecs"

GITHUB_URL="https://github.com/${REPO}"
RELEASES_URL="${GITHUB_URL}/releases"
UNAME=$(uname)
if [ "$UNAME" != "Linux" -a "$UNAME" != "Darwin" ] ; then
    echo "Sorry, OS not supported: ${UNAME}. Download binary from ${RELEASES_URL}"
    exit 1
fi

if [ "$UNAME" = "Darwin" ] ; then
  OSX_ARCH=$(uname -m)
  if [ "${OSX_ARCH}" = "x86_64" ] ; then
    PLATFORM="darwin_amd64"
  else
    echo "Sorry, architecture not supported: ${OSX_ARCH}. Download binary from ${RELEASES_URL}"
    exit 1
  fi
elif [ "$UNAME" = "Linux" ] ; then
  LINUX_ARCH=$(uname -m)
  if [ "${LINUX_ARCH}" = "i686" ] ; then
    PLATFORM="linux_386"
  elif [ "${LINUX_ARCH}" = "x86_64" ] ; then
    PLATFORM="linux_amd64"
  else
    echo "Sorry, architecture not supported: ${LINUX_ARCH}. Download binary from ${RELEASES_URL}"
    exit 1
  fi
fi

LATEST=$(curl -s https://api.github.com/repos/${REPO}/tags | grep name | head -n 1 | sed 's/[," ]//g' | cut -d ':' -f 2)
URL="${RELEASES_URL}/download/$LATEST/${PROJECT}_$PLATFORM"

curl -sL ${RELEASES_URL}/download/$LATEST/${PROJECT}_$PLATFORM -o /usr/local/bin/${PROJECT}
chmod +x $_

}

install
