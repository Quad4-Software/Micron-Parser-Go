// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

//! Safe Micron bindings over libmicron via dynamic loading.

const std = @import("std");
const builtin = @import("builtin");

const ConvertFn = *const fn ([*:0]const u8, c_int, c_int) callconv(.c) ?[*:0]u8;
const ParseHeaderFn = *const fn ([*:0]const u8) callconv(.c) ?[*:0]u8;
const CollectFieldsFn = *const fn ([*:0]const u8) callconv(.c) ?[*:0]u8;
const BuildPayloadFn = *const fn ([*:0]const u8, [*:0]const u8, [*:0]const u8) callconv(.c) ?[*:0]u8;
const FreeFn = *const fn (?[*:0]u8) callconv(.c) void;

const Api = struct {
    lib: std.DynLib,
    convert_fn: ConvertFn,
    parse_header_tags_fn: ParseHeaderFn,
    collect_form_fields_fn: CollectFieldsFn,
    build_request_payload_fn: BuildPayloadFn,
    free_fn: FreeFn,
};

var api_cell: ?Api = null;

fn libName() []const u8 {
    return switch (builtin.os.tag) {
        .macos => "libmicron.dylib",
        .windows => "libmicron.dll",
        else => "libmicron.so",
    };
}

fn candidatePaths(allocator: std.mem.Allocator) ![][]const u8 {
    var list: std.ArrayListUnmanaged([]const u8) = .empty;
    errdefer {
        for (list.items) |p| allocator.free(p);
        list.deinit(allocator);
    }

    if (std.c.getenv("MICRON_LIB_PATH")) |env| {
        try list.append(allocator, try allocator.dupe(u8, std.mem.span(env)));
    }

    const name = libName();
    try list.append(allocator, try std.fs.path.join(allocator, &.{ "native", name }));
    try list.append(allocator, try std.fs.path.join(allocator, &.{ "..", "..", "dist", name }));
    try list.append(allocator, try std.fs.path.join(allocator, &.{ "dist", name }));

    return try list.toOwnedSlice(allocator);
}

fn loadApi(allocator: std.mem.Allocator) !Api {
    if (api_cell) |api| return api;

    const paths = try candidatePaths(allocator);
    defer {
        for (paths) |p| allocator.free(p);
        allocator.free(paths);
    }

    var last_err: anyerror = error.LibNotFound;
    for (paths) |path| {
        var lib = std.DynLib.open(path) catch |err| {
            last_err = err;
            continue;
        };
        errdefer lib.close();

        const convert_fn = lib.lookup(ConvertFn, "micron_convert") orelse continue;
        const parse_header_tags_fn = lib.lookup(ParseHeaderFn, "micron_parse_header_tags") orelse continue;
        const collect_form_fields_fn = lib.lookup(CollectFieldsFn, "micron_collect_form_fields") orelse continue;
        const build_request_payload_fn = lib.lookup(BuildPayloadFn, "micron_build_request_payload") orelse continue;
        const free_fn = lib.lookup(FreeFn, "micron_free") orelse continue;

        const api = Api{
            .lib = lib,
            .convert_fn = convert_fn,
            .parse_header_tags_fn = parse_header_tags_fn,
            .collect_form_fields_fn = collect_form_fields_fn,
            .build_request_payload_fn = build_request_payload_fn,
            .free_fn = free_fn,
        };
        api_cell = api;
        return api;
    }
    return last_err;
}

fn takeString(api: Api, ptr: ?[*:0]u8, allocator: std.mem.Allocator) ![]u8 {
    if (ptr == null) return try allocator.dupe(u8, "");
    defer api.free_fn(ptr);
    return try allocator.dupe(u8, std.mem.span(ptr.?));
}

pub fn convert(allocator: std.mem.Allocator, markup: []const u8, dark_theme: bool, force_monospace: bool) ![]u8 {
    const api = try loadApi(allocator);
    const c_markup = try allocator.dupeZ(u8, markup);
    defer allocator.free(c_markup);
    const ptr = api.convert_fn(c_markup.ptr, @intFromBool(dark_theme), @intFromBool(force_monospace));
    return takeString(api, ptr, allocator);
}

pub fn parseHeaderTags(allocator: std.mem.Allocator, markup: []const u8) ![]u8 {
    const api = try loadApi(allocator);
    const c_markup = try allocator.dupeZ(u8, markup);
    defer allocator.free(c_markup);
    return takeString(api, api.parse_header_tags_fn(c_markup.ptr), allocator);
}

pub fn collectFormFields(allocator: std.mem.Allocator, inputs_json: []const u8) ![]u8 {
    const api = try loadApi(allocator);
    const c_json = try allocator.dupeZ(u8, inputs_json);
    defer allocator.free(c_json);
    return takeString(api, api.collect_form_fields_fn(c_json.ptr), allocator);
}

pub fn buildRequestPayload(
    allocator: std.mem.Allocator,
    fields_json: []const u8,
    destination: []const u8,
    fields_spec: []const u8,
) ![]u8 {
    const api = try loadApi(allocator);
    const c_fields = try allocator.dupeZ(u8, fields_json);
    defer allocator.free(c_fields);
    const c_dest = try allocator.dupeZ(u8, destination);
    defer allocator.free(c_dest);
    const c_spec = try allocator.dupeZ(u8, fields_spec);
    defer allocator.free(c_spec);
    return takeString(api, api.build_request_payload_fn(c_fields.ptr, c_dest.ptr, c_spec.ptr), allocator);
}
