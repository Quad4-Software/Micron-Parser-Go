# frozen_string_literal: true
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

require "ffi"
require "json"

module Micron
  extend FFI::Library

  def self.lib_name
    case FFI::Platform::OS
    when "darwin" then "libmicron.dylib"
    when "windows" then "libmicron.dll"
    else "libmicron.so"
    end
  end

  def self.resolve_lib
    env = ENV["MICRON_LIB_PATH"]
    return env if env && File.file?(env)

    name = lib_name
    candidates = [
      File.expand_path("../native/#{name}", __dir__),
      File.expand_path("native/#{name}", __dir__),
      File.expand_path("../../../dist/#{name}", __dir__),
      File.expand_path("dist/#{name}"),
    ]
    candidates.find { |p| File.file?(p) } || name
  end

  ffi_lib resolve_lib

  attach_function :micron_convert, [:string, :int, :int], :pointer
  attach_function :micron_parse_header_tags, [:string], :pointer
  attach_function :micron_collect_form_fields, [:string], :pointer
  attach_function :micron_build_request_payload, [:string, :string, :string], :pointer
  attach_function :micron_free, [:pointer], :void

  module_function

  def take_string(ptr)
    return "" if ptr.null?

    begin
      ptr.read_string
    ensure
      micron_free(ptr)
    end
  end

  def convert(markup, dark_theme: true, force_monospace: true)
    take_string(micron_convert(markup.to_s, dark_theme ? 1 : 0, force_monospace ? 1 : 0))
  end

  def parse_header_tags(markup)
    JSON.parse(take_string(micron_parse_header_tags(markup.to_s)))
  end

  def collect_form_fields(inputs)
    JSON.parse(take_string(micron_collect_form_fields(JSON.generate(inputs))))
  end

  def build_request_payload(fields, destination, fields_spec)
    JSON.parse(
      take_string(
        micron_build_request_payload(JSON.generate(fields), destination.to_s, fields_spec.to_s)
      )
    )
  end
end
