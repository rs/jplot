#!/bin/bash

set -e

VERSION=$1
USER=rs
NAME=jplot
DESC="iTerm2 expvar/JSON monitoring tool"

if [ -z $VERSION ]; then
  echo "usage: $0 VERSION"
  exit 1
fi

github-release release -u $USER -r $NAME -t $VERSION

rm -rf dist
mkdir dist
cleanup() {
  rm -rf dist
}
trap cleanup EXIT

for env in darwin/amd64 darwin/386; do
    eval $(echo $env | tr '/' ' ' | xargs printf 'export GOOS=%s; export GOARCH=%s\n')

    GOOS=${env%/*}
    GOARCH=${env#*/}

    bin=$NAME
    if [ $GOOS == "windows" ]; then
        bin="$NAME.exe"
    fi

    mkdir -p dist

    echo "Building for GOOS=$GOOS GOARCH=$GOARCH"

    CGO_ENABLED=0 go build -o dist/$bin
    file=${NAME}_${VERSION}_${GOOS}_${GOARCH}.zip
    zip dist/$file -j dist/$bin
    rm -f dist/$bin

    github-release upload -u $USER -r $NAME -t $VERSION -n $file -f dist/$file
done

url=https://github.com/${USER}/${NAME}/archive/${VERSION}.tar.gz
darwin_amd64=${NAME}_${VERSION}_darwin_amd64.zip
darwin_386=${NAME}_${VERSION}_darwin_386.zip

cat << EOF > jplot.rb
class Jplot < Formula
  desc "$DESC"
  homepage "https://github.com/${USER}/${NAME}"
  url "$url"
  sha256 "$(curl $url | shasum -a 256 | awk '{print $1}')"
  head "https://github.com/${USER}/${NAME}.git"

  if Hardware::CPU.is_64_bit?
    url "https://github.com/${USER}/${NAME}/releases/download/${VERSION}/${darwin_amd64}"
    sha256 "$(shasum -a 256 dist/${darwin_amd64} | awk '{print $1}')"
  else
    url "https://github.com/${USER}/${NAME}/releases/download/${VERSION}/${darwin_386}"
    sha256 "$(shasum -a 256 dist/${darwin_386} | awk '{print $1}')"
  end

  depends_on "go" => :build

  def install
    bin.install "$NAME"
  end
end
EOF
