"use strict";

function escapeHtml(s) {
  return String(s)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#39;");
}

class FakeTextNode {
  constructor(text) {
    this.textContent = String(text);
  }
  get outerHTML() {
    return escapeHtml(this.textContent);
  }
}

class FakeElement {
  constructor(tag) {
    this.tagName = String(tag).toLowerCase();
    this.style = {};
    this.attrs = {};
    this.children = [];
    this._innerHTML = "";
    this._textContent = "";
    this.classList = {
      add: (name) => {
        const cur = this.attrs.class ? this.attrs.class.split(/\s+/).filter(Boolean) : [];
        if (!cur.includes(name)) cur.push(name);
        this.attrs.class = cur.join(" ");
      },
    };
  }
  appendChild(child) {
    this.children.push(child);
    return child;
  }
  setAttribute(name, value) {
    this.attrs[String(name)] = String(value);
  }
  set innerHTML(v) {
    this._innerHTML = String(v);
  }
  get innerHTML() {
    if (this._innerHTML) return this._innerHTML;
    if (this.children.length) return this.children.map((c) => c.outerHTML || "").join("");
    return "";
  }
  set textContent(v) {
    this._textContent = String(v);
    this._innerHTML = "";
    this.children = [];
  }
  get textContent() {
    if (this._textContent) return this._textContent;
    if (this.children.length) return this.children.map((c) => c.textContent || "").join("");
    return "";
  }
  get outerHTML() {
    const attrs = [];
    const style = Object.entries(this.style)
      .filter(([, v]) => v !== undefined && v !== null && String(v) !== "")
      .map(([k, v]) => `${k.replace(/[A-Z]/g, (m) => "-" + m.toLowerCase())}:${v};`)
      .join("");
    if (style.length) attrs.push(`style="${escapeHtml(style)}"`);
    if (this.tagName === "a") {
      if (this.href !== undefined) attrs.push(`href="${escapeHtml(String(this.href))}"`);
      if (this.title !== undefined) attrs.push(`title="${escapeHtml(String(this.title))}"`);
    }
    if (this.tagName === "input") {
      if (this.type !== undefined) attrs.push(`type="${escapeHtml(String(this.type))}"`);
      if (this.name !== undefined) attrs.push(`name="${escapeHtml(String(this.name))}"`);
      if (this.value !== undefined) attrs.push(`value="${escapeHtml(String(this.value))}"`);
      if (this.size !== undefined) attrs.push(`size="${escapeHtml(String(this.size))}"`);
    }
    for (const [k, v] of Object.entries(this.attrs)) {
      attrs.push(`${k}="${escapeHtml(v)}"`);
    }
    const attrText = attrs.length ? " " + attrs.join(" ") : "";
    const body = this._innerHTML
      ? this._innerHTML
      : this.children.length
      ? this.children.map((c) => c.outerHTML || "").join("")
      : this._textContent
      ? escapeHtml(this._textContent)
      : "";
    if (this.tagName === "input" || this.tagName === "br" || this.tagName === "hr") {
      return `<${this.tagName}${attrText}>`;
    }
    return `<${this.tagName}${attrText}>${body}</${this.tagName}>`;
  }
}

globalThis.document = {
  getElementById: () => null,
  createElement: (tag) => new FakeElement(tag),
  createTextNode: (txt) => new FakeTextNode(txt),
  createDocumentFragment: () => ({ appendChild() {} }),
  body: { appendChild() {} },
  head: { appendChild() {} },
};
globalThis.DOMPurify = { sanitize: (x) => x };
