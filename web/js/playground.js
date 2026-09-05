/* Copyright Quad4 2026
 * SPDX-License-Identifier: 0BSD
 */

(function () {
  const TAB_NAME_LIMIT = 48;
  const STORAGE_KEY = "micron.parser.go.tabs.v2";
  const PREFS_KEY = "micron.parser.go.prefs.v1";
  const DEFAULT_TAB_NAME = "NomadNet Guide";
  const LANGUAGES_TAB_NAME = "Languages";
  const GUIDE_URL = "static/data/nomadnet_guide.mu";
  const LANGUAGES_URL = "static/data/languages.mu";
  const FONT_MIN = 0.75;
  const FONT_MAX = 1.25;
  const FONT_STEP = 0.0625;
  const MOBILE_MQ = "(max-width: 900px)";

  const input = document.getElementById("input");
  const lineNumbers = document.getElementById("line-numbers");
  const preview = document.getElementById("preview");
  const statusEl = document.getElementById("status");
  const tabList = document.getElementById("tab-list");
  const tabScroll = document.getElementById("tab-scroll");
  const tabAddBtn = document.getElementById("tab-add");
  const layout = document.querySelector(".layout");
  const splitter = document.getElementById("splitter");
  const ctxMenu = document.getElementById("context-menu");
  const repoSourceLink = document.getElementById("repo-source-link");
  const headerDownloadBtn = document.getElementById("header-download");
  const headerToggleSourceBtn = document.getElementById("header-toggle-source");
  const editorMeta = document.getElementById("editor-meta");
  const cursorStat = document.getElementById("cursor-stat");
  const renderStat = document.getElementById("render-stat");
  const toolWrap = document.getElementById("tool-wrap");
  const toolColorPickers = document.getElementById("tool-color-pickers");
  const toolFontDec = document.getElementById("tool-font-dec");
  const toolFontInc = document.getElementById("tool-font-inc");
  const toolReset = document.getElementById("tool-reset");
  const toolDownload = document.getElementById("tool-download");
  const previewWrap = document.getElementById("preview-wrap");
  const colorDecos = document.getElementById("color-decos");
  const editorMirror = document.getElementById("editor-mirror");
  const editorStack = document.getElementById("editor-stack");
  const viewBtns = Array.from(document.querySelectorAll(".view-btn"));
  const diagList = document.getElementById("diag-list");
  const diagCount = document.getElementById("diag-count");
  const diagPanel = document.getElementById("diag-panel");
  const diagToggle = document.getElementById("diag-toggle");
  const diagClose = document.getElementById("diag-close");
  const toolShowDiag = document.getElementById("tool-show-diag");

  let convert = null;
  let lintFn = null;
  let rafId = 0;
  let colorDecoRaf = 0;
  let seedContent = "";
  let languagesContent = "";
  let state = null;
  let prefs = loadPrefs();
  let partialRunID = 0;
  let ctxTabId = null;
  const partialIntervals = new Map();
  let measureCanvas = null;
  let colorPickBusy = false;

  function setStatus(msg, err) {
    statusEl.textContent = msg;
    statusEl.className = "status" + (err ? " err" : "");
  }

  function loadPrefs() {
    try {
      const raw = localStorage.getItem(PREFS_KEY);
      if (!raw) return defaultPrefs();
      const parsed = JSON.parse(raw);
      const view = parsed.mobileView;
      return {
        wrap: Boolean(parsed.wrap),
        fontSize: clampFont(Number(parsed.fontSize) || 0.875),
        mobileView: view === "source" || view === "preview" || view === "split" ? view : "split",
        diagCollapsed: Boolean(parsed.diagCollapsed),
        diagHidden: Boolean(parsed.diagHidden),
        colorPickers: parsed.colorPickers !== false,
      };
    } catch (_) {
      return defaultPrefs();
    }
  }

  function defaultPrefs() {
    return {
      wrap: false,
      fontSize: 0.875,
      mobileView: "split",
      diagCollapsed: false,
      diagHidden: false,
      colorPickers: true,
    };
  }

  function savePrefs() {
    try {
      localStorage.setItem(PREFS_KEY, JSON.stringify(prefs));
    } catch (_) {}
  }

  function clampFont(n) {
    return Math.min(FONT_MAX, Math.max(FONT_MIN, n));
  }

  function isMobileLayout() {
    return window.matchMedia(MOBILE_MQ).matches;
  }

  function applyPrefs() {
    input.classList.toggle("wrap-on", !!prefs.wrap);
    if (editorMirror) {
      editorMirror.classList.toggle("wrap-on", !!prefs.wrap);
    }
    toolWrap.classList.toggle("active", !!prefs.wrap);
    if (toolColorPickers) {
      toolColorPickers.classList.toggle("active", !!prefs.colorPickers);
    }
    if (editorStack) {
      editorStack.classList.toggle("colors-on", !!prefs.colorPickers);
    }
    if (colorDecos) {
      colorDecos.classList.toggle("color-decos-off", !prefs.colorPickers);
    }
    document.documentElement.style.setProperty("--editor-font-size", prefs.fontSize + "rem");
    applyMobileView();
    applyDiagPanel();
    updateLineNumbers();
    scheduleColorDecorations();
  }

  function applyDiagPanel() {
    if (!diagPanel) return;
    diagPanel.classList.toggle("diag-collapsed", !!prefs.diagCollapsed);
    diagPanel.classList.toggle("diag-hidden", !!prefs.diagHidden);
    if (diagToggle) {
      diagToggle.setAttribute("aria-expanded", prefs.diagCollapsed ? "false" : "true");
      diagToggle.title = prefs.diagCollapsed ? "Expand diagnostics" : "Collapse diagnostics";
    }
    if (toolShowDiag) {
      toolShowDiag.hidden = !prefs.diagHidden;
    }
  }

  function applyMobileView() {
    document.body.classList.remove("mview-source", "mview-preview", "mview-split");
    const view = prefs.mobileView || "split";
    document.body.classList.add("mview-" + view);
    for (const btn of viewBtns) {
      btn.classList.toggle("active", btn.getAttribute("data-view") === view);
      btn.setAttribute("aria-selected", btn.getAttribute("data-view") === view ? "true" : "false");
    }
  }

  function setMobileView(view) {
    if (view !== "source" && view !== "preview" && view !== "split") return;
    prefs.mobileView = view;
    if (view === "preview") {
      state.sourceHidden = false;
    }
    savePrefs();
    saveState();
    applySourceHidden();
    applyMobileView();
  }

  function activeTab() {
    const tab = state.tabs.find((t) => t.id === state.activeId);
    return tab || state.tabs[0];
  }

  function applySourceHidden() {
    document.body.classList.toggle("preview-only", !!state.sourceHidden && !isMobileLayout());
    if (headerToggleSourceBtn) {
      headerToggleSourceBtn.classList.toggle("active", !state.sourceHidden);
    }
  }

  function updateTabScrollFade() {
    if (!tabScroll) return;
    const max = tabScroll.scrollWidth - tabScroll.clientWidth;
    const atEnd = max <= 2 || tabScroll.scrollLeft >= max - 2;
    tabScroll.classList.toggle("at-end", atEnd);
  }

  function scrollActiveTabIntoView() {
    const el = tabList.querySelector('.tab-item.active');
    if (!el || !el.scrollIntoView) return;
    el.scrollIntoView({ inline: "nearest", block: "nearest", behavior: "smooth" });
    requestAnimationFrame(updateTabScrollFade);
  }

  function selectTab(id) {
    if (!state.tabs.some((t) => t.id === id)) return;
    state.activeId = id;
    saveState();
    syncEditorWithActiveTab();
  }

  function cycleTab(delta) {
    if (!state.tabs.length) return;
    const idx = state.tabs.findIndex((t) => t.id === state.activeId);
    const next = (idx + delta + state.tabs.length) % state.tabs.length;
    selectTab(state.tabs[next].id);
  }

  function renderTabs() {
    tabList.innerHTML = "";
    for (const tab of state.tabs) {
      const active = tab.id === state.activeId;
      const wrapEl = document.createElement("div");
      wrapEl.className = "tab-item" + (active ? " active" : "");
      wrapEl.dataset.tabId = tab.id;
      wrapEl.setAttribute("role", "presentation");

      const label = document.createElement("button");
      label.type = "button";
      label.className = "tab-btn tab-label";
      label.setAttribute("role", "tab");
      label.setAttribute("aria-selected", active ? "true" : "false");
      label.id = "tab-" + tab.id;
      label.textContent = tab.name;
      label.title = tab.name + " · double-click to rename · middle-click to close";
      label.addEventListener("click", (ev) => {
        ev.stopPropagation();
        selectTab(tab.id);
      });
      label.addEventListener("dblclick", (ev) => {
        ev.preventDefault();
        ev.stopPropagation();
        renameTab(tab);
      });
      label.addEventListener("auxclick", (ev) => {
        if (ev.button !== 1) return;
        ev.preventDefault();
        ev.stopPropagation();
        closeTabById(tab.id);
      });
      label.addEventListener("mousedown", (ev) => {
        if (ev.button === 1) ev.preventDefault();
      });
      let touchTap = 0;
      label.addEventListener("touchend", (ev) => {
        const now = Date.now();
        if (now - touchTap < 350) {
          ev.preventDefault();
          renameTab(tab);
          touchTap = 0;
        } else {
          touchTap = now;
        }
      }, { passive: false });

      const closeBtn = document.createElement("button");
      closeBtn.type = "button";
      closeBtn.className = "tab-close";
      closeBtn.setAttribute("aria-label", "Close " + tab.name);
      closeBtn.textContent = "\u00d7";
      closeBtn.addEventListener("click", (ev) => {
        ev.stopPropagation();
        closeTabById(tab.id);
      });

      wrapEl.addEventListener("contextmenu", (ev) => {
        ev.preventDefault();
        ev.stopPropagation();
        ctxTabId = tab.id;
        if (state.activeId !== tab.id) {
          selectTab(tab.id);
        }
        showContextMenu(ev.clientX, ev.clientY);
      });

      wrapEl.appendChild(label);
      wrapEl.appendChild(closeBtn);
      tabList.appendChild(wrapEl);
    }
    requestAnimationFrame(() => {
      scrollActiveTabIntoView();
      updateTabScrollFade();
    });
  }

  function renameTab(tab) {
    const next = prompt("Tab name (max " + TAB_NAME_LIMIT + " chars)", tab.name);
    if (next === null) return;
    const name = validName(next);
    if (!name) return;
    tab.name = name;
    saveState();
    renderTabs();
  }

  function closeTabById(id) {
    if (state.tabs.length === 1) {
      const t = state.tabs[0];
      t.content = "";
      t.name = validName("Untitled") || "Untitled";
      state.activeId = t.id;
      saveState();
      syncEditorWithActiveTab();
      return;
    }
    const idx = state.tabs.findIndex((t) => t.id === id);
    if (idx < 0) return;
    state.tabs.splice(idx, 1);
    if (state.activeId === id) {
      const next = state.tabs[Math.min(idx, state.tabs.length - 1)];
      state.activeId = next.id;
    }
    saveState();
    syncEditorWithActiveTab();
  }

  function addNewTab() {
    const id = makeID();
    const count = state.tabs.length + 1;
    state.tabs.push({ id: id, name: "Tab " + count, content: "" });
    state.activeId = id;
    saveState();
    syncEditorWithActiveTab();
    input.focus();
  }

  function syncEditorWithActiveTab() {
    const tab = activeTab();
    if (!tab) return;
    if (input.value !== tab.content) {
      input.value = tab.content;
    }
    applySourceHidden();
    renderTabs();
    updateLineNumbers();
    updateEditorMeta();
    updateCursorStat();
    scheduleRender();
    scheduleColorDecorations();
  }

  function saveState() {
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } catch (_) {}
  }

  function validName(name) {
    const trimmed = String(name || "").trim();
    if (!trimmed) return null;
    return trimmed.slice(0, TAB_NAME_LIMIT);
  }

  function makeID() {
    if (globalThis.crypto && typeof globalThis.crypto.randomUUID === "function") {
      return globalThis.crypto.randomUUID();
    }
    return "tab-" + Date.now() + "-" + Math.random().toString(16).slice(2);
  }

  function makeDefaultState() {
    const guideId = makeID();
    const langId = makeID();
    return {
      activeId: guideId,
      sourceHidden: false,
      tabs: [
        { id: guideId, name: DEFAULT_TAB_NAME, content: seedContent },
        { id: langId, name: LANGUAGES_TAB_NAME, content: languagesContent },
      ],
    };
  }

  function normalizeLoadedState(raw) {
    if (!raw || !Array.isArray(raw.tabs) || raw.tabs.length === 0) {
      return makeDefaultState();
    }
    const tabs = raw.tabs
      .map((t) => ({
        id: String(t.id || makeID()),
        name: validName(t.name) || "Untitled",
        content: String(t.content || ""),
      }))
      .filter((t) => t.id);
    if (tabs.length === 0) {
      return makeDefaultState();
    }
    let activeId = String(raw.activeId || "");
    if (!tabs.some((t) => t.id === activeId)) {
      activeId = tabs[0].id;
    }
    const sourceHidden = Boolean(raw.sourceHidden);
    return { activeId: activeId, sourceHidden: sourceHidden, tabs: tabs };
  }

  function loadState() {
    try {
      const raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return makeDefaultState();
      return normalizeLoadedState(JSON.parse(raw));
    } catch (_) {
      return makeDefaultState();
    }
  }

  function countLines(text) {
    if (!text) return 1;
    let n = 1;
    for (let i = 0; i < text.length; i++) {
      if (text.charCodeAt(i) === 10) n++;
    }
    return n;
  }

  function updateLineNumbers() {
    const n = countLines(input.value);
    const prevActive = lineNumbers.querySelector(".ln.active");
    const prevLine = prevActive ? prevActive.getAttribute("data-line") : "";
    lineNumbers.innerHTML = "";
    for (let i = 1; i <= n; i++) {
      const btn = document.createElement("button");
      btn.type = "button";
      btn.className = "ln";
      btn.dataset.line = String(i);
      btn.textContent = String(i);
      btn.title = "Go to preview for line " + i;
      btn.setAttribute("aria-label", "Source line " + i);
      if (String(i) === prevLine) {
        btn.classList.add("active");
      }
      btn.addEventListener("click", (ev) => {
        ev.preventDefault();
        jumpSourceLineToPreview(i);
      });
      lineNumbers.appendChild(btn);
    }
    lineNumbers.scrollTop = input.scrollTop;
    const digits = String(n).length;
    lineNumbers.style.minWidth = Math.max(2.75, 1.4 + digits * 0.55) + "rem";
  }

  function selectSourceLine(line) {
    const text = input.value;
    let start = 0;
    let cur = 1;
    while (cur < line && start <= text.length) {
      const nl = text.indexOf("\n", start);
      if (nl < 0) {
        start = text.length;
        break;
      }
      start = nl + 1;
      cur++;
    }
    let end = text.indexOf("\n", start);
    if (end < 0) end = text.length;
    input.focus();
    input.setSelectionRange(start, end);
    const lineHeight = parseFloat(getComputedStyle(input).lineHeight) || 16;
    const topPad = parseFloat(getComputedStyle(input).paddingTop) || 0;
    input.scrollTop = Math.max(0, (line - 1) * lineHeight - input.clientHeight / 3 + topPad);
    updateCursorStat();
    for (const el of lineNumbers.querySelectorAll(".ln")) {
      el.classList.toggle("active", el.dataset.line === String(line));
    }
  }

  function jumpSourceLineToPreview(line) {
    selectSourceLine(line);
    preview.querySelectorAll(".mu-line-flash").forEach((el) => el.classList.remove("mu-line-flash"));
    let el = preview.querySelector('[data-mu-line="' + line + '"]');
    if (!el) {
      const nodes = preview.querySelectorAll("[data-mu-line]");
      let best = null;
      let bestLine = -1;
      for (const n of nodes) {
        const ln = parseInt(n.getAttribute("data-mu-line"), 10);
        if (!Number.isFinite(ln)) continue;
        if (ln <= line && ln >= bestLine) {
          best = n;
          bestLine = ln;
        }
      }
      el = best;
    }
    if (!el) return;
    el.classList.add("mu-line-flash");
    el.scrollIntoView({ block: "center", behavior: "smooth" });
    globalThis.setTimeout(() => el.classList.remove("mu-line-flash"), 1200);
  }

  function applyPreviewChrome() {
    if (!previewWrap) return;
    const root = preview.firstElementChild;
    if (!root) {
      previewWrap.style.backgroundColor = "";
      previewWrap.style.color = "";
      return;
    }
    const bg = root.style.backgroundColor || "";
    const fg = root.style.color || "";
    previewWrap.style.backgroundColor = bg || "";
    previewWrap.style.color = fg || "";
  }

  function updateEditorMeta() {
    const text = input.value;
    const lines = countLines(text);
    editorMeta.textContent = lines + " lines · " + text.length + " chars";
  }

  function updateCursorStat() {
    const pos = input.selectionStart || 0;
    const before = input.value.slice(0, pos);
    const line = countLines(before);
    const col = before.length - before.lastIndexOf("\n");
    cursorStat.textContent = "Ln " + line + ", Col " + col;
  }

  function isHexColorToken(s) {
    return /^[0-9a-fA-F]{3}$|^[0-9a-fA-F]{6}$/.test(s);
  }

  function expandHexToCss(hex) {
    const h = String(hex || "").toLowerCase();
    if (h.length === 3) {
      return "#" + h[0] + h[0] + h[1] + h[1] + h[2] + h[2];
    }
    if (h.length === 6) {
      return "#" + h;
    }
    return "#000000";
  }

  function cssToMicronHex(css, preferLen) {
    let h = String(css || "").replace(/^#/, "").toLowerCase();
    if (!/^[0-9a-f]{6}$/.test(h)) return "000000";
    if (preferLen === 3 && h[0] === h[1] && h[2] === h[3] && h[4] === h[5]) {
      return h[0] + h[2] + h[4];
    }
    return h;
  }

  function offsetToLineColZero(text, offset) {
    let line = 0;
    let col = 0;
    const end = Math.max(0, Math.min(offset, text.length));
    for (let i = 0; i < end; i++) {
      if (text.charCodeAt(i) === 10) {
        line++;
        col = 0;
      } else {
        col++;
      }
    }
    return { line: line, col: col };
  }

  function findColorTokens(text) {
    const out = [];
    const headerRe = /^#!((?:fg|bg))=(\S+)/gm;
    let m;
    while ((m = headerRe.exec(text)) !== null) {
      const value = m[2];
      const valueStart = m.index + m[0].length - value.length;
      out.push({
        kind: m[1],
        value: value,
        start: valueStart,
        end: valueStart + value.length,
        valid: isHexColorToken(value),
      });
    }
    const inlineRe = /`([FB])([0-9a-fA-F]{3}|[0-9a-fA-F]{6})(?![0-9a-fA-F])/g;
    while ((m = inlineRe.exec(text)) !== null) {
      const value = m[2];
      const valueStart = m.index + 2;
      out.push({
        kind: m[1] === "F" ? "fg" : "bg",
        value: value,
        start: valueStart,
        end: valueStart + value.length,
        valid: true,
      });
    }
    out.sort((a, b) => a.start - b.start);
    return out;
  }

  function escapeMirrorHtml(s) {
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;");
  }

  function buildMirrorHtml(text, tokens) {
    if (!tokens.length) {
      return escapeMirrorHtml(text);
    }
    let html = "";
    let cursor = 0;
    for (const tok of tokens) {
      if (tok.start < cursor) continue;
      html += escapeMirrorHtml(text.slice(cursor, tok.start));
      html += '<span class="mu-color-slot">' + escapeMirrorHtml(text.slice(tok.start, tok.end)) + "</span>";
      cursor = tok.end;
    }
    html += escapeMirrorHtml(text.slice(cursor));
    return html;
  }

  function syncEditorMirror(text, tokens) {
    if (!editorMirror) return;
    if (!prefs.colorPickers) {
      editorMirror.textContent = "";
      return;
    }
    editorMirror.innerHTML = buildMirrorHtml(text, tokens || findColorTokens(text));
    editorMirror.scrollTop = input.scrollTop;
    editorMirror.scrollLeft = input.scrollLeft;
  }

  function editorMetrics() {
    const cs = getComputedStyle(input);
    if (!measureCanvas) {
      measureCanvas = document.createElement("canvas");
    }
    const ctx = measureCanvas.getContext("2d");
    ctx.font = cs.font;
    const charW = ctx.measureText("M").width || 8;
    let lineH = parseFloat(cs.lineHeight);
    if (!Number.isFinite(lineH) || lineH <= 0) {
      lineH = (parseFloat(cs.fontSize) || 14) * 1.45;
    }
    return {
      charW: charW,
      lineH: lineH,
      padT: parseFloat(cs.paddingTop) || 0,
      padL: parseFloat(cs.paddingLeft) || 0,
    };
  }

  function positionColorSwatch(el, start, valueLen, metrics) {
    const pos = offsetToLineColZero(input.value, start);
    const size = Math.max(metrics.lineH * 0.78, 12);
    const top = metrics.padT + pos.line * metrics.lineH - input.scrollTop + (metrics.lineH - size) * 0.5;
    const left = metrics.padL + pos.col * metrics.charW - input.scrollLeft;
    el.style.top = top + "px";
    el.style.left = left + "px";
    el.style.width = "";
    el.style.height = "";
    return { top: top, left: left, width: size, height: size };
  }

  function replaceColorToken(start, end, nextHex, activeEl) {
    const v = input.value;
    if (start < 0 || end > v.length || start >= end) return;
    const selStart = input.selectionStart;
    const selEnd = input.selectionEnd;
    const delta = nextHex.length - (end - start);
    input.value = v.slice(0, start) + nextHex + v.slice(end);
    let ns = selStart;
    let ne = selEnd;
    if (selStart >= end) ns += delta;
    else if (selStart > start) ns = start + nextHex.length;
    if (selEnd >= end) ne += delta;
    else if (selEnd > start) ne = start + nextHex.length;
    try {
      input.setSelectionRange(ns, ne);
    } catch (_) {}
    if (activeEl) {
      activeEl.dataset.end = String(start + nextHex.length);
      activeEl.title = (activeEl.dataset.kind === "bg" ? "Background" : "Foreground") +
        " #" + nextHex;
      activeEl.setAttribute("aria-label", activeEl.title);
      activeEl.classList.remove("color-swatch-invalid");
      positionColorSwatch(activeEl, start, nextHex.length, editorMetrics());
    }
    colorPickBusy = true;
    onEditorChange();
    syncEditorMirror(input.value);
    colorPickBusy = false;
  }

  function clearColorDecorations() {
    if (colorDecos) colorDecos.innerHTML = "";
    if (editorMirror) editorMirror.textContent = "";
  }

  function updateColorDecorations() {
    if (!colorDecos) return;
    if (!prefs.colorPickers) {
      clearColorDecorations();
      return;
    }
    if (colorPickBusy) return;
    const active = document.activeElement;
    if (active && active.classList && active.classList.contains("color-swatch") &&
        colorDecos.contains(active)) {
      syncEditorMirror(input.value);
      return;
    }
    const text = input.value;
    const tokens = findColorTokens(text);
    syncEditorMirror(text, tokens);
    const metrics = editorMetrics();
    const viewH = input.clientHeight;
    const viewW = input.clientWidth;
    const frag = document.createDocumentFragment();
    const maxVisible = 200;
    let shown = 0;

    for (const tok of tokens) {
      if (shown >= maxVisible) break;
      const el = document.createElement("input");
      el.type = "color";
      el.className = "color-swatch" + (tok.valid ? "" : " color-swatch-invalid");
      el.value = expandHexToCss(tok.valid ? tok.value : "000000");
      el.title = (tok.kind === "bg" ? "Background" : "Foreground") + " #" + tok.value +
        (tok.valid ? "" : " (invalid, pick to fix)");
      el.setAttribute("aria-label", el.title);
      el.dataset.start = String(tok.start);
      el.dataset.end = String(tok.end);
      el.dataset.kind = tok.kind;
      el.dataset.preferLen = String(tok.valid && tok.value.length === 3 ? 3 : 6);

      const box = positionColorSwatch(el, tok.start, tok.value.length, metrics);
      if (box.top + box.height < -4 || box.top > viewH + 4) continue;
      if (box.left + box.width < -4 || box.left > viewW + 4) continue;

      el.addEventListener("input", () => {
        const start = Number(el.dataset.start);
        const end = Number(el.dataset.end);
        const prefer = Number(el.dataset.preferLen) || 6;
        const next = cssToMicronHex(el.value, prefer);
        replaceColorToken(start, end, next, el);
      });
      el.addEventListener("change", () => {
        scheduleColorDecorations();
      });
      el.addEventListener("blur", () => {
        scheduleColorDecorations();
      });
      el.addEventListener("mousedown", (ev) => {
        ev.stopPropagation();
      });
      el.addEventListener("click", (ev) => {
        ev.stopPropagation();
      });

      frag.appendChild(el);
      shown++;
    }

    colorDecos.innerHTML = "";
    colorDecos.appendChild(frag);
  }

  function scheduleColorDecorations() {
    if (colorDecoRaf) {
      cancelAnimationFrame(colorDecoRaf);
    }
    colorDecoRaf = requestAnimationFrame(() => {
      colorDecoRaf = 0;
      updateColorDecorations();
    });
  }

  function renderPreview() {
    if (typeof convert !== "function") {
      return;
    }
    try {
      clearPartialIntervals();
      partialRunID += 1;
      const t0 = performance.now();
      preview.innerHTML = convert(input.value, true, true);
      const ms = performance.now() - t0;
      applyPreviewChrome();
      hydratePartials(partialRunID);
      renderStat.textContent = ms.toFixed(1) + " ms";
      renderDiagnostics();
      setStatus("");
    } catch (_) {
      setStatus("WASM failed", true);
      renderStat.textContent = "";
    }
  }

  function severityLabel(n) {
    if (n === 2) return "error";
    if (n === 1) return "warning";
    return "info";
  }

  function offsetToPos(text, offset) {
    let line = 1;
    let col = 1;
    const end = Math.max(0, Math.min(offset, text.length));
    for (let i = 0; i < end; i++) {
      if (text.charCodeAt(i) === 10) {
        line++;
        col = 1;
      } else {
        col++;
      }
    }
    return { line: line, col: col };
  }

  function renderDiagnostics() {
    if (!diagList || !diagCount) return;
    diagList.innerHTML = "";
    let diags = [];
    if (typeof lintFn !== "function") {
      diagCount.textContent = "n/a";
      const li = document.createElement("li");
      li.className = "diag-item diag-warning";
      li.textContent = "micronLint missing - rebuild WASM (make wasm) and hard-refresh; old cached micron.wasm has no lint export";
      diagList.appendChild(li);
      return;
    }
    try {
      const raw = lintFn(input.value);
      diags = JSON.parse(raw || "[]");
      if (!Array.isArray(diags)) diags = [];
    } catch (err) {
      diagCount.textContent = "!";
      const li = document.createElement("li");
      li.className = "diag-item diag-error";
      li.textContent = "lint failed: " + (err && err.message ? err.message : String(err));
      diagList.appendChild(li);
      return;
    }
    diagCount.textContent = String(diags.length);
    for (const d of diags) {
      const li = document.createElement("li");
      li.className = "diag-item diag-" + severityLabel(d.severity);
      const pos = offsetToPos(input.value, (d.span && d.span.start) || 0);
      li.textContent = severityLabel(d.severity) + ": " + (d.code || "") + " - " + (d.message || "") +
        " (Ln " + pos.line + ", Col " + pos.col + ")";
      li.tabIndex = 0;
      li.addEventListener("click", () => {
        const start = (d.span && d.span.start) || 0;
        const end = (d.span && d.span.end) || start;
        input.focus();
        input.setSelectionRange(start, end);
        updateCursorStat();
      });
      diagList.appendChild(li);
    }
  }

  function clearPartialIntervals() {
    for (const id of partialIntervals.values()) {
      clearInterval(id);
    }
    partialIntervals.clear();
  }

  function isFetchableURL(url) {
    const u = String(url || "").trim();
    return u.startsWith("/") || u.startsWith("./") || u.startsWith("../") ||
      u.startsWith("http://") || u.startsWith("https://");
  }

  async function renderPartialNode(node, runID) {
    const raw = String(node.getAttribute("data-partial-url") || "").trim();
    node.classList.add("pending");
    if (!isFetchableURL(raw)) {
      node.textContent = "[partial unsupported URL: " + raw + "]";
      node.classList.remove("pending");
      node.classList.add("error");
      return;
    }
    try {
      const res = await fetch(raw, { cache: "no-store" });
      if (!res.ok) {
        throw new Error("HTTP " + res.status);
      }
      const markup = await res.text();
      if (runID !== partialRunID) {
        return;
      }
      node.innerHTML = convert(markup, true, true);
      node.classList.remove("error");
    } catch (_) {
      node.textContent = "[partial load failed: " + raw + "]";
      node.classList.add("error");
    } finally {
      node.classList.remove("pending");
    }
  }

  function hydratePartials(runID) {
    const nodes = preview.querySelectorAll(".Mu-partial[data-partial-url]");
    for (const node of nodes) {
      renderPartialNode(node, runID);
      const refreshRaw = String(node.getAttribute("data-partial-refresh") || "");
      const refresh = parseInt(refreshRaw, 10);
      if (Number.isFinite(refresh) && refresh > 0) {
        const intervalID = setInterval(() => {
          if (runID !== partialRunID) {
            return;
          }
          renderPartialNode(node, runID);
        }, refresh * 1000);
        partialIntervals.set(node, intervalID);
      }
    }
  }

  function scheduleRender() {
    if (rafId) {
      cancelAnimationFrame(rafId);
    }
    rafId = requestAnimationFrame(function tick() {
      rafId = 0;
      renderPreview();
    });
  }

  function onEditorChange() {
    const tab = activeTab();
    if (!tab) return;
    tab.content = input.value;
    saveState();
    updateLineNumbers();
    updateEditorMeta();
    updateCursorStat();
    scheduleRender();
    if (!colorPickBusy) {
      scheduleColorDecorations();
    }
  }

  input.addEventListener("input", onEditorChange);
  input.addEventListener("scroll", () => {
    lineNumbers.scrollTop = input.scrollTop;
    if (editorMirror && prefs.colorPickers) {
      editorMirror.scrollTop = input.scrollTop;
      editorMirror.scrollLeft = input.scrollLeft;
    }
    scheduleColorDecorations();
  });
  input.addEventListener("keyup", updateCursorStat);
  input.addEventListener("click", updateCursorStat);
  input.addEventListener("select", updateCursorStat);

  input.addEventListener("keydown", (ev) => {
    if (ev.key !== "Tab" || ev.ctrlKey || ev.metaKey || ev.altKey) return;
    ev.preventDefault();
    const start = input.selectionStart;
    const end = input.selectionEnd;
    input.setRangeText("  ", start, end, "end");
    onEditorChange();
  });

  preview.addEventListener("click", (ev) => {
    const link = ev.target && ev.target.closest ? ev.target.closest("a[data-action='openNode']") : null;
    if (!link) return;
    ev.preventDefault();
    if (typeof globalThis.micronResolveLink !== "function") {
      return;
    }
    const payloadRaw = globalThis.micronResolveLink(
      "#preview",
      String(link.getAttribute("data-destination") || ""),
      String(link.getAttribute("data-fields") || "")
    );
    let payload = null;
    try {
      payload = JSON.parse(String(payloadRaw || "{}"));
    } catch (_) {
      payload = null;
    }
    if (typeof globalThis.onMicronLink === "function") {
      try {
        globalThis.onMicronLink(payload, link);
      } catch (_) {}
    }
  });

  function upgradeInputToTextarea(inputEl, options) {
    if (!inputEl || !inputEl.tagName || inputEl.tagName !== "INPUT") return inputEl;
    const owner = inputEl.ownerDocument || document;
    if (!owner) return inputEl;
    const inputType = (inputEl.type || "").toLowerCase();
    if (inputType === "password") return inputEl;
    const opts = options || {};
    const ta = owner.createElement("textarea");
    ta.name = inputEl.name;
    const currentValue = (typeof inputEl.value === "string" && inputEl.value.length > 0)
      ? inputEl.value
      : (inputEl.getAttribute("value") || "");
    ta.value = currentValue;
    const cols = (typeof opts.cols === "number") ? opts.cols
      : (inputEl.size > 0 ? inputEl.size : null);
    if (cols) ta.cols = cols;
    if (typeof opts.rows === "number") ta.rows = opts.rows;
    else if (!ta.rows || ta.rows < 2) ta.rows = 4;
    if (opts.wrap) ta.wrap = opts.wrap;
    if (inputEl.disabled) ta.disabled = true;
    if (inputEl.readOnly) ta.readOnly = true;
    if (inputEl.placeholder) ta.placeholder = inputEl.placeholder;
    if (inputEl.required) ta.required = true;
    if (inputEl.autocomplete) ta.autocomplete = inputEl.autocomplete;
    if (inputEl.style && inputEl.style.cssText) ta.style.cssText = inputEl.style.cssText;
    const skipAttrs = new Set(["type", "value", "size", "name"]);
    for (const attr of Array.from(inputEl.attributes || [])) {
      if (skipAttrs.has(attr.name)) continue;
      if (attr.name === "style") continue;
      try { ta.setAttribute(attr.name, attr.value); } catch (e) {}
    }
    if (inputEl.classList && inputEl.classList.length > 0) {
      ta.className = inputEl.className;
    }
    const wasFocused = (owner.activeElement === inputEl);
    const selStart = (typeof inputEl.selectionStart === "number") ? inputEl.selectionStart : null;
    const selEnd = (typeof inputEl.selectionEnd === "number") ? inputEl.selectionEnd : null;
    ta.setAttribute("data-micron-multiline", "1");
    ta.setAttribute("data-micron-original-tag", "input");
    ta.setAttribute("data-micron-original-type", inputEl.type || "text");
    inputEl.replaceWith(ta);
    if (wasFocused && typeof ta.focus === "function") {
      try {
        ta.focus();
        if (selStart !== null && selEnd !== null && typeof ta.setSelectionRange === "function") {
          ta.setSelectionRange(selStart, selEnd);
        }
      } catch (e) {}
    }
    ta.dispatchEvent(new CustomEvent("micron-field-upgraded", {
      bubbles: true,
      detail: { from: "input", to: "textarea", element: ta, previous: inputEl }
    }));
    return ta;
  }

  function enableDoubleEnterMultiline(root, options) {
    if (!root || typeof root.addEventListener !== "function") return function () {};
    const opts = options || {};
    const windowMs = typeof opts.windowMs === "number" ? opts.windowMs : 500;
    const rows = typeof opts.rows === "number" ? opts.rows : 4;
    const filter = typeof opts.filter === "function" ? opts.filter : null;
    const suppressFirst = opts.suppressFirstEnter !== false;
    const lastEnter = new WeakMap();
    const armTimers = new WeakMap();
    const hintEl = document.getElementById("multiline-hint");
    let hintTimer = 0;
    const disarm = (el) => {
      if (lastEnter.has(el)) {
        lastEnter.delete(el);
        try {
          el.dispatchEvent(new CustomEvent("micron-multiline-disarmed", {
            bubbles: true,
            detail: { element: el }
          }));
        } catch (_) {}
      }
      const t = armTimers.get(el);
      if (t) {
        clearTimeout(t);
        armTimers.delete(el);
      }
      if (hintEl) {
        hintEl.classList.remove("visible");
        clearTimeout(hintTimer);
      }
    };
    const onKey = (e) => {
      if (e.key !== "Enter" || e.isComposing) return;
      if (e.shiftKey || e.ctrlKey || e.metaKey || e.altKey) return;
      const el = e.target;
      if (!el || el.tagName !== "INPUT") return;
      const t = (el.type || "text").toLowerCase();
      if (t === "password") return;
      if (t !== "text" && t !== "") return;
      if (filter && !filter(el)) return;
      const now = Date.now();
      const prev = lastEnter.get(el) || 0;
      if (prev > 0 && (now - prev) <= windowMs) {
        e.preventDefault();
        disarm(el);
        const cursor = (typeof el.selectionStart === "number") ? el.selectionStart : el.value.length;
        const ta = upgradeInputToTextarea(el, { rows });
        const before = (ta.value || "").slice(0, cursor);
        const after = (ta.value || "").slice(cursor);
        ta.value = before + "\n" + after;
        try {
          ta.focus();
          if (typeof ta.setSelectionRange === "function") {
            ta.setSelectionRange(before.length + 1, before.length + 1);
          }
        } catch (_) {}
        try {
          ta.dispatchEvent(new CustomEvent("micron-field-multiline-enabled", {
            bubbles: true,
            detail: { element: ta, trigger: "double-enter" }
          }));
        } catch (_) {}
        return;
      }
      if (suppressFirst) e.preventDefault();
      lastEnter.set(el, now);
      try {
        el.dispatchEvent(new CustomEvent("micron-multiline-armed", {
          bubbles: true,
          detail: { element: el, windowMs }
        }));
      } catch (_) {}
      if (hintEl) {
        hintEl.classList.add("visible");
        clearTimeout(hintTimer);
        hintTimer = setTimeout(() => {
          hintEl.classList.remove("visible");
        }, windowMs + 200);
      }
      const tid = setTimeout(() => disarm(el), windowMs + 16);
      armTimers.set(el, tid);
    };
    const onBlur = (e) => {
      if (e.target && e.target.tagName === "INPUT") disarm(e.target);
    };
    root.addEventListener("keydown", onKey);
    root.addEventListener("blur", onBlur, true);
    return function detach() {
      root.removeEventListener("keydown", onKey);
      root.removeEventListener("blur", onBlur, true);
    };
  }

  function sanitizeFileBase(name) {
    let s = String(name || "untitled").trim() || "untitled";
    s = s.replace(/[\\/:*?"<>|]+/g, "-").replace(/\s+/g, " ").trim();
    if (!s) s = "untitled";
    return s.slice(0, 120);
  }

  function downloadCurrentTab() {
    const tab = activeTab();
    if (!tab) return;
    const blob = new Blob([tab.content], { type: "text/plain;charset=utf-8" });
    const a = document.createElement("a");
    a.href = URL.createObjectURL(blob);
    a.download = sanitizeFileBase(tab.name) + ".mu";
    a.click();
    URL.revokeObjectURL(a.href);
  }

  function copySource() {
    const tab = activeTab();
    if (!tab) return;
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(tab.content).catch(() => {});
    }
  }

  function toggleSourcePane() {
    if (isMobileLayout()) {
      setMobileView(prefs.mobileView === "preview" ? "source" : "preview");
      return;
    }
    state.sourceHidden = !state.sourceHidden;
    saveState();
    applySourceHidden();
  }

  function openRepoInNewTab() {
    const href = repoSourceLink && repoSourceLink.href;
    if (href) {
      globalThis.open(href, "_blank", "noopener,noreferrer");
    }
  }

  function resetToDefaultTabs() {
    state = makeDefaultState();
    saveState();
    syncEditorWithActiveTab();
  }

  function ctxTargetTab() {
    if (ctxTabId) {
      const t = state.tabs.find((x) => x.id === ctxTabId);
      if (t) return t;
    }
    return activeTab();
  }

  if (headerDownloadBtn) headerDownloadBtn.addEventListener("click", () => downloadCurrentTab());
  if (toolDownload) toolDownload.addEventListener("click", () => downloadCurrentTab());
  if (headerToggleSourceBtn) headerToggleSourceBtn.addEventListener("click", () => toggleSourcePane());
  if (tabAddBtn) tabAddBtn.addEventListener("click", () => addNewTab());
  if (tabScroll) tabScroll.addEventListener("scroll", updateTabScrollFade, { passive: true });

  toolWrap.addEventListener("click", () => {
    prefs.wrap = !prefs.wrap;
    savePrefs();
    applyPrefs();
  });
  if (toolColorPickers) {
    toolColorPickers.addEventListener("click", () => {
      prefs.colorPickers = !prefs.colorPickers;
      savePrefs();
      applyPrefs();
    });
  }
  toolFontDec.addEventListener("click", () => {
    prefs.fontSize = clampFont(prefs.fontSize - FONT_STEP);
    savePrefs();
    applyPrefs();
  });
  toolFontInc.addEventListener("click", () => {
    prefs.fontSize = clampFont(prefs.fontSize + FONT_STEP);
    savePrefs();
    applyPrefs();
  });
  if (toolReset) toolReset.addEventListener("click", () => resetToDefaultTabs());
  if (diagToggle) {
    diagToggle.addEventListener("click", () => {
      prefs.diagCollapsed = !prefs.diagCollapsed;
      savePrefs();
      applyDiagPanel();
    });
  }
  if (diagClose) {
    diagClose.addEventListener("click", () => {
      prefs.diagHidden = true;
      savePrefs();
      applyDiagPanel();
    });
  }
  if (toolShowDiag) {
    toolShowDiag.addEventListener("click", () => {
      prefs.diagHidden = false;
      prefs.diagCollapsed = false;
      savePrefs();
      applyDiagPanel();
      renderDiagnostics();
    });
  }

  for (const btn of viewBtns) {
    btn.addEventListener("click", () => setMobileView(btn.getAttribute("data-view")));
  }

  window.matchMedia(MOBILE_MQ).addEventListener("change", () => {
    applySourceHidden();
    applyMobileView();
    scheduleColorDecorations();
  });

  window.addEventListener("resize", () => {
    scheduleColorDecorations();
  });

  function hideContextMenu() {
    ctxMenu.style.display = "none";
    ctxMenu.setAttribute("aria-hidden", "true");
    ctxTabId = null;
  }

  function showContextMenu(clientX, clientY) {
    ctxMenu.style.display = "block";
    ctxMenu.setAttribute("aria-hidden", "false");
    const pad = 8;
    let x = clientX;
    let y = clientY;
    ctxMenu.style.left = "0";
    ctxMenu.style.top = "0";
    const w = ctxMenu.offsetWidth;
    const h = ctxMenu.offsetHeight;
    if (x + w + pad > window.innerWidth) x = window.innerWidth - w - pad;
    if (y + h + pad > window.innerHeight) y = window.innerHeight - h - pad;
    ctxMenu.style.left = Math.max(pad, x) + "px";
    ctxMenu.style.top = Math.max(pad, y) + "px";
  }

  document.addEventListener("click", (ev) => {
    if (ev.target.closest && ev.target.closest("#context-menu")) return;
    hideContextMenu();
  });

  ctxMenu.querySelectorAll("button[data-ctx]").forEach((btn) => {
    btn.addEventListener("click", (ev) => {
      ev.stopPropagation();
      const act = btn.getAttribute("data-ctx");
      const target = ctxTargetTab();
      hideContextMenu();
      if (act === "rename") {
        if (target) renameTab(target);
      } else if (act === "close") {
        if (target) closeTabById(target.id);
      } else if (act === "new") {
        addNewTab();
      } else if (act === "toggle-source") {
        toggleSourcePane();
      } else if (act === "open-repo") {
        openRepoInNewTab();
      } else if (act === "download") {
        downloadCurrentTab();
      } else if (act === "copy") {
        copySource();
      }
    });
  });

  document.addEventListener("contextmenu", (ev) => {
    if (ev.target.closest && ev.target.closest("#context-menu")) return;
    if (ev.target.closest && ev.target.closest(".tab-item")) return;
    ev.preventDefault();
    ctxTabId = state && state.activeId ? state.activeId : null;
    showContextMenu(ev.clientX, ev.clientY);
  });

  window.addEventListener("keydown", (ev) => {
    if (ev.key === "Escape") {
      hideContextMenu();
      return;
    }
    const mod = ev.ctrlKey || ev.metaKey;
    const inField = ev.target === input || (ev.target && (ev.target.tagName === "TEXTAREA" || ev.target.tagName === "INPUT"));

    if (mod && !ev.altKey && (ev.key === "Tab" || ev.key === "PageDown" || ev.key === "PageUp")) {
      ev.preventDefault();
      if (ev.key === "PageUp" || (ev.key === "Tab" && ev.shiftKey)) {
        cycleTab(-1);
      } else {
        cycleTab(1);
      }
      return;
    }

    if (mod && ev.key.toLowerCase() === "t") {
      ev.preventDefault();
      addNewTab();
      return;
    }
    if (mod && ev.key.toLowerCase() === "w") {
      ev.preventDefault();
      const t = activeTab();
      if (t) closeTabById(t.id);
      return;
    }
    if (mod && ev.shiftKey && ev.key.toLowerCase() === "d") {
      ev.preventDefault();
      downloadCurrentTab();
      return;
    }
    if (mod && ev.key.toLowerCase() === "b") {
      ev.preventDefault();
      toggleSourcePane();
      return;
    }
    if (mod && ev.key.toLowerCase() === "c" && inField && ev.target === input && input.selectionStart === input.selectionEnd) {
      ev.preventDefault();
      copySource();
    }
  }, true);

  let dragging = false;
  function setSplitByClientX(clientX) {
    const rect = layout.getBoundingClientRect();
    if (rect.width <= 0) return;
    const pct = ((clientX - rect.left) / rect.width) * 100;
    const clamped = Math.max(20, Math.min(80, pct));
    layout.style.setProperty("--left-size", clamped + "%");
  }
  function nudgeSplit(deltaPct) {
    const cur = parseFloat(getComputedStyle(layout).getPropertyValue("--left-size")) || 50;
    const next = Math.max(20, Math.min(80, cur + deltaPct));
    layout.style.setProperty("--left-size", next + "%");
  }
  splitter.addEventListener("pointerdown", (ev) => {
    if (isMobileLayout()) return;
    dragging = true;
    splitter.setPointerCapture(ev.pointerId);
    ev.preventDefault();
  });
  splitter.addEventListener("pointermove", (ev) => {
    if (!dragging || isMobileLayout()) return;
    setSplitByClientX(ev.clientX);
    scheduleColorDecorations();
  });
  splitter.addEventListener("pointerup", (ev) => {
    dragging = false;
    if (splitter.hasPointerCapture(ev.pointerId)) {
      splitter.releasePointerCapture(ev.pointerId);
    }
  });
  splitter.addEventListener("pointercancel", () => {
    dragging = false;
  });
  splitter.addEventListener("keydown", (ev) => {
    if (isMobileLayout()) return;
    if (ev.key === "ArrowLeft") {
      ev.preventDefault();
      nudgeSplit(-2);
    } else if (ev.key === "ArrowRight") {
      ev.preventDefault();
      nudgeSplit(2);
    }
  });

  async function loadGuideSeed() {
    try {
      const res = await fetch(GUIDE_URL, { cache: "force-cache" });
      if (!res.ok) throw new Error("HTTP " + res.status);
      seedContent = await res.text();
    } catch (_) {
      seedContent = "`c`!Micron-Parser-Go`!\n`a\n\n# Guide failed to load. Paste Micron markup here.\n";
    }
    try {
      const res = await fetch(LANGUAGES_URL, { cache: "no-store" });
      if (!res.ok) throw new Error("HTTP " + res.status);
      languagesContent = await res.text();
    } catch (_) {
      languagesContent = "> Languages\n# languages.mu failed to load\n";
    }
  }

  function bootWasm() {
    if (typeof Go === "undefined") {
      setStatus("WASM failed", true);
      return;
    }
    const go = new Go();
    // Bust browser / CDN caches so micronLint from a new build is picked up.
    const wasmURL = "micron.wasm?v=" + encodeURIComponent(String(Date.now()));
    WebAssembly.instantiateStreaming(fetch(wasmURL, { cache: "no-store" }), go.importObject)
      .then((result) => {
        go.run(result.instance);
        convert = globalThis.micronConvert;
        lintFn = globalThis.micronLint;
        if (typeof convert !== "function") {
          setStatus("WASM failed", true);
          return;
        }
        if (typeof lintFn !== "function") {
          setStatus("Lint unavailable", true);
        } else {
          setStatus("");
        }
        enableDoubleEnterMultiline(preview);
        renderPreview();
      })
      .catch(() => {
        setStatus("WASM failed", true);
      });
  }

  applyPrefs();
  loadGuideSeed().then(() => {
    state = loadState();
    syncEditorWithActiveTab();
    saveState();
    bootWasm();
  });
})();
