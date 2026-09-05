// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

'use strict';

const {
  convert,
  parseHeaderTags,
  collectFormFields,
  buildRequestPayload,
} = require('./index.js');

const html = convert('> Title\n\nHello <world> & `*bold`*.\n', true, false);
if (!html.includes('Hello') || !html.includes('&lt;world&gt;') || !html.includes('bold')) {
  throw new Error('convert failed: ' + html.slice(0, 200));
}

const colors = parseHeaderTags('#!fg=ccc\n#!bg=222\n\nBody\n');
if (colors.fg !== 'ccc' || colors.bg !== '222') {
  throw new Error('parseHeaderTags failed: ' + JSON.stringify(colors));
}

const fields = collectFormFields([
  { type: 'text', name: 'user', value: 'alice', checked: false },
  { type: 'checkbox', name: 'opts', value: '1', checked: true },
]);
if (fields.user !== 'alice' || fields.opts !== '1') {
  throw new Error('collectFormFields failed: ' + JSON.stringify(fields));
}

const payload = buildRequestPayload(fields, '/page`x=1', 'user|opts');
if (payload.destination !== '/page' || payload.fields.user !== 'alice') {
  throw new Error('buildRequestPayload failed: ' + JSON.stringify(payload));
}

console.log('node smoke ok');
