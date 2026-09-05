// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

using Quad4.Micron;

var html = Micron.Convert("> Title\n\nHello <world> & `*bold`*.\n", true, false);
if (!html.Contains("Hello") || !html.Contains("bold"))
{
    throw new Exception("convert failed");
}

var colors = Micron.ParseHeaderTags("#!fg=ccc\n#!bg=222\n\nBody\n");
if (colors.GetProperty("fg").GetString() != "ccc" || colors.GetProperty("bg").GetString() != "222")
{
    throw new Exception("headers failed: " + colors);
}

var fields = Micron.CollectFormFields(
    "[{\"type\":\"text\",\"name\":\"user\",\"value\":\"alice\",\"checked\":false}]"
);
if (fields.GetProperty("user").GetString() != "alice")
{
    throw new Exception("fields failed: " + fields);
}

var payload = Micron.BuildRequestPayload("{\"user\":\"alice\"}", "/page`x=1", "user");
if (payload.GetProperty("destination").GetString() != "/page")
{
    throw new Exception("payload failed: " + payload);
}

Console.WriteLine("csharp smoke ok");
