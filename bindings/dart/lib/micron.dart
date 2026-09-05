// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

import 'dart:convert';
import 'dart:ffi';
import 'dart:io';

import 'package:ffi/ffi.dart';

typedef _ConvertNative = Pointer<Utf8> Function(Pointer<Utf8>, Int32, Int32);
typedef _ConvertDart = Pointer<Utf8> Function(Pointer<Utf8>, int, int);

typedef _ParseHeaderNative = Pointer<Utf8> Function(Pointer<Utf8>);
typedef _ParseHeaderDart = Pointer<Utf8> Function(Pointer<Utf8>);

typedef _CollectFieldsNative = Pointer<Utf8> Function(Pointer<Utf8>);
typedef _CollectFieldsDart = Pointer<Utf8> Function(Pointer<Utf8>);

typedef _BuildPayloadNative = Pointer<Utf8> Function(
  Pointer<Utf8>,
  Pointer<Utf8>,
  Pointer<Utf8>,
);
typedef _BuildPayloadDart = Pointer<Utf8> Function(
  Pointer<Utf8>,
  Pointer<Utf8>,
  Pointer<Utf8>,
);

typedef _FreeNative = Void Function(Pointer<Utf8>);
typedef _FreeDart = void Function(Pointer<Utf8>);

String _libName() {
  if (Platform.isMacOS) return 'libmicron.dylib';
  if (Platform.isWindows) return 'libmicron.dll';
  return 'libmicron.so';
}

DynamicLibrary _openLib() {
  final env = Platform.environment['MICRON_LIB_PATH'];
  if (env != null && env.isNotEmpty && File(env).existsSync()) {
    return DynamicLibrary.open(env);
  }
  final name = _libName();
  final scriptDir = File(Platform.script.toFilePath()).parent.parent.path;
  final candidates = <String>[
    '$scriptDir/native/$name',
    '$scriptDir/../../dist/$name',
    'native/$name',
    '../../dist/$name',
    'dist/$name',
  ];
  for (final path in candidates) {
    if (File(path).existsSync()) {
      return DynamicLibrary.open(File(path).absolute.path);
    }
  }
  throw StateError(
    'libmicron not found; set MICRON_LIB_PATH or place the shared library under bindings/dart/native or dist/',
  );
}

final DynamicLibrary _lib = _openLib();

final _ConvertDart _convertFn =
    _lib.lookupFunction<_ConvertNative, _ConvertDart>('micron_convert');
final _ParseHeaderDart _parseHeaderFn = _lib
    .lookupFunction<_ParseHeaderNative, _ParseHeaderDart>(
      'micron_parse_header_tags',
    );
final _CollectFieldsDart _collectFieldsFn = _lib
    .lookupFunction<_CollectFieldsNative, _CollectFieldsDart>(
      'micron_collect_form_fields',
    );
final _BuildPayloadDart _buildPayloadFn = _lib
    .lookupFunction<_BuildPayloadNative, _BuildPayloadDart>(
      'micron_build_request_payload',
    );
final _FreeDart _freeFn =
    _lib.lookupFunction<_FreeNative, _FreeDart>('micron_free');

String _takeString(Pointer<Utf8> ptr) {
  if (ptr == nullptr) return '';
  try {
    return ptr.toDartString();
  } finally {
    _freeFn(ptr);
  }
}

/// Convert Micron markup to an HTML fragment.
String convert(
  String markup, {
  bool darkTheme = true,
  bool forceMonospace = true,
}) {
  final p = markup.toNativeUtf8();
  try {
    return _takeString(
      _convertFn(p, darkTheme ? 1 : 0, forceMonospace ? 1 : 0),
    );
  } finally {
    malloc.free(p);
  }
}

/// Return page colors from leading #!fg= / #!bg= lines.
Map<String, String> parseHeaderTags(String markup) {
  final p = markup.toNativeUtf8();
  try {
    final data =
        jsonDecode(_takeString(_parseHeaderFn(p))) as Map<String, dynamic>;
    return {
      'fg': (data['fg'] ?? '') as String,
      'bg': (data['bg'] ?? '') as String,
    };
  } finally {
    malloc.free(p);
  }
}

/// Collect form field values from input snapshots.
Map<String, String> collectFormFields(List<Map<String, Object?>> inputs) {
  final p = jsonEncode(inputs).toNativeUtf8();
  try {
    final data =
        jsonDecode(_takeString(_collectFieldsFn(p))) as Map<String, dynamic>;
    return data.map((k, v) => MapEntry(k, '$v'));
  } finally {
    malloc.free(p);
  }
}

/// Build a Micron-style request payload.
Map<String, Object?> buildRequestPayload(
  Map<String, String> fields,
  String destination,
  String fieldsSpec,
) {
  final f = jsonEncode(fields).toNativeUtf8();
  final d = destination.toNativeUtf8();
  final s = fieldsSpec.toNativeUtf8();
  try {
    return jsonDecode(_takeString(_buildPayloadFn(f, d, s)))
        as Map<String, Object?>;
  } finally {
    malloc.free(f);
    malloc.free(d);
    malloc.free(s);
  }
}
