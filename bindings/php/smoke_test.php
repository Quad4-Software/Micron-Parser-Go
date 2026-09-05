<?php
# Copyright Quad4 2026
# SPDX-License-Identifier: 0BSD

declare(strict_types=1);

require_once __DIR__ . '/Micron.php';

$html = Micron::convert("> Title\n\nHello <world> & `*bold`*.\n", true, false);
if (!str_contains($html, 'Hello') || !str_contains($html, '&lt;world&gt;') || !str_contains($html, 'bold')) {
    fwrite(STDERR, "convert failed\n");
    exit(1);
}

$colors = Micron::parseHeaderTags("#!fg=ccc\n#!bg=222\n\nBody\n");
if (($colors['fg'] ?? '') !== 'ccc' || ($colors['bg'] ?? '') !== '222') {
    fwrite(STDERR, "headers failed\n");
    exit(1);
}

$fields = Micron::collectFormFields([
    ['type' => 'text', 'name' => 'user', 'value' => 'alice', 'checked' => false],
    ['type' => 'checkbox', 'name' => 'opts', 'value' => '1', 'checked' => true],
]);
if (($fields['user'] ?? '') !== 'alice' || ($fields['opts'] ?? '') !== '1') {
    fwrite(STDERR, "fields failed\n");
    exit(1);
}

$payload = Micron::buildRequestPayload($fields, '/page`x=1', 'user|opts');
if (($payload['destination'] ?? '') !== '/page') {
    fwrite(STDERR, "payload failed\n");
    exit(1);
}

echo "php smoke ok\n";
