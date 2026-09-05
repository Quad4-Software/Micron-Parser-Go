# frozen_string_literal: true
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

Gem::Specification.new do |spec|
  spec.name          = "micron-parser"
  spec.version       = "1.1.0"
  spec.authors       = ["Quad4"]
  spec.summary       = "Micron markup parser and HTML renderer (libmicron bindings)"
  spec.license       = "0BSD"
  spec.files         = Dir["lib/**/*", "native/**/*"]
  spec.require_paths = ["lib"]
  spec.required_ruby_version = ">= 3.0.0"
  spec.add_runtime_dependency "ffi", "~> 1.15"
end
