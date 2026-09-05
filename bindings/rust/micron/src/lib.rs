// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

//! Safe Micron bindings over libmicron.

use serde::{Deserialize, Serialize};
use serde_json::{Map, Value};
use std::collections::HashMap;

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct PageColors {
    pub fg: String,
    pub bg: String,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct FieldInput {
    #[serde(rename = "type")]
    pub type_: String,
    pub name: String,
    pub value: String,
    #[serde(default)]
    pub checked: bool,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct RequestPayload {
    pub destination: String,
    pub fields: HashMap<String, String>,
    pub request_vars: HashMap<String, String>,
}

pub fn convert(markup: &str, dark_theme: bool, force_monospace: bool) -> Result<String, String> {
    micron_sys::convert(markup, dark_theme, force_monospace)
}

pub fn parse_header_tags(markup: &str) -> Result<PageColors, String> {
    let raw = micron_sys::parse_header_tags(markup)?;
    serde_json::from_str(&raw).map_err(|e| e.to_string())
}

pub fn collect_form_fields(inputs: &[FieldInput]) -> Result<HashMap<String, String>, String> {
    let raw = micron_sys::collect_form_fields(&serde_json::to_string(inputs).map_err(|e| e.to_string())?)?;
    serde_json::from_str(&raw).map_err(|e| e.to_string())
}

pub fn build_request_payload(
    fields: &HashMap<String, String>,
    destination: &str,
    fields_spec: &str,
) -> Result<RequestPayload, String> {
    let mut map = Map::new();
    for (k, v) in fields {
        map.insert(k.clone(), Value::String(v.clone()));
    }
    let fields_json = Value::Object(map).to_string();
    let raw = micron_sys::build_request_payload(&fields_json, destination, fields_spec)?;
    serde_json::from_str(&raw).map_err(|e| e.to_string())
}
