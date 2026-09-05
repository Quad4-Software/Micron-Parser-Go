// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const mod = b.addModule("micron", .{
        .root_source_file = b.path("src/root.zig"),
        .target = target,
        .optimize = optimize,
        .link_libc = true,
    });

    const smoke = b.addExecutable(.{
        .name = "micron-smoke",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/smoke.zig"),
            .target = target,
            .optimize = optimize,
            .imports = &.{
                .{ .name = "micron", .module = mod },
            },
            .link_libc = true,
        }),
    });
    b.installArtifact(smoke);

    const run_smoke = b.addRunArtifact(smoke);
    const smoke_step = b.step("smoke", "Run micron Zig smoke test");
    smoke_step.dependOn(&run_smoke.step);
}
