// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

import Micron
import Foundation

let html = Micron.convert("> Title\n\nHello <world> & `*bold`*.\n", darkTheme: true, forceMonospace: false)
guard html.contains("Hello"), html.contains("&lt;world&gt;"), html.contains("bold") else {
    fputs("convert failed\n", stderr)
    exit(1)
}

let colors = Micron.parseHeaderTags("#!fg=ccc\n#!bg=222\n\nBody\n")
guard colors["fg"] == "ccc", colors["bg"] == "222" else {
    fputs("headers failed\n", stderr)
    exit(1)
}

let fields = Micron.collectFormFields([
    ["type": "text", "name": "user", "value": "alice", "checked": false],
    ["type": "checkbox", "name": "opts", "value": "1", "checked": true],
])
guard fields["user"] == "alice", fields["opts"] == "1" else {
    fputs("fields failed\n", stderr)
    exit(1)
}

let payload = Micron.buildRequestPayload(fields, destination: "/page`x=1", fieldsSpec: "user|opts")
guard (payload["destination"] as? String) == "/page" else {
    fputs("payload failed\n", stderr)
    exit(1)
}

print("swift smoke ok")
