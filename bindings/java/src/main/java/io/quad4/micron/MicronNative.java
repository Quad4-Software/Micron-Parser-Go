// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package io.quad4.micron;

import com.sun.jna.Library;
import com.sun.jna.Pointer;

interface MicronNative extends Library {
    Pointer micron_convert(String markup, int darkTheme, int forceMonospace);

    Pointer micron_parse_header_tags(String markup);

    Pointer micron_collect_form_fields(String inputsJson);

    Pointer micron_build_request_payload(String fieldsJson, String destination, String fieldsSpec);

    void micron_free(Pointer ptr);
}
