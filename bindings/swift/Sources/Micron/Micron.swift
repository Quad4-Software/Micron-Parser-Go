// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

import Foundation
#if canImport(Glibc)
import Glibc
#elseif canImport(Darwin)
import Darwin
#endif

/// Micron markup parser bindings over libmicron.
public enum Micron {
    private static let handle: UnsafeMutableRawPointer = {
        if let env = ProcessInfo.processInfo.environment["MICRON_LIB_PATH"],
           FileManager.default.fileExists(atPath: env),
           let h = dlopen(env, RTLD_NOW | RTLD_LOCAL) {
            return h
        }
        let name: String
        #if os(macOS)
        name = "libmicron.dylib"
        #elseif os(Windows)
        name = "libmicron.dll"
        #else
        name = "libmicron.so"
        #endif
        let cwd = FileManager.default.currentDirectoryPath
        let candidates = [
            "native/\(name)",
            "../../dist/\(name)",
            "dist/\(name)",
            "\(cwd)/dist/\(name)",
            "\(cwd)/../../dist/\(name)",
        ]
        for path in candidates {
            if FileManager.default.fileExists(atPath: path),
               let h = dlopen(path, RTLD_NOW | RTLD_LOCAL) {
                return h
            }
        }
        fatalError(
            "libmicron not found; set MICRON_LIB_PATH or place the shared library under bindings/swift/native or dist/"
        )
    }()

    private typealias ConvertFn = @convention(c) (UnsafePointer<CChar>?, Int32, Int32) -> UnsafeMutablePointer<CChar>?
    private typealias ParseHeaderFn = @convention(c) (UnsafePointer<CChar>?) -> UnsafeMutablePointer<CChar>?
    private typealias CollectFieldsFn = @convention(c) (UnsafePointer<CChar>?) -> UnsafeMutablePointer<CChar>?
    private typealias BuildPayloadFn = @convention(c) (
        UnsafePointer<CChar>?, UnsafePointer<CChar>?, UnsafePointer<CChar>?
    ) -> UnsafeMutablePointer<CChar>?
    private typealias FreeFn = @convention(c) (UnsafeMutablePointer<CChar>?) -> Void

    private static let convertFn: ConvertFn = loadSymbol("micron_convert")
    private static let parseHeaderFn: ParseHeaderFn = loadSymbol("micron_parse_header_tags")
    private static let collectFieldsFn: CollectFieldsFn = loadSymbol("micron_collect_form_fields")
    private static let buildPayloadFn: BuildPayloadFn = loadSymbol("micron_build_request_payload")
    private static let freeFn: FreeFn = loadSymbol("micron_free")

    private static func loadSymbol<T>(_ name: String) -> T {
        guard let sym = dlsym(handle, name) else {
            fatalError("missing symbol \(name)")
        }
        return unsafeBitCast(sym, to: T.self)
    }

    private static func takeString(_ ptr: UnsafeMutablePointer<CChar>?) -> String {
        guard let ptr else { return "" }
        defer { freeFn(ptr) }
        return String(cString: ptr)
    }

    public static func convert(
        _ markup: String,
        darkTheme: Bool = true,
        forceMonospace: Bool = true
    ) -> String {
        markup.withCString { cMarkup in
            takeString(convertFn(cMarkup, darkTheme ? 1 : 0, forceMonospace ? 1 : 0))
        }
    }

    public static func parseHeaderTags(_ markup: String) -> [String: String] {
        let raw = markup.withCString { takeString(parseHeaderFn($0)) }
        guard let data = raw.data(using: .utf8),
              let obj = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else {
            return ["fg": "", "bg": ""]
        }
        return [
            "fg": obj["fg"] as? String ?? "",
            "bg": obj["bg"] as? String ?? "",
        ]
    }

    public static func collectFormFields(_ inputs: [[String: Any]]) -> [String: String] {
        guard let data = try? JSONSerialization.data(withJSONObject: inputs),
              let json = String(data: data, encoding: .utf8) else {
            return [:]
        }
        let raw = json.withCString { takeString(collectFieldsFn($0)) }
        guard let outData = raw.data(using: .utf8),
              let obj = try? JSONSerialization.jsonObject(with: outData) as? [String: Any] else {
            return [:]
        }
        var result: [String: String] = [:]
        for (k, v) in obj {
            result[k] = "\(v)"
        }
        return result
    }

    public static func buildRequestPayload(
        _ fields: [String: String],
        destination: String,
        fieldsSpec: String
    ) -> [String: Any] {
        guard let data = try? JSONSerialization.data(withJSONObject: fields),
              let json = String(data: data, encoding: .utf8) else {
            return [:]
        }
        let raw = json.withCString { cFields in
            destination.withCString { cDest in
                fieldsSpec.withCString { cSpec in
                    takeString(buildPayloadFn(cFields, cDest, cSpec))
                }
            }
        }
        guard let outData = raw.data(using: .utf8),
              let obj = try? JSONSerialization.jsonObject(with: outData) as? [String: Any] else {
            return [:]
        }
        return obj
    }
}
