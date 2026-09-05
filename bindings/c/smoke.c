/* Copyright Quad4 2026
 * SPDX-License-Identifier: 0BSD
 *
 * Build: gcc -o micron_c_smoke smoke.c -L../../dist -lmicron -Wl,-rpath,'$ORIGIN/../../dist'
 */
#include <stdio.h>
#include <string.h>
#include "micron.h"

int main(void) {
    char *html = micron_convert("> Title\n\nHello <world>\n", 1, 0);
    if (!html || !strstr(html, "Hello") || !strstr(html, "&lt;world&gt;")) {
        fprintf(stderr, "convert failed\n");
        return 1;
    }
    micron_free(html);

    char *colors = micron_parse_header_tags("#!fg=ccc\n#!bg=222\n\nBody\n");
    if (!colors || !strstr(colors, "\"fg\":\"ccc\"")) {
        fprintf(stderr, "headers failed\n");
        return 1;
    }
    micron_free(colors);

    puts("c smoke ok");
    return 0;
}
