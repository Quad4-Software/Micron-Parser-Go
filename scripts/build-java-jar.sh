#!/usr/bin/env bash
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD
# Build a shaded micron-parser JAR without Maven (downloads JNA once).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
PLATFORM="${1:?usage: build-java-jar.sh <os>-<arch> <lib-path> <out-jar>}"
LIB_SRC="${2:?}"
OUT_JAR="${3:?}"
CACHE="${ROOT}/dist/.cache"
JNA_VER="5.14.0"
JNA_JAR="${CACHE}/jna-${JNA_VER}.jar"
mkdir -p "${CACHE}"

case "${PLATFORM}" in
  windows-*) LIB_NAME="libmicron.dll" ;;
  darwin-*) LIB_NAME="libmicron.dylib" ;;
  *) LIB_NAME="libmicron.so" ;;
esac

if [[ ! -f "${JNA_JAR}" ]]; then
  curl -fsSL -o "${JNA_JAR}" \
    "https://repo1.maven.org/maven2/net/java/dev/jna/jna/${JNA_VER}/jna-${JNA_VER}.jar"
fi

WORKDIR="$(mktemp -d)"
trap 'rm -rf "${WORKDIR}"' EXIT
mkdir -p "${WORKDIR}/classes" "${WORKDIR}/natives/${PLATFORM}"
cp "${LIB_SRC}" "${WORKDIR}/natives/${PLATFORM}/${LIB_NAME}"

javac -cp "${JNA_JAR}" -d "${WORKDIR}/classes" \
  "${ROOT}/bindings/java/src/main/java/io/quad4/micron/"*.java

# Unpack JNA into classes for a fat jar
(
  cd "${WORKDIR}/classes"
  jar xf "${JNA_JAR}"
  rm -rf META-INF/MANIFEST.MF META-INF/*.SF META-INF/*.RSA META-INF/*.DSA 2>/dev/null || true
)
mkdir -p "${WORKDIR}/classes/natives/${PLATFORM}"
cp "${LIB_SRC}" "${WORKDIR}/classes/natives/${PLATFORM}/${LIB_NAME}"

printf 'Main-Class: io.quad4.micron.Micron\n' > "${WORKDIR}/manifest.txt"
jar cfm "${OUT_JAR}" "${WORKDIR}/manifest.txt" -C "${WORKDIR}/classes" .
echo "wrote ${OUT_JAR}"
