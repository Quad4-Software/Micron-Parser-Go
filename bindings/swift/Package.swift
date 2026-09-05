// swift-tools-version: 5.9
// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

import PackageDescription

let package = Package(
    name: "Micron",
    products: [
        .library(name: "Micron", targets: ["Micron"]),
        .executable(name: "micron-smoke", targets: ["MicronSmoke"]),
    ],
    targets: [
        .target(name: "Micron"),
        .executableTarget(
            name: "MicronSmoke",
            dependencies: ["Micron"],
            path: "Sources/MicronSmoke"
        ),
    ]
)
