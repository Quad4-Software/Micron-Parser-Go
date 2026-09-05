/* Copyright Quad4 2026
 * SPDX-License-Identifier: 0BSD
 *
 * Stable C ABI for libmicron. Do not use the auto-generated c-shared header.
 * Strings returned by micron_* must be freed with micron_free.
 */

#ifndef MICRON_H
#define MICRON_H

#ifdef __cplusplus
extern "C" {
#endif

#if defined(_WIN32) || defined(__CYGWIN__)
#  ifdef MICRON_BUILD
#    define MICRON_API __declspec(dllexport)
#  else
#    define MICRON_API __declspec(dllimport)
#  endif
#else
#  define MICRON_API __attribute__((visibility("default")))
#endif

/*
 * Convert Micron markup to an HTML fragment.
 * dark_theme and force_monospace are boolean (0 / non-zero).
 * Returns a heap string the caller must free with micron_free.
 */
MICRON_API char *micron_convert(const char *markup, int dark_theme, int force_monospace);

/*
 * Parse leading #!fg= / #!bg= header tags.
 * Returns JSON object {"fg":"...","bg":"..."} (values may be empty).
 * Caller must micron_free the result.
 */
MICRON_API char *micron_parse_header_tags(const char *markup);

/*
 * Collect form field values from a JSON array of objects:
 * [{"type":"text","name":"user","value":"alice","checked":false}, ...]
 * Returns a JSON object mapping name to value. Caller must micron_free.
 */
MICRON_API char *micron_collect_form_fields(const char *inputs_json);

/*
 * Build a Micron-style request payload.
 * fields_json is a JSON object of name to value.
 * Returns JSON {"destination","fields","request_vars"}. Caller must micron_free.
 */
MICRON_API char *micron_build_request_payload(
    const char *fields_json,
    const char *destination,
    const char *fields_spec);

/*
 * Lint Micron markup. Returns a JSON array of diagnostics:
 * [{"severity":0|1|2,"code":"...","message":"...","span":{"start":0,"end":0}}, ...]
 * Caller must micron_free.
 */
MICRON_API char *micron_lint(const char *markup);

/*
 * Parse Micron markup to a JSON Document IR.
 * dark_theme and force_monospace are boolean (0 / non-zero).
 * Caller must micron_free.
 */
MICRON_API char *micron_parse_json(const char *markup, int dark_theme, int force_monospace);

/* Free a string returned by any micron_* function above. */
MICRON_API void micron_free(char *ptr);

#ifdef __cplusplus
}
#endif

#endif /* MICRON_H */
