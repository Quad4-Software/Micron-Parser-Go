// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

using System.Runtime.InteropServices;
using System.Text.Json;

namespace Quad4.Micron;

public static class Micron
{
    private static readonly IntPtr LibHandle = LoadLibrary();

    private static string LibFileName()
    {
        if (OperatingSystem.IsWindows()) return "libmicron.dll";
        if (OperatingSystem.IsMacOS()) return "libmicron.dylib";
        return "libmicron.so";
    }

    private static IntPtr LoadLibrary()
    {
        var env = Environment.GetEnvironmentVariable("MICRON_LIB_PATH");
        if (!string.IsNullOrWhiteSpace(env) && File.Exists(env))
        {
            return NativeLibrary.Load(env);
        }

        var name = LibFileName();
        var rid = RuntimeInformation.RuntimeIdentifier;
        var candidates = new[]
        {
            Path.Combine(AppContext.BaseDirectory, "runtimes", rid, "native", name),
            Path.Combine(AppContext.BaseDirectory, "native", name),
            Path.Combine("dist", name),
            Path.GetFullPath(Path.Combine("dist", name)),
        };
        foreach (var path in candidates)
        {
            if (File.Exists(path))
            {
                return NativeLibrary.Load(path);
            }
        }
        return NativeLibrary.Load(name);
    }

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr ConvertDelegate(IntPtr markup, int dark, int mono);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr ParseHeaderDelegate(IntPtr markup);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr CollectFieldsDelegate(IntPtr inputsJson);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate IntPtr BuildPayloadDelegate(IntPtr fieldsJson, IntPtr destination, IntPtr fieldsSpec);

    [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
    private delegate void FreeDelegate(IntPtr ptr);

    private static readonly ConvertDelegate ConvertFn =
        Marshal.GetDelegateForFunctionPointer<ConvertDelegate>(NativeLibrary.GetExport(LibHandle, "micron_convert"));
    private static readonly ParseHeaderDelegate ParseHeaderFn =
        Marshal.GetDelegateForFunctionPointer<ParseHeaderDelegate>(NativeLibrary.GetExport(LibHandle, "micron_parse_header_tags"));
    private static readonly CollectFieldsDelegate CollectFieldsFn =
        Marshal.GetDelegateForFunctionPointer<CollectFieldsDelegate>(NativeLibrary.GetExport(LibHandle, "micron_collect_form_fields"));
    private static readonly BuildPayloadDelegate BuildPayloadFn =
        Marshal.GetDelegateForFunctionPointer<BuildPayloadDelegate>(NativeLibrary.GetExport(LibHandle, "micron_build_request_payload"));
    private static readonly FreeDelegate FreeFn =
        Marshal.GetDelegateForFunctionPointer<FreeDelegate>(NativeLibrary.GetExport(LibHandle, "micron_free"));

    private static IntPtr AllocUtf8(string s)
    {
        return Marshal.StringToCoTaskMemUTF8(s ?? "");
    }

    private static string TakeString(IntPtr ptr)
    {
        if (ptr == IntPtr.Zero) return "";
        try
        {
            return Marshal.PtrToStringUTF8(ptr) ?? "";
        }
        finally
        {
            FreeFn(ptr);
        }
    }

    public static string Convert(string markup, bool darkTheme = true, bool forceMonospace = true)
    {
        var p = AllocUtf8(markup);
        try
        {
            return TakeString(ConvertFn(p, darkTheme ? 1 : 0, forceMonospace ? 1 : 0));
        }
        finally
        {
            Marshal.FreeCoTaskMem(p);
        }
    }

    public static JsonElement ParseHeaderTags(string markup)
    {
        var p = AllocUtf8(markup);
        try
        {
            return JsonDocument.Parse(TakeString(ParseHeaderFn(p))).RootElement.Clone();
        }
        finally
        {
            Marshal.FreeCoTaskMem(p);
        }
    }

    public static JsonElement CollectFormFields(string inputsJson)
    {
        var p = AllocUtf8(inputsJson);
        try
        {
            return JsonDocument.Parse(TakeString(CollectFieldsFn(p))).RootElement.Clone();
        }
        finally
        {
            Marshal.FreeCoTaskMem(p);
        }
    }

    public static JsonElement BuildRequestPayload(string fieldsJson, string destination, string fieldsSpec)
    {
        var f = AllocUtf8(fieldsJson);
        var d = AllocUtf8(destination);
        var s = AllocUtf8(fieldsSpec);
        try
        {
            return JsonDocument.Parse(TakeString(BuildPayloadFn(f, d, s))).RootElement.Clone();
        }
        finally
        {
            Marshal.FreeCoTaskMem(f);
            Marshal.FreeCoTaskMem(d);
            Marshal.FreeCoTaskMem(s);
        }
    }
}
