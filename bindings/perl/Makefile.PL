# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

use ExtUtils::MakeMaker;

WriteMakefile(
    NAME             => 'Micron',
    VERSION_FROM     => 'lib/Micron.pm',
    ABSTRACT         => 'Micron markup parser and HTML renderer (libmicron bindings)',
    AUTHOR           => 'Quad4',
    LICENSE          => 'open_source',
    MIN_PERL_VERSION => '5.020',
    PREREQ_PM        => {
        'FFI::Platypus' => '2.00',
        'JSON::PP'      => '0',
    },
    META_MERGE => {
        'meta-spec' => { version => 2 },
        resources   => {
            repository => {
                type => 'git',
                url  => 'https://github.com/Quad4-Software/micron-parser-go.git',
            },
        },
    },
);
