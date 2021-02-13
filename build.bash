#!/bin/bash
# apt install build-essential gcc-aarch64-linux-gnu gcc-arm-linux-gnueabi gcc-arm-linux-gnueabihf gcc-mipsel-linux-gnu gcc-mips64el-linux-gnuabi64 linux-libc-dev-i386-cross gcc-mips-linux-gnu pbzip2
set -x
KCPVPN_CC=""
KCPVPN_CFLAGS=""
KCPVPN_LDFLAGS=""
function set_cc_for_architecture() {
  case $1 in
  "386")
    KCPVPN_CC="gcc"
    KCPVPN_CFLAGS="-m32 -L/usr/lib32 -I/usr/i686-linux-gnu/include"
    KCPVPN_LDFLAGS="-m32 -L/usr/lib32"
    ;;
  "amd64")
    KCPVPN_CC="gcc"
    KCPVPN_CFLAGS=""
    KCPVPN_LDFLAGS=""
    ;;
  "arm")
    KCPVPN_CC="arm-linux-gnueabihf-gcc"
    KCPVPN_CFLAGS=""
    KCPVPN_LDFLAGS=""
    ;;
  "arm64")
    KCPVPN_CC="aarch64-linux-gnu-gcc"
    KCPVPN_CFLAGS=""
    KCPVPN_LDFLAGS=""
    ;;
  "mips")
    KCPVPN_CC="mips-linux-gnu-gcc"
    KCPVPN_CFLAGS=""
    KCPVPN_LDFLAGS=""
    ;;
  "mipsle")
    KCPVPN_CC="mipsel-linux-gnu-gcc"
    KCPVPN_CFLAGS=""
    KCPVPN_LDFLAGS=""
    ;;
  "mips64le")
    KCPVPN_CC="mips64el-linux-gnuabi64-gcc"
    KCPVPN_CFLAGS=""
    KCPVPN_LDFLAGS=""
    ;;
  *)
    echo "unknown archicture"
    exit 1
    ;;
  esac
}

BUILD_DIR="kcpvpn-build"
ARCHITECTURES="${1}"
if [ "${ARCHITECTURES}" == "" ]; then
  ARCHITECTURES="386 amd64 arm arm64 mips mipsle mips64le"
fi
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}
pushd ${BUILD_DIR} || exit 1

git clone https://github.com/yzsme/libtuntap4go.git || exit 1
wget -O- https://git.kernel.org/pub/scm/linux/kernel/git/shemminger/bridge-utils.git/snapshot/bridge-utils-1.6.tar.gz | tar -zxvf- || exit 1
echo 'LyogbGliYnJpZGdlL2NvbmZpZy5oLiAgR2VuZXJhdGVkIGZyb20gY29uZmlnLmguaW4gYnkgY29u
ZmlndXJlLiAgKi8KLyogbGliYnJpZGdlL2NvbmZpZy5oLmluLiAgR2VuZXJhdGVkIGZyb20gY29u
ZmlndXJlLmluIGJ5IGF1dG9oZWFkZXIuICAqLwoKLyogRGVmaW5lIHRvIDEgaWYgeW91IGhhdmUg
dGhlIGBnZXRob3N0bmFtZScgZnVuY3Rpb24uICovCiNkZWZpbmUgSEFWRV9HRVRIT1NUTkFNRSAx
CgovKiBEZWZpbmUgdG8gMSBpZiB5b3UgaGF2ZSB0aGUgYGlmX2luZGV4dG9uYW1lJyBmdW5jdGlv
bi4gKi8KI2RlZmluZSBIQVZFX0lGX0lOREVYVE9OQU1FIDEKCi8qIERlZmluZSB0byAxIGlmIHlv
dSBoYXZlIHRoZSBgaWZfbmFtZXRvaW5kZXgnIGZ1bmN0aW9uLiAqLwojZGVmaW5lIEhBVkVfSUZf
TkFNRVRPSU5ERVggMQoKLyogRGVmaW5lIHRvIDEgaWYgeW91IGhhdmUgdGhlIDxpbnR0eXBlcy5o
PiBoZWFkZXIgZmlsZS4gKi8KI2RlZmluZSBIQVZFX0lOVFRZUEVTX0ggMQoKLyogRGVmaW5lIHRv
IDEgaWYgeW91IGhhdmUgdGhlIDxtZW1vcnkuaD4gaGVhZGVyIGZpbGUuICovCiNkZWZpbmUgSEFW
RV9NRU1PUllfSCAxCgovKiBEZWZpbmUgdG8gMSBpZiB5b3UgaGF2ZSB0aGUgYHNvY2tldCcgZnVu
Y3Rpb24uICovCiNkZWZpbmUgSEFWRV9TT0NLRVQgMQoKLyogRGVmaW5lIHRvIDEgaWYgeW91IGhh
dmUgdGhlIDxzdGRpbnQuaD4gaGVhZGVyIGZpbGUuICovCiNkZWZpbmUgSEFWRV9TVERJTlRfSCAx
CgovKiBEZWZpbmUgdG8gMSBpZiB5b3UgaGF2ZSB0aGUgPHN0ZGxpYi5oPiBoZWFkZXIgZmlsZS4g
Ki8KI2RlZmluZSBIQVZFX1NURExJQl9IIDEKCi8qIERlZmluZSB0byAxIGlmIHlvdSBoYXZlIHRo
ZSBgc3RyZHVwJyBmdW5jdGlvbi4gKi8KI2RlZmluZSBIQVZFX1NUUkRVUCAxCgovKiBEZWZpbmUg
dG8gMSBpZiB5b3UgaGF2ZSB0aGUgPHN0cmluZ3MuaD4gaGVhZGVyIGZpbGUuICovCiNkZWZpbmUg
SEFWRV9TVFJJTkdTX0ggMQoKLyogRGVmaW5lIHRvIDEgaWYgeW91IGhhdmUgdGhlIDxzdHJpbmcu
aD4gaGVhZGVyIGZpbGUuICovCiNkZWZpbmUgSEFWRV9TVFJJTkdfSCAxCgovKiBEZWZpbmUgdG8g
MSBpZiB5b3UgaGF2ZSB0aGUgPHN5cy9pb2N0bC5oPiBoZWFkZXIgZmlsZS4gKi8KI2RlZmluZSBI
QVZFX1NZU19JT0NUTF9IIDEKCi8qIERlZmluZSB0byAxIGlmIHlvdSBoYXZlIHRoZSA8c3lzL3N0
YXQuaD4gaGVhZGVyIGZpbGUuICovCiNkZWZpbmUgSEFWRV9TWVNfU1RBVF9IIDEKCi8qIERlZmlu
ZSB0byAxIGlmIHlvdSBoYXZlIHRoZSA8c3lzL3RpbWUuaD4gaGVhZGVyIGZpbGUuICovCiNkZWZp
bmUgSEFWRV9TWVNfVElNRV9IIDEKCi8qIERlZmluZSB0byAxIGlmIHlvdSBoYXZlIHRoZSA8c3lz
L3R5cGVzLmg+IGhlYWRlciBmaWxlLiAqLwojZGVmaW5lIEhBVkVfU1lTX1RZUEVTX0ggMQoKLyog
RGVmaW5lIHRvIDEgaWYgeW91IGhhdmUgdGhlIGB1bmFtZScgZnVuY3Rpb24uICovCiNkZWZpbmUg
SEFWRV9VTkFNRSAxCgovKiBEZWZpbmUgdG8gMSBpZiB5b3UgaGF2ZSB0aGUgPHVuaXN0ZC5oPiBo
ZWFkZXIgZmlsZS4gKi8KI2RlZmluZSBIQVZFX1VOSVNURF9IIDEKCi8qIERlZmluZSB0byB0aGUg
YWRkcmVzcyB3aGVyZSBidWcgcmVwb3J0cyBmb3IgdGhpcyBwYWNrYWdlIHNob3VsZCBiZSBzZW50
LiAqLwojZGVmaW5lIFBBQ0tBR0VfQlVHUkVQT1JUICIiCgovKiBEZWZpbmUgdG8gdGhlIGZ1bGwg
bmFtZSBvZiB0aGlzIHBhY2thZ2UuICovCiNkZWZpbmUgUEFDS0FHRV9OQU1FICJicmlkZ2UtdXRp
bHMiCgovKiBEZWZpbmUgdG8gdGhlIGZ1bGwgbmFtZSBhbmQgdmVyc2lvbiBvZiB0aGlzIHBhY2th
Z2UuICovCiNkZWZpbmUgUEFDS0FHRV9TVFJJTkcgImJyaWRnZS11dGlscyAxLjYiCgovKiBEZWZp
bmUgdG8gdGhlIG9uZSBzeW1ib2wgc2hvcnQgbmFtZSBvZiB0aGlzIHBhY2thZ2UuICovCiNkZWZp
bmUgUEFDS0FHRV9UQVJOQU1FICJicmlkZ2UtdXRpbHMiCgovKiBEZWZpbmUgdG8gdGhlIGhvbWUg
cGFnZSBmb3IgdGhpcyBwYWNrYWdlLiAqLwojZGVmaW5lIFBBQ0tBR0VfVVJMICIiCgovKiBEZWZp
bmUgdG8gdGhlIHZlcnNpb24gb2YgdGhpcyBwYWNrYWdlLiAqLwojZGVmaW5lIFBBQ0tBR0VfVkVS
U0lPTiAiMS42IgoKLyogRGVmaW5lIHRvIDEgaWYgeW91IGhhdmUgdGhlIEFOU0kgQyBoZWFkZXIg
ZmlsZXMuICovCiNkZWZpbmUgU1REQ19IRUFERVJTIDEKCi8qIERlZmluZSB0byAxIGlmIHlvdSBj
YW4gc2FmZWx5IGluY2x1ZGUgYm90aCA8c3lzL3RpbWUuaD4gYW5kIDx0aW1lLmg+LiAqLwojZGVm
aW5lIFRJTUVfV0lUSF9TWVNfVElNRSAxCgovKiBEZWZpbmUgdG8gZW1wdHkgaWYgYGNvbnN0JyBk
b2VzIG5vdCBjb25mb3JtIHRvIEFOU0kgQy4gKi8KLyogI3VuZGVmIGNvbnN0ICovCg==' | base64 -d >bridge-utils-1.6/libbridge/config.h
echo 'cmake_minimum_required(VERSION 3.0)
project(bridge C)

set(CMAKE_C_STANDARD 11)

include_directories(libbridge)

set(source_files libbridge/libbridge_devif.c
        libbridge/libbridge_if.c
        libbridge/libbridge_init.c
        libbridge/libbridge_misc.c)

add_library(bridge SHARED ${source_files})
add_library(bridgeStatic STATIC ${source_files})
set_target_properties(bridgeStatic PROPERTIES OUTPUT_NAME bridge)' >bridge-utils-1.6/CMakeLists.txt

for arch in ${ARCHITECTURES}; do
  set_cc_for_architecture ${arch}
  libtuntap4go_build_dir="libtuntap4go-${arch}"
  mkdir ${libtuntap4go_build_dir}
  pushd ${libtuntap4go_build_dir}
  CC=${KCPVPN_CC} CFLAGS=${KCPVPN_CFLAGS} LDFLAGS=${KCPVPN_LDFLAGS} cmake -DCMAKE_BUILD_TYPE=Release ../libtuntap4go && make || exit 1
  popd

  libbridge_build_dir="libbridge-${arch}"
  mkdir ${libbridge_build_dir}
  pushd ${libbridge_build_dir}
  CC=${KCPVPN_CC} CFLAGS=${KCPVPN_CFLAGS} LDFLAGS=${KCPVPN_LDFLAGS} cmake -DCMAKE_BUILD_TYPE=Release ../bridge-utils-1.6 && make || exit 1
  popd

  CC=${KCPVPN_CC} CGO_ENABLED=1 GOOS=linux GOARCH=${arch} GO111MODULE=on CGO_CFLAGS="-g -O2 -I$(pwd)/libtuntap4go -I$(pwd)/bridge-utils-1.6/libbridge ${KCPVPN_CFLAGS}" CGO_LDFLAGS="-g -O2 -L$(pwd)/${libtuntap4go_build_dir} -L$(pwd)/${libbridge_build_dir}" go build -ldflags "-s -w -linkmode \"external\" -extldflags \"-static ${LDFLAGS}\"" -o kcpvpn-${arch} ../ || exit 1
  pbzip2 kcpvpn-${arch}
done

sha1sum kcpvpn-*.bz2
popd || exit 1
