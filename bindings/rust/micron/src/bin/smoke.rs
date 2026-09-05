// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

use micron::{
    FieldInput, build_request_payload, collect_form_fields, convert, parse_header_tags,
};
use std::collections::HashMap;

fn main() {
    let html = convert("> Title\n\nHello <world> & `*bold`*.\n", true, false).expect("convert");
    assert!(html.contains("Hello"));
    assert!(html.contains("bold"));

    let colors = parse_header_tags("#!fg=ccc\n#!bg=222\n\nBody\n").expect("headers");
    assert_eq!(colors.fg, "ccc");
    assert_eq!(colors.bg, "222");

    let fields = collect_form_fields(&[
        FieldInput {
            type_: "text".into(),
            name: "user".into(),
            value: "alice".into(),
            checked: false,
        },
        FieldInput {
            type_: "checkbox".into(),
            name: "opts".into(),
            value: "1".into(),
            checked: true,
        },
    ])
    .expect("fields");
    assert_eq!(fields.get("user").map(String::as_str), Some("alice"));

    let mut map = HashMap::new();
    map.insert("user".into(), "alice".into());
    let payload = build_request_payload(&map, "/page`x=1", "user").expect("payload");
    assert_eq!(payload.destination, "/page");
    println!("rust smoke ok");
}
