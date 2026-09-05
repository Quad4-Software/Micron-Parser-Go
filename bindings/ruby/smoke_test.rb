# frozen_string_literal: true
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

require_relative "lib/micron"

html = Micron.convert("> Title\n\nHello <world> & `*bold`*.\n", dark_theme: true, force_monospace: false)
raise "convert failed" unless html.include?("Hello") && html.include?("bold")

colors = Micron.parse_header_tags("#!fg=ccc\n#!bg=222\n\nBody\n")
raise "headers failed" unless colors["fg"] == "ccc" && colors["bg"] == "222"

fields = Micron.collect_form_fields([
  { "type" => "text", "name" => "user", "value" => "alice", "checked" => false },
  { "type" => "checkbox", "name" => "opts", "value" => "1", "checked" => true },
])
raise "fields failed" unless fields["user"] == "alice" && fields["opts"] == "1"

payload = Micron.build_request_payload(fields, "/page`x=1", "user|opts")
raise "payload failed" unless payload["destination"] == "/page" && payload["fields"]["user"] == "alice"

puts "ruby smoke ok"
