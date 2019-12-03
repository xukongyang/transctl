#!/bin/bash

set -e

SRC=$(realpath $(cd -P "$(dirname "${BASH_SOURCE[0]}")" && pwd))

TAGS=""

pushd $SRC &> /dev/null

BUILD=$SRC/build
VER=$(git tag -l v* $SRC|grep -E '^v[0-9]+\.[0-9]+\.[0-9]+(\.[0-9]+)?$'|sort -r -V|head -1||:)

OPTIND=1
while getopts "b:v:" opt; do
case "$opt" in
  b) BUILD=$OPTARG ;;
  v) VERSION=$OPTARG ;;
esac
done

if [ -z "$VER" ]; then
  VER='v0.0.0-dev'
fi

PLATFORM=$(uname|sed -e 's/_.*//'|tr '[:upper:]' '[:lower:]'|sed -e 's/^\(msys\|mingw\).*/windows/')

NAME=$(basename $SRC)
EXT=tar.bz2
DIR=$BUILD/$PLATFORM/$VER
BIN=$DIR/$NAME

case $PLATFORM in
  windows)
    EXT=zip
    BIN=$BIN.exe
  ;;

  linux|darwin)
  ;;
esac

OUT=$DIR/$NAME-$VER-$PLATFORM-amd64.$EXT

echo "PLATFORM:    $PLATFORM"
echo "VER:         $VER"
#echo "DIR:         $DIR"
#echo "BIN:         $BIN"
#echo "OUT:         $OUT"
echo "TAGS:        $TAGS"

if [ -d $DIR ]; then
  echo "REMOVING:    $DIR"
  rm -rf $DIR
fi

mkdir -p $DIR

echo "BUILDING:    $BIN"
go build \
  -tags "$TAGS" \
  -ldflags="-s -w -X main.version=$VER" \
  -o $BIN

case $PLATFORM in
  linux|windows|darwin)
    echo "STRIPPING:   $BIN"
    strip $BIN
  ;;
esac

case $PLATFORM in
  linux|windows|darwin)
#    echo "COMPRESSING: $BIN"
    COMPRESSED=$(upx -q -q $BIN|awk '{print $1 " -> " $3 " (" $4 ")"}')
    echo "COMPRESSED:  $COMPRESSED"
  ;;
esac

#echo "CHECKING:    $NAME --version"
BUILT_VER=$($BIN --version)
if [ "$BUILT_VER" != "$NAME $VER" ]; then
  echo -e "\n\nerror: expected $NAME --version to report '$NAME $VER', got: '$BUILT_VER'"
  exit 1
fi
echo "REPORTED:    $BUILT_VER"

#echo "PACKING:     $OUT"
case $EXT in
  tar.bz2)
    tar -C $DIR -cjf $OUT $(basename $BIN)
  ;;
  zip)
    zip $OUT -j $BIN
  ;;
esac

echo "PACKED:      $OUT ($(du -sh $OUT|awk '{print $1}'))"

popd &> /dev/null
