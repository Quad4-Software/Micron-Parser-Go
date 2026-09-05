// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

package io.quad4.micron;

import com.sun.jna.Native;
import com.sun.jna.Pointer;

import java.io.IOException;
import java.io.InputStream;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Locale;

/**
 * Micron markup parser facade over libmicron.
 */
public final class Micron {
    private static final MicronNative LIB = load();

    private Micron() {}

    private static String libFileName() {
        String os = System.getProperty("os.name", "").toLowerCase(Locale.ROOT);
        if (os.contains("mac") || os.contains("darwin")) {
            return "libmicron.dylib";
        }
        if (os.contains("win")) {
            return "libmicron.dll";
        }
        return "libmicron.so";
    }

    private static String osArchDir() {
        String os = System.getProperty("os.name", "").toLowerCase(Locale.ROOT);
        String arch = System.getProperty("os.arch", "").toLowerCase(Locale.ROOT);
        String osKey;
        if (os.contains("mac") || os.contains("darwin")) {
            osKey = "darwin";
        } else if (os.contains("win")) {
            osKey = "windows";
        } else {
            osKey = "linux";
        }
        String archKey;
        if (arch.contains("aarch64") || arch.equals("arm64")) {
            archKey = "arm64";
        } else if (arch.contains("64")) {
            archKey = "amd64";
        } else {
            archKey = arch;
        }
        return osKey + "-" + archKey;
    }

    private static MicronNative load() {
        String env = System.getenv("MICRON_LIB_PATH");
        if (env != null && !env.isBlank()) {
            Path p = Path.of(env);
            System.setProperty("jna.library.path", p.getParent().toAbsolutePath().toString());
            return Native.load(stripExt(p.getFileName().toString()), MicronNative.class);
        }
        Path extracted = extractBundled();
        if (extracted != null) {
            System.setProperty("jna.library.path", extracted.getParent().toString());
            return Native.load(stripExt(extracted.getFileName().toString()), MicronNative.class);
        }
        Path dist = Path.of("dist", libFileName()).toAbsolutePath();
        if (Files.isRegularFile(dist)) {
            System.setProperty("jna.library.path", dist.getParent().toString());
            return Native.load(stripExt(libFileName()), MicronNative.class);
        }
        return Native.load("micron", MicronNative.class);
    }

    private static String stripExt(String name) {
        if (name.startsWith("lib")) {
            name = name.substring(3);
        }
        int dot = name.lastIndexOf('.');
        return dot > 0 ? name.substring(0, dot) : name;
    }

    private static Path extractBundled() {
        String resource = "/natives/" + osArchDir() + "/" + libFileName();
        try (InputStream in = Micron.class.getResourceAsStream(resource)) {
            if (in == null) {
                return null;
            }
            Path dir = Files.createTempDirectory("libmicron");
            dir.toFile().deleteOnExit();
            Path out = dir.resolve(libFileName());
            Files.copy(in, out);
            out.toFile().deleteOnExit();
            return out;
        } catch (IOException e) {
            return null;
        }
    }

    private static String takeString(Pointer ptr) {
        if (ptr == null) {
            return "";
        }
        try {
            return ptr.getString(0, "utf8");
        } finally {
            LIB.micron_free(ptr);
        }
    }

    public static String convert(String markup, boolean darkTheme, boolean forceMonospace) {
        return takeString(
                LIB.micron_convert(markup == null ? "" : markup, darkTheme ? 1 : 0, forceMonospace ? 1 : 0));
    }

    public static String parseHeaderTagsJson(String markup) {
        return takeString(LIB.micron_parse_header_tags(markup == null ? "" : markup));
    }

    public static String collectFormFieldsJson(String inputsJson) {
        return takeString(LIB.micron_collect_form_fields(inputsJson == null ? "[]" : inputsJson));
    }

    public static String buildRequestPayloadJson(String fieldsJson, String destination, String fieldsSpec) {
        return takeString(
                LIB.micron_build_request_payload(
                        fieldsJson == null ? "{}" : fieldsJson,
                        destination == null ? "" : destination,
                        fieldsSpec == null ? "" : fieldsSpec));
    }

    /** Smoke entrypoint: java -cp ... io.quad4.micron.Micron */
    public static void main(String[] args) {
        String html = convert("> Title\n\nHello <world> & `*bold`*.\n", true, false);
        if (!html.contains("Hello") || !html.contains("bold")) {
            throw new IllegalStateException("convert failed");
        }
        String colors = parseHeaderTagsJson("#!fg=ccc\n#!bg=222\n\nBody\n");
        if (!colors.contains("\"fg\":\"ccc\"") || !colors.contains("\"bg\":\"222\"")) {
            throw new IllegalStateException("headers failed: " + colors);
        }
        String fields = collectFormFieldsJson(
                "[{\"type\":\"text\",\"name\":\"user\",\"value\":\"alice\",\"checked\":false}]");
        if (!fields.contains("alice")) {
            throw new IllegalStateException("fields failed: " + fields);
        }
        String payload = buildRequestPayloadJson("{\"user\":\"alice\"}", "/page`x=1", "user");
        if (!payload.contains("\"destination\":\"/page\"")) {
            throw new IllegalStateException("payload failed: " + payload);
        }
        System.out.println("java smoke ok");
    }
}
