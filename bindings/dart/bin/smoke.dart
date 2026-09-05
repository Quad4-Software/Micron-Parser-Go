// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

import 'package:micron/micron.dart';

void main() {
  final html = convert(
    '> Title\n\nHello <world> & `*bold`*.\n',
    darkTheme: true,
    forceMonospace: false,
  );
  if (!html.contains('Hello') ||
      !html.contains('&lt;world&gt;') ||
      !html.contains('bold')) {
    throw StateError('convert failed');
  }

  final colors = parseHeaderTags('#!fg=ccc\n#!bg=222\n\nBody\n');
  if (colors['fg'] != 'ccc' || colors['bg'] != '222') {
    throw StateError('headers failed');
  }

  final fields = collectFormFields([
    {'type': 'text', 'name': 'user', 'value': 'alice', 'checked': false},
    {'type': 'checkbox', 'name': 'opts', 'value': '1', 'checked': true},
  ]);
  if (fields['user'] != 'alice' || fields['opts'] != '1') {
    throw StateError('fields failed');
  }

  final payload = buildRequestPayload(fields, '/page`x=1', 'user|opts');
  if (payload['destination'] != '/page') {
    throw StateError('payload failed');
  }

  print('dart smoke ok');
}
