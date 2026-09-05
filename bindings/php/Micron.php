<?php
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

declare(strict_types=1);

/**
 * Micron markup parser bindings over libmicron (PHP FFI).
 */
final class Micron
{
    private static ?FFI $ffi = null;

    private static function libName(): string
    {
        return match (PHP_OS_FAMILY) {
            'Darwin' => 'libmicron.dylib',
            'Windows' => 'libmicron.dll',
            default => 'libmicron.so',
        };
    }

    private static function resolveLib(): string
    {
        $env = getenv('MICRON_LIB_PATH');
        if (is_string($env) && $env !== '' && is_file($env)) {
            return $env;
        }

        $name = self::libName();
        $here = __DIR__;
        $candidates = [
            $here . '/native/' . $name,
            $here . '/../../dist/' . $name,
            getcwd() . '/dist/' . $name,
        ];
        foreach ($candidates as $path) {
            if (is_file($path)) {
                return $path;
            }
        }
        throw new RuntimeException(
            'libmicron not found; set MICRON_LIB_PATH or place the shared library under bindings/php/native or dist/'
        );
    }

    private static function ffi(): FFI
    {
        if (self::$ffi !== null) {
            return self::$ffi;
        }
        if (!extension_loaded('ffi')) {
            throw new RuntimeException('PHP FFI extension is required (enable extension=ffi and ffi.enable=true)');
        }
        $cdef = <<<'CDEF'
char *micron_convert(const char *markup, int dark_theme, int force_monospace);
char *micron_parse_header_tags(const char *markup);
char *micron_collect_form_fields(const char *inputs_json);
char *micron_build_request_payload(const char *fields_json, const char *destination, const char *fields_spec);
void micron_free(char *ptr);
CDEF;
        self::$ffi = FFI::cdef($cdef, self::resolveLib());
        return self::$ffi;
    }

    private static function takeString(?FFI\CData $ptr): string
    {
        if ($ptr === null) {
            return '';
        }
        try {
            return FFI::string($ptr);
        } finally {
            self::ffi()->micron_free($ptr);
        }
    }

    public static function convert(string $markup, bool $darkTheme = true, bool $forceMonospace = true): string
    {
        return self::takeString(
            self::ffi()->micron_convert($markup, $darkTheme ? 1 : 0, $forceMonospace ? 1 : 0)
        );
    }

    /** @return array{fg: string, bg: string} */
    public static function parseHeaderTags(string $markup): array
    {
        $data = json_decode(self::takeString(self::ffi()->micron_parse_header_tags($markup)), true);
        return [
            'fg' => (string)($data['fg'] ?? ''),
            'bg' => (string)($data['bg'] ?? ''),
        ];
    }

    /** @param list<array<string, mixed>> $inputs */
    public static function collectFormFields(array $inputs): array
    {
        $json = json_encode($inputs, JSON_THROW_ON_ERROR);
        $data = json_decode(self::takeString(self::ffi()->micron_collect_form_fields($json)), true);
        return is_array($data) ? $data : [];
    }

    /** @param array<string, string> $fields */
    public static function buildRequestPayload(array $fields, string $destination, string $fieldsSpec): array
    {
        $json = json_encode($fields, JSON_THROW_ON_ERROR);
        $data = json_decode(
            self::takeString(
                self::ffi()->micron_build_request_payload($json, $destination, $fieldsSpec)
            ),
            true
        );
        return is_array($data) ? $data : [];
    }
}
