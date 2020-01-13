#!/bin/bash
# Downloads the Go sources, building the assembler into a standalone package.
# Cleanup: rm -rf $GOPATH/src/github.com/twitchyliquid64/golang-asm/{obj,objabi,sys,src,dwarf}

BASE_PKG_PATH="github.com/twitchyliquid64/golang-asm"

if [[ "$GOPATH" == "" ]]; then
  echo "Error: GOPATH must be set."
  exit 1
fi

TMP_PATH="${GOPATH}/src/${BASE_PKG_PATH}/tmp"

if [[ -d "$TMP_PATH" ]]; then
  echo "Deleting old output."
  rm -rf "$TMP_PATH"
fi

mkdir -pv $TMP_PATH
cd $TMP_PATH
git clone https://github.com/golang/go

# Move obj.
cp -rv ${TMP_PATH}/go/src/cmd/internal/obj ${GOPATH}/src/${BASE_PKG_PATH}/obj
find ${GOPATH}/src/${BASE_PKG_PATH}/obj -type f -exec sed -i "s_\"cmd/internal/obj_\"${BASE_PKG_PATH}/obj_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/obj -type f -exec sed -i "s_\"cmd/internal/dwarf_\"${BASE_PKG_PATH}/dwarf_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/obj -type f -exec sed -i "s_\"cmd/internal/src_\"${BASE_PKG_PATH}/src_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/obj -type f -exec sed -i "s_\"cmd/internal/sys_\"${BASE_PKG_PATH}/sys_g" {} \;
# Move objabi.
cp -rv ${TMP_PATH}/go/src/cmd/internal/objabi ${GOPATH}/src/${BASE_PKG_PATH}/objabi
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s_\"cmd/internal/obj_\"${BASE_PKG_PATH}/obj_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s_\"cmd/internal/dwarf_\"${BASE_PKG_PATH}/dwarf_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s_\"cmd/internal/src_\"${BASE_PKG_PATH}/src_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s_\"cmd/internal/sys_\"${BASE_PKG_PATH}/sys_g" {} \;
# Move arch.
mkdir -pv ${GOPATH}/src/${BASE_PKG_PATH}/asm
cp -rv ${TMP_PATH}/go/src/cmd/asm/internal/arch ${GOPATH}/src/${BASE_PKG_PATH}/asm/arch
find ${GOPATH}/src/${BASE_PKG_PATH}/asm/arch -type f -exec sed -i "s_\"cmd/internal/obj_\"${BASE_PKG_PATH}/obj_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/asm/arch -type f -exec sed -i "s_\"cmd/internal/dwarf_\"${BASE_PKG_PATH}/dwarf_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/asm/arch -type f -exec sed -i "s_\"cmd/internal/src_\"${BASE_PKG_PATH}/src_g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/asm/arch -type f -exec sed -i "s_\"cmd/internal/sys_\"${BASE_PKG_PATH}/sys_g" {} \;

# Move dwarf.
cp -rv ${TMP_PATH}/go/src/cmd/internal/dwarf ${GOPATH}/src/${BASE_PKG_PATH}/dwarf
find ${GOPATH}/src/${BASE_PKG_PATH}/dwarf -type f -exec sed -i "s_\"cmd/internal/obj_\"${BASE_PKG_PATH}/obj_g" {} \;
# Move src.
cp -rv ${TMP_PATH}/go/src/cmd/internal/src ${GOPATH}/src/${BASE_PKG_PATH}/src
# Move sys.
cp -rv ${TMP_PATH}/go/src/cmd/internal/sys ${GOPATH}/src/${BASE_PKG_PATH}/sys


# Rewrite identifiers for generated (at build time) constants.
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/stackGuardMultiplierDefault/1/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/defaultGOOS/\"linux\"/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/defaultGOARCH/\"$(go env GOARCH)\"/g" {} \;

find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/defaultGO386/\"\"/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/defaultGOARM/\"7\"/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/defaultGOMIPS64/\"hardfloat\"/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/defaultGOMIPS/\"hardfloat\"/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/= version/= \"\"/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/defaultGO_EXTLINK_ENABLED/\"\"/g" {} \;
find ${GOPATH}/src/${BASE_PKG_PATH}/objabi -type f -exec sed -i "s/goexperiment/\"\"/g" {} \;


# Remove tests (they have package dependencies we could do without).
find ${GOPATH}/src/${BASE_PKG_PATH} -name "*_test.go" -type f -delete

# Remove temporary folder.
rm -rf ${TMP_PATH}

# Write README.
cat > ${GOPATH}/src/${BASE_PKG_PATH}/README.md << "EOF"
# golang-asm

A mirror of the assembler from the Go compiler, with import paths re-written for the assembler to be functional as a standalone library.

License as per the Go project.
EOF

# Write license file.
cat > ${GOPATH}/src/${BASE_PKG_PATH}/LICENSE << "EOF"
Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOTto be standalone
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
EOF
