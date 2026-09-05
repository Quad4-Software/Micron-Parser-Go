// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

//! Dynamic loader for libmicron.

use libloading::{Library, Symbol};
use std::ffi::{CStr, CString, c_char, c_int};
use std::path::{Path, PathBuf};
use std::sync::OnceLock;

type ConvertFn = unsafe extern "C" fn(*const c_char, c_int, c_int) -> *mut c_char;
type ParseHeaderFn = unsafe extern "C" fn(*const c_char) -> *mut c_char;
type CollectFieldsFn = unsafe extern "C" fn(*const c_char) -> *mut c_char;
type BuildPayloadFn =
    unsafe extern "C" fn(*const c_char, *const c_char, *const c_char) -> *mut c_char;
type FreeFn = unsafe extern "C" fn(*mut c_char);

struct Api {
    _lib: Library,
    convert: ConvertFn,
    parse_header_tags: ParseHeaderFn,
    collect_form_fields: CollectFieldsFn,
    build_request_payload: BuildPayloadFn,
    free: FreeFn,
}

fn default_lib_name() -> &'static str {
    if cfg!(target_os = "windows") {
        "libmicron.dll"
    } else if cfg!(target_os = "macos") {
        "libmicron.dylib"
    } else {
        "libmicron.so"
    }
}

fn candidate_paths() -> Vec<PathBuf> {
    let mut out = Vec::new();
    if let Ok(p) = std::env::var("MICRON_LIB_PATH") {
        out.push(PathBuf::from(p));
    }
    let name = default_lib_name();
    if let Ok(manifest) = std::env::var("CARGO_MANIFEST_DIR") {
        let root = Path::new(&manifest);
        out.push(root.join("native").join(name));
        out.push(root.join("..").join("..").join("..").join("dist").join(name));
    }
    out.push(PathBuf::from("dist").join(name));
    out.push(PathBuf::from(name));
    out
}

fn try_load() -> Result<Api, String> {
    let mut last = String::from("libmicron not found");
    for path in candidate_paths() {
        if !path.is_file() {
            continue;
        }
        let lib = unsafe { Library::new(&path) }.map_err(|e| {
            last = format!("{}: {}", path.display(), e);
            last.clone()
        });
        let lib = match lib {
            Ok(l) => l,
            Err(_) => continue,
        };
        unsafe {
            let convert: Symbol<ConvertFn> = lib
                .get(b"micron_convert\0")
                .map_err(|e| e.to_string())?;
            let parse_header_tags: Symbol<ParseHeaderFn> = lib
                .get(b"micron_parse_header_tags\0")
                .map_err(|e| e.to_string())?;
            let collect_form_fields: Symbol<CollectFieldsFn> = lib
                .get(b"micron_collect_form_fields\0")
                .map_err(|e| e.to_string())?;
            let build_request_payload: Symbol<BuildPayloadFn> = lib
                .get(b"micron_build_request_payload\0")
                .map_err(|e| e.to_string())?;
            let free: Symbol<FreeFn> = lib.get(b"micron_free\0").map_err(|e| e.to_string())?;
            return Ok(Api {
                convert: *convert,
                parse_header_tags: *parse_header_tags,
                collect_form_fields: *collect_form_fields,
                build_request_payload: *build_request_payload,
                free: *free,
                _lib: lib,
            });
        }
    }
    Err(format!(
        "libmicron not found; set MICRON_LIB_PATH. Last error: {last}"
    ))
}

fn load_api() -> Result<&'static Api, String> {
    static API: OnceLock<Result<Api, String>> = OnceLock::new();
    match API.get_or_init(try_load) {
        Ok(api) => Ok(api),
        Err(e) => Err(e.clone()),
    }
}

unsafe fn take_string(api: &Api, ptr: *mut c_char) -> String {
    if ptr.is_null() {
        return String::new();
    }
    let s = unsafe { CStr::from_ptr(ptr) }
        .to_string_lossy()
        .into_owned();
    unsafe { (api.free)(ptr) };
    s
}

pub fn convert(markup: &str, dark_theme: bool, force_monospace: bool) -> Result<String, String> {
    let api = load_api()?;
    let c_markup = CString::new(markup).map_err(|e| e.to_string())?;
    let ptr = unsafe {
        (api.convert)(
            c_markup.as_ptr(),
            dark_theme as c_int,
            force_monospace as c_int,
        )
    };
    Ok(unsafe { take_string(api, ptr) })
}

pub fn parse_header_tags(markup: &str) -> Result<String, String> {
    let api = load_api()?;
    let c_markup = CString::new(markup).map_err(|e| e.to_string())?;
    let ptr = unsafe { (api.parse_header_tags)(c_markup.as_ptr()) };
    Ok(unsafe { take_string(api, ptr) })
}

pub fn collect_form_fields(inputs_json: &str) -> Result<String, String> {
    let api = load_api()?;
    let c_json = CString::new(inputs_json).map_err(|e| e.to_string())?;
    let ptr = unsafe { (api.collect_form_fields)(c_json.as_ptr()) };
    Ok(unsafe { take_string(api, ptr) })
}

pub fn build_request_payload(
    fields_json: &str,
    destination: &str,
    fields_spec: &str,
) -> Result<String, String> {
    let api = load_api()?;
    let c_fields = CString::new(fields_json).map_err(|e| e.to_string())?;
    let c_dest = CString::new(destination).map_err(|e| e.to_string())?;
    let c_spec = CString::new(fields_spec).map_err(|e| e.to_string())?;
    let ptr = unsafe {
        (api.build_request_payload)(c_fields.as_ptr(), c_dest.as_ptr(), c_spec.as_ptr())
    };
    Ok(unsafe { take_string(api, ptr) })
}
