#!/usr/bin/env perl
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD
use strict;
use warnings;
use FindBin;
use lib "$FindBin::Bin/lib";
use JSON::PP ();
use Micron qw(convert parse_header_tags collect_form_fields build_request_payload);

my $html = convert("> Title\n\nHello <world> & `*bold`*.\n", dark_theme => 1, force_monospace => 0);
die "convert failed\n" unless index($html, 'Hello') >= 0 && index($html, '&lt;world&gt;') >= 0 && index($html, 'bold') >= 0;

my $colors = parse_header_tags("#!fg=ccc\n#!bg=222\n\nBody\n");
die "headers failed\n" unless ($colors->{fg} // '') eq 'ccc' && ($colors->{bg} // '') eq '222';

my $fields = collect_form_fields([
    { type => 'text', name => 'user', value => 'alice', checked => $JSON::PP::false },
    { type => 'checkbox', name => 'opts', value => '1', checked => $JSON::PP::true },
]);
die "fields failed\n" unless ($fields->{user} // '') eq 'alice' && ($fields->{opts} // '') eq '1';

my $payload = build_request_payload($fields, '/page`x=1', 'user|opts');
die "payload failed\n" unless ($payload->{destination} // '') eq '/page';

print "perl smoke ok\n";
