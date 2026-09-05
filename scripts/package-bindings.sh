#!/usr/bin/env bash
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD
# Package language bindings with an already-built libmicron for one platform.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DIST="${ROOT}/dist"
PLATFORM="${1:?usage: package-bindings.sh <os>-<arch> <path-to-libmicron>}"
LIB_SRC="${2:?usage: package-bindings.sh <os>-<arch> <path-to-libmicron>}"
OUT="${DIST}/packages"
mkdir -p "${OUT}"

case "${PLATFORM}" in
  windows-*) LIB_NAME="libmicron.dll" ;;
  darwin-*) LIB_NAME="libmicron.dylib" ;;
  *) LIB_NAME="libmicron.so" ;;
esac

if [[ ! -f "${LIB_SRC}" ]]; then
  echo "missing library: ${LIB_SRC}" >&2
  exit 1
fi

stage_native() {
  local dest="$1"
  rm -rf "${dest}"
  mkdir -p "${dest}"
  cp "${LIB_SRC}" "${dest}/${LIB_NAME}"
}

# Python wheel
PY_PKG="${ROOT}/bindings/python/micron"
# Clear previous staged libs
rm -f "${PY_PKG}/libmicron.so" "${PY_PKG}/libmicron.dylib" "${PY_PKG}/libmicron.dll"
cp "${LIB_SRC}" "${PY_PKG}/${LIB_NAME}"
(
  cd "${ROOT}/bindings/python"
  VENV="${DIST}/.venv-wheel"
  if [[ ! -d "${VENV}" ]]; then
    python3 -m venv "${VENV}"
  fi
  if [[ -x "${VENV}/bin/python" ]]; then
    VENV_PY="${VENV}/bin/python"
    VENV_PIP="${VENV}/bin/pip"
  elif [[ -x "${VENV}/Scripts/python.exe" ]]; then
    VENV_PY="${VENV}/Scripts/python.exe"
    VENV_PIP="${VENV}/Scripts/pip.exe"
  else
    echo "venv python missing under ${VENV}" >&2
    exit 1
  fi
  "${VENV_PIP}" install --quiet wheel setuptools
  "${VENV_PY}" setup.py bdist_wheel --dist-dir "${OUT}" >/dev/null
)
rm -f "${PY_PKG}/libmicron.so" "${PY_PKG}/libmicron.dylib" "${PY_PKG}/libmicron.dll"
shopt -s nullglob
for whl in "${OUT}"/micron_parser-*.whl "${OUT}"/micron-parser-*.whl; do
  [[ -f "${whl}" ]] || continue
  base="$(basename "${whl}")"
  case "${base}" in
    *-"${PLATFORM}".whl) ;;
    *) mv "${whl}" "${OUT}/micron-parser-${PLATFORM}.whl" ;;
  esac
done
shopt -u nullglob

# Node tarball
NODE_NATIVE="${ROOT}/bindings/node/native"
stage_native "${NODE_NATIVE}"
tar -C "${ROOT}/bindings/node" -czf "${OUT}/micron-parser-node-${PLATFORM}.tar.gz" \
  package.json index.js native

# Java JAR with embedded natives
JAVA_NATIVES="${ROOT}/bindings/java/src/main/resources/natives/${PLATFORM}"
rm -rf "${ROOT}/bindings/java/src/main/resources/natives"
mkdir -p "${JAVA_NATIVES}"
cp "${LIB_SRC}" "${JAVA_NATIVES}/${LIB_NAME}"
chmod +x "${ROOT}/scripts/build-java-jar.sh"
if command -v javac >/dev/null 2>&1 && command -v jar >/dev/null 2>&1; then
  "${ROOT}/scripts/build-java-jar.sh" "${PLATFORM}" "${LIB_SRC}" "${OUT}/micron-parser-${PLATFORM}.jar"
elif command -v mvn >/dev/null 2>&1; then
  (
    cd "${ROOT}/bindings/java"
    mvn -q -DskipTests package
    cp target/micron-parser-*.jar "${OUT}/micron-parser-${PLATFORM}.jar"
  )
fi

# C# nupkg
case "${PLATFORM}" in
  linux-amd64) RID=linux-x64 ;;
  linux-arm64) RID=linux-arm64 ;;
  darwin-amd64) RID=osx-x64 ;;
  darwin-arm64) RID=osx-arm64 ;;
  windows-amd64) RID=win-x64 ;;
  *) RID="${PLATFORM}" ;;
esac
CS_NATIVE="${ROOT}/bindings/csharp/runtimes"
rm -rf "${CS_NATIVE}"
mkdir -p "${CS_NATIVE}/${RID}/native"
cp "${LIB_SRC}" "${CS_NATIVE}/${RID}/native/${LIB_NAME}"
if command -v dotnet >/dev/null 2>&1; then
  (
    cd "${ROOT}/bindings/csharp"
    dotnet pack -c Release -o "${OUT}" -p:PackageVersion=1.1.0
    shopt -s nullglob
    for nupkg in "${OUT}"/Quad4.Micron*.nupkg; do
      mv "${nupkg}" "${OUT}/Quad4.Micron-${PLATFORM}.nupkg"
    done
    shopt -u nullglob
  )
fi

# Ruby gem
RB_NATIVE="${ROOT}/bindings/ruby/native"
stage_native "${RB_NATIVE}"
if command -v gem >/dev/null 2>&1; then
  (
    cd "${ROOT}/bindings/ruby"
    gem build micron-parser.gemspec --output "${OUT}/micron-parser-${PLATFORM}.gem" >/dev/null
  )
fi

# PHP tarball
PHP_NATIVE="${ROOT}/bindings/php/native"
stage_native "${PHP_NATIVE}"
tar -C "${ROOT}/bindings/php" -czf "${OUT}/micron-parser-php-${PLATFORM}.tar.gz" \
  Micron.php native

# Dart tarball
DART_NATIVE="${ROOT}/bindings/dart/native"
stage_native "${DART_NATIVE}"
tar -C "${ROOT}/bindings/dart" -czf "${OUT}/micron-parser-dart-${PLATFORM}.tar.gz" \
  pubspec.yaml lib bin native

# Perl tarball
PERL_NATIVE="${ROOT}/bindings/perl/native"
stage_native "${PERL_NATIVE}"
tar -C "${ROOT}/bindings/perl" -czf "${OUT}/micron-parser-perl-${PLATFORM}.tar.gz" \
  Makefile.PL lib native

# Zig and Swift are source-only (dynamic load at runtime); ship sources with native lib
ZIG_NATIVE="${ROOT}/bindings/zig/native"
stage_native "${ZIG_NATIVE}"
tar -C "${ROOT}/bindings/zig" -czf "${OUT}/micron-parser-zig-${PLATFORM}.tar.gz" \
  build.zig build.zig.zon src native

SWIFT_NATIVE="${ROOT}/bindings/swift/native"
stage_native "${SWIFT_NATIVE}"
tar -C "${ROOT}/bindings/swift" -czf "${OUT}/micron-parser-swift-${PLATFORM}.tar.gz" \
  Package.swift Sources native

cp "${ROOT}/bindings/c/micron.h" "${DIST}/micron.h"
echo "packaged bindings for ${PLATFORM} into ${OUT}"
ls -la "${OUT}" || true
