// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

const std = @import("std");
const micron = @import("micron");

fn contains(haystack: []const u8, needle: []const u8) bool {
    return std.mem.indexOf(u8, haystack, needle) != null;
}

pub fn main() !void {
    var dbg = std.heap.DebugAllocator(.{}){};
    defer _ = dbg.deinit();
    const allocator = dbg.allocator();

    const html = try micron.convert(allocator, "> Title\n\nHello <world> & `*bold`*.\n", true, false);
    defer allocator.free(html);
    if (!contains(html, "Hello") or !contains(html, "&lt;world&gt;") or !contains(html, "bold")) {
        std.debug.print("convert failed\n", .{});
        std.process.exit(1);
    }

    const colors = try micron.parseHeaderTags(allocator, "#!fg=ccc\n#!bg=222\n\nBody\n");
    defer allocator.free(colors);
    if (!contains(colors, "\"fg\":\"ccc\"") or !contains(colors, "\"bg\":\"222\"")) {
        std.debug.print("headers failed: {s}\n", .{colors});
        std.process.exit(1);
    }

    const fields = try micron.collectFormFields(
        allocator,
        "[{\"type\":\"text\",\"name\":\"user\",\"value\":\"alice\",\"checked\":false},{\"type\":\"checkbox\",\"name\":\"opts\",\"value\":\"1\",\"checked\":true}]",
    );
    defer allocator.free(fields);
    if (!contains(fields, "alice")) {
        std.debug.print("fields failed: {s}\n", .{fields});
        std.process.exit(1);
    }

    const payload = try micron.buildRequestPayload(allocator, "{\"user\":\"alice\",\"opts\":\"1\"}", "/page`x=1", "user|opts");
    defer allocator.free(payload);
    if (!contains(payload, "\"destination\":\"/page\"")) {
        std.debug.print("payload failed: {s}\n", .{payload});
        std.process.exit(1);
    }

    std.debug.print("zig smoke ok\n", .{});
}
