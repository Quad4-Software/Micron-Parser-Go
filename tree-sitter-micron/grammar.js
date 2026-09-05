/**
 * Tree-sitter grammar for NomadNet Micron (.mu).
 * Aligns with micron-parser-go Document IR / NomadNet Python dialect.
 */

module.exports = grammar({
  name: "micron",

  extras: ($) => [/\r?\n/, /[ \t]/],

  rules: {
    source_file: ($) => repeat($._line),

    _line: ($) =>
      choice(
        $.header_directive,
        $.comment,
        $.section,
        $.section_reset,
        $.literal_toggle,
        $.table_fence,
        $.partial,
        $.hr,
        $.divider,
        $.content_line,
        $.blank_line
      ),

    blank_line: ($) => /\r?\n/,

    header_directive: ($) =>
      seq("#!", choice("fg", "bg"), "=", $.hex_color),

    hex_color: ($) => /[0-9a-fA-F]{3}([0-9a-fA-F]{3})?/,

    comment: ($) => seq("#", /[^\n]*/),

    section: ($) => seq($.section_markers, optional($.content_line)),

    section_markers: ($) => />+/,

    section_reset: ($) => seq("<", optional($.content_line)),

    literal_toggle: ($) => prec(1, seq("`", "=")),

    table_fence: ($) => seq("`", "t", optional(/[lcr]?[0-9]*/)),

    partial: ($) => seq("`", "{", /[^}]*/, "}"),

    hr: ($) => prec(1, "-"),

    divider: ($) => seq("-", /[^\n]+/),

    content_line: ($) => repeat1($._inline),

    _inline: ($) =>
      choice(
        $.format_span,
        $.link,
        $.field,
        $.text
      ),

    format_span: ($) =>
      seq(
        "`",
        choice("!", "*", "_", "c", "l", "r", "a", "f", "b", $.color_tag),
        optional($.text)
      ),

    color_tag: ($) =>
      choice(
        seq("F", $.hex_color),
        seq("B", $.hex_color),
        seq("FT", /[0-9a-fA-F]{6}/),
        seq("BT", /[0-9a-fA-F]{6}/)
      ),

    link: ($) =>
      seq("`", "[", /[^\]`]*/, optional(seq("`", /[^\]`]*/)), optional(seq("`", /[^\]`]*/)), "]"),

    field: ($) => seq("`", "<", /[^>`]*/, "`", /[^>]*/, ">"),

    text: ($) => /[^`\n]+/,
  },
});
