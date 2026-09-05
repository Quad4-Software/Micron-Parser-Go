# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD
package Micron;

use strict;
use warnings;
use Carp qw(croak);
use Exporter qw(import);
use File::Basename qw(dirname);
use File::Spec;
use JSON::PP qw(decode_json encode_json);

our @EXPORT_OK = qw(convert parse_header_tags collect_form_fields build_request_payload);
our $VERSION = '1.1.0';

my $ffi;
my $ready;

sub _lib_name {
    return 'libmicron.dylib' if $^O eq 'darwin';
    return 'libmicron.dll'   if $^O eq 'MSWin32';
    return 'libmicron.so';
}

sub _resolve_lib {
    my $env = $ENV{MICRON_LIB_PATH};
    return $env if defined $env && length $env && -f $env;

    my $name = _lib_name();
    my $here = dirname(dirname(File::Spec->rel2abs(__FILE__)));
    my @candidates = (
        File::Spec->catfile($here, 'native', $name),
        File::Spec->catfile($here, '..', '..', 'dist', $name),
        File::Spec->catfile(File::Spec->curdir(), 'dist', $name),
    );
    for my $path (@candidates) {
        return $path if -f $path;
    }
    croak 'libmicron not found; set MICRON_LIB_PATH or place the shared library under bindings/perl/native or dist/';
}

sub _api {
    return if $ready;
    require FFI::Platypus;
    $ffi = FFI::Platypus->new(api => 2);
    $ffi->lib(_resolve_lib());
    $ffi->attach(micron_convert               => ['string', 'int', 'int'] => 'opaque');
    $ffi->attach(micron_parse_header_tags     => ['string'] => 'opaque');
    $ffi->attach(micron_collect_form_fields   => ['string'] => 'opaque');
    $ffi->attach(micron_build_request_payload => ['string', 'string', 'string'] => 'opaque');
    $ffi->attach(micron_free                  => ['opaque'] => 'void');
    $ready = 1;
}

sub _take_string {
    my ($ptr) = @_;
    return '' unless defined $ptr && $ptr;
    my $s = $ffi->cast('opaque' => 'string', $ptr);
    micron_free($ptr);
    return defined $s ? $s : '';
}

sub convert {
    my ($markup, %opts) = @_;
    _api();
    my $dark = ($opts{dark_theme} // 1) ? 1 : 0;
    my $mono = ($opts{force_monospace} // 1) ? 1 : 0;
    return _take_string(micron_convert($markup // '', $dark, $mono));
}

sub parse_header_tags {
    my ($markup) = @_;
    _api();
    my $data = decode_json(_take_string(micron_parse_header_tags($markup // '')) || '{"fg":"","bg":""}');
    return {
        fg => $data->{fg} // '',
        bg => $data->{bg} // '',
    };
}

sub collect_form_fields {
    my ($inputs) = @_;
    _api();
    my $raw = _take_string(micron_collect_form_fields(encode_json($inputs // [])));
    return decode_json($raw || '{}');
}

sub build_request_payload {
    my ($fields, $destination, $fields_spec) = @_;
    _api();
    my $raw = _take_string(
        micron_build_request_payload(
            encode_json($fields // {}),
            $destination // '',
            $fields_spec // '',
        )
    );
    return decode_json($raw || '{"destination":"","fields":{},"request_vars":{}}');
}

1;
