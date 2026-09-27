// Upgrades exercise textareas to CodeMirror editors with an optional vim
// mode. The textarea stays in the form as the source of truth, so Run and
// Submit (htmx or plain form posts) work unchanged. Without JavaScript the
// plain textarea is used.
//
// Build: npm ci && npm run build  (writes ../static/editor.js)

import { EditorState, Compartment } from "@codemirror/state";
import { EditorView, keymap, lineNumbers, drawSelection, highlightActiveLine, highlightActiveLineGutter } from "@codemirror/view";
import { defaultKeymap, history, historyKeymap, indentWithTab } from "@codemirror/commands";
import { StreamLanguage, syntaxHighlighting, indentUnit, bracketMatching, indentOnInput } from "@codemirror/language";
import { go } from "@codemirror/legacy-modes/mode/go";
import { classHighlighter } from "@lezer/highlight";
import { vim, Vim } from "@replit/codemirror-vim";

const storageKey = "goland-vim-mode";

function vimEnabled() {
	try {
		return localStorage.getItem(storageKey) === "on";
	} catch {
		return false;
	}
}

function rememberVim(on) {
	try {
		localStorage.setItem(storageKey, on ? "on" : "off");
	} catch {
		// Private windows can refuse storage; the toggle still works for this page.
	}
}

// Presses one of the exercise form's action buttons, so htmx handles it
// exactly like a click.
function press(view, action) {
	view.dom.closest("form")?.querySelector(`button[value="${action}"]`)?.click();
}

// :w submits, like saving in vim.
Vim.defineEx("write", "w", (cm) => press(cm.cm6, "submit"));

const editors = new Set();

function enhance(textarea) {
	if (textarea.dataset.enhanced) return;
	textarea.dataset.enhanced = "true";

	const vimMode = new Compartment();
	const view = new EditorView({
		parent: textarea.parentElement,
		state: EditorState.create({
			doc: textarea.value,
			extensions: [
				vimMode.of(vimEnabled() ? vim({ status: true }) : []), // must precede other keymaps
				lineNumbers(),
				highlightActiveLineGutter(),
				highlightActiveLine(),
				drawSelection(),
				history(),
				bracketMatching(),
				indentOnInput(),
				indentUnit.of("\t"),
				EditorState.tabSize.of(4),
				StreamLanguage.define(go),
				syntaxHighlighting(classHighlighter),
				keymap.of([
					{ key: "Mod-Enter", run: (v) => (press(v, "run"), true) },
					{ key: "Shift-Mod-Enter", run: (v) => (press(v, "submit"), true) },
					indentWithTab,
					...defaultKeymap,
					...historyKeymap,
				]),
				EditorView.updateListener.of((u) => {
					if (u.docChanged) textarea.value = u.state.doc.toString();
				}),
				EditorView.contentAttributes.of({ "aria-label": "Your code", spellcheck: "false", autocapitalize: "off" }),
			],
		}),
	});
	textarea.after(view.dom);
	textarea.hidden = true;

	const editor = { view, vimMode };
	editors.add(editor);
	// Forget editors whose DOM htmx swapped out (e.g. Reset code).
	for (const e of editors) {
		if (!e.view.dom.isConnected) {
			e.view.destroy();
			editors.delete(e);
		}
	}

	const form = textarea.closest("form");
	const toggle = form?.querySelector(".vim-toggle");
	const box = toggle?.querySelector("input");
	if (box) {
		toggle.hidden = false;
		box.checked = vimEnabled();
		box.addEventListener("change", () => {
			rememberVim(box.checked);
			view.dispatch({ effects: vimMode.reconfigure(box.checked ? vim({ status: true }) : []) });
			view.focus();
		});
	}
	const hint = form?.querySelector(".hint");
	if (hint) {
		hint.textContent = "Ctrl/⌘+Enter runs, Shift+Ctrl/⌘+Enter submits (:w in vim mode). Esc then Tab leaves the editor.";
	}
}

function enhanceAll(root) {
	root.querySelectorAll("textarea[data-editor]").forEach(enhance);
}

document.addEventListener("DOMContentLoaded", () => enhanceAll(document));
// htmx swaps in a fresh editor when the code is reset.
document.addEventListener("htmx:load", (e) => enhanceAll(e.detail.elt));
