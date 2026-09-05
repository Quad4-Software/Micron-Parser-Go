// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

'use strict';

const fs = require('fs');
const path = require('path');
const koffi = require('koffi');

function defaultLibName() {
  if (process.platform === 'darwin') return 'libmicron.dylib';
  if (process.platform === 'win32') return 'libmicron.dll';
  return 'libmicron.so';
}

function resolveLibPath() {
  if (process.env.MICRON_LIB_PATH && fs.existsSync(process.env.MICRON_LIB_PATH)) {
    return process.env.MICRON_LIB_PATH;
  }
  const name = defaultLibName();
  const candidates = [
    path.join(__dirname, 'native', name),
    path.join(__dirname, '..', '..', 'dist', name),
    path.join(process.cwd(), 'dist', name),
  ];
  for (const p of candidates) {
    if (fs.existsSync(p)) return p;
  }
  throw new Error(
    'libmicron not found; set MICRON_LIB_PATH or place the shared library under bindings/node/native or dist/'
  );
}

const lib = koffi.load(resolveLibPath());

const micron_convert = lib.func('micron_convert', 'void *', ['str', 'int', 'int']);
const micron_parse_header_tags = lib.func('micron_parse_header_tags', 'void *', ['str']);
const micron_collect_form_fields = lib.func('micron_collect_form_fields', 'void *', ['str']);
const micron_build_request_payload = lib.func('micron_build_request_payload', 'void *', [
  'str',
  'str',
  'str',
]);
const micron_free = lib.func('micron_free', 'void', ['void *']);

function takeString(ptr) {
  if (!ptr) return '';
  try {
    return koffi.decode(ptr, 'char', -1);
  } finally {
    micron_free(ptr);
  }
}

function convert(markup, darkTheme = true, forceMonospace = true) {
  return takeString(
    micron_convert(String(markup ?? ''), darkTheme ? 1 : 0, forceMonospace ? 1 : 0)
  );
}

function parseHeaderTags(markup) {
  return JSON.parse(takeString(micron_parse_header_tags(String(markup ?? ''))) || '{"fg":"","bg":""}');
}

function collectFormFields(inputs) {
  return JSON.parse(
    takeString(micron_collect_form_fields(JSON.stringify(inputs || []))) || '{}'
  );
}

function buildRequestPayload(fields, destination, fieldsSpec) {
  return JSON.parse(
    takeString(
      micron_build_request_payload(
        JSON.stringify(fields || {}),
        String(destination ?? ''),
        String(fieldsSpec ?? '')
      )
    ) || '{"destination":"","fields":{},"request_vars":{}}'
  );
}

module.exports = {
  convert,
  parseHeaderTags,
  collectFormFields,
  buildRequestPayload,
};
