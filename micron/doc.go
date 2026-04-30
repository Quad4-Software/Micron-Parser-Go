// Copyright Quad4 2026
// SPDX-License-Identifier: 0BSD

// Package micron parses Micron markup and renders HTML fragments intended for
// embedding in a host page.
//
// # HTML and security
//
// ConvertMicronToHTML returns a fragment: user text is escaped; attribute
// values on generated elements use HTML escaping. The host application still
// controls how that fragment is mounted (for example innerHTML vs trusted
// DOM APIs), Content-Security-Policy, and how link destinations and partial
// fetch URLs are interpreted. Link href and data-* values follow Micron /
// NomadNet URL rules (see FormatNomadnetworkURL); they are not an arbitrary
// URL allowlist.
//
// # Concurrency
//
// Parser holds only DarkTheme and ForceMonospace. It has no per-conversion
// mutable state; the same Parser value may be used from multiple goroutines.
//
// # Reference implementation
//
// Behavioral parity with the JavaScript reference is validated by tests that
// compare structural signatures of the HTML output, not byte-identical
// strings. The reference script used in tests lives under micron/testdata/.
package micron
