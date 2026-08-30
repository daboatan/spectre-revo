import {Compartment, EditorState} from "@codemirror/state";
import {
	EditorView,
	crosshairCursor,
	drawSelection,
	dropCursor,
	highlightActiveLine,
	highlightActiveLineGutter,
	keymap,
	lineNumbers,
	rectangularSelection,
} from "@codemirror/view";
import {defaultKeymap, history, historyKeymap, indentWithTab} from "@codemirror/commands";
import {highlightSelectionMatches, searchKeymap} from "@codemirror/search";
import {HighlightStyle, bracketMatching, foldGutter, foldKeymap, indentOnInput, syntaxHighlighting} from "@codemirror/language";
import {tags} from "@lezer/highlight";
const syntaxHighlightLimit = 512 * 1024;

const languageLoaders = {
	html: () => import("@codemirror/lang-html").then(module => module.html()),
	css: () => import("@codemirror/lang-css").then(module => module.css()),
	javascript: options => import("@codemirror/lang-javascript").then(module => module.javascript(options)),
	json: () => import("@codemirror/lang-json").then(module => module.json()),
	markdown: () => import("@codemirror/lang-markdown").then(module => module.markdown()),
	python: () => import("@codemirror/lang-python").then(module => module.python()),
	sql: () => import("@codemirror/lang-sql").then(module => module.sql()),
	cpp: () => import("@codemirror/lang-cpp").then(module => module.cpp()),
	java: () => import("@codemirror/lang-java").then(module => module.java()),
	rust: () => import("@codemirror/lang-rust").then(module => module.rust()),
	php: () => import("@codemirror/lang-php").then(module => module.php()),
	go: () => import("@codemirror/lang-go").then(module => module.go()),
};

const brightHighlightStyle = HighlightStyle.define([
	{tag: tags.comment, color: "#7f8c98", fontStyle: "italic"},
	{tag: [tags.keyword, tags.operatorKeyword, tags.controlKeyword], color: "#ff79c6"},
	{tag: [tags.name, tags.deleted, tags.character, tags.propertyName], color: "#f8f8f2"},
	{tag: [tags.function(tags.variableName), tags.labelName], color: "#50fae6"},
	{tag: [tags.color, tags.constant(tags.name), tags.standard(tags.name)], color: "#bd93f9"},
	{tag: [tags.definition(tags.name), tags.separator], color: "#8be9fd"},
	{tag: [tags.typeName, tags.className, tags.number, tags.changed, tags.annotation, tags.modifier, tags.self, tags.namespace], color: "#ffb86c"},
	{tag: [tags.string, tags.special(tags.string), tags.inserted], color: "#a8ff60"},
	{tag: [tags.regexp, tags.escape, tags.link], color: "#ff6e67"},
	{tag: [tags.meta, tags.documentMeta], color: "#f1fa8c"},
	{tag: tags.invalid, color: "#ffffff", backgroundColor: "#ff5555"},
]);

const spectreTheme = EditorView.theme({
	"&": {height: "100%", backgroundColor: "#282828", color: "#f8f8f2"},
	"&.cm-focused": {outline: "none"},
	".cm-scroller": {overflow: "auto", fontFamily: "EnvyCodeRWeb, monospace"},
	".cm-content": {caretColor: "#ffffff", padding: "8px 0"},
	".cm-cursor, .cm-dropCursor": {borderLeftColor: "#ffffff", borderLeftWidth: "2px"},
	".cm-selectionBackground, &.cm-focused .cm-selectionBackground, ::selection": {backgroundColor: "#4b5675 !important"},
	".cm-gutters": {backgroundColor: "#242424", color: "#777777", borderRight: "1px solid #444444"},
	".cm-activeLineGutter": {backgroundColor: "#343434", color: "#dddddd"},
	".cm-activeLine": {backgroundColor: "#303030"},
	".cm-foldPlaceholder": {backgroundColor: "#3a3a3a", border: "none", color: "#dddddd"},
}, {dark: true});

function selectedLanguage(source) {
	const selector = document.querySelector("#langbox");
	return (selector && selector.value) || source.dataset.language || "text";
}

async function languageExtension(language, documentLength, status) {
	if(documentLength > syntaxHighlightLimit) {
		status.textContent = "Large paste: syntax highlighting disabled";
		return [];
	}
	const name = String(language || "").toLowerCase();
	const aliases = {
		text: "html", html5: "html", less: "css", scss: "css", js: "javascript",
		md: "markdown", python3: "python", py: "python", py3: "python",
		c: "cpp", "c++": "cpp",
	};
	const resolved = aliases[name] || name;
	let options;
	if(name === "jsx") options = {jsx: true};
	if(name === "ts" || name === "typescript") options = {typescript: true};
	if(name === "tsx") options = {jsx: true, typescript: true};
	const loader = languageLoaders[resolved === "jsx" || resolved === "ts" || resolved === "typescript" || resolved === "tsx" ? "javascript" : resolved];
	if(!loader) {
		status.textContent = "";
		return [];
	}
	status.textContent = "Loading syntax highlighting…";
	try {
		const extension = await loader(options);
		status.textContent = "";
		return extension;
	} catch(error) {
		console.error("Failed to load CodeMirror language", error);
		status.textContent = "Syntax highlighting unavailable";
		return [];
	}
}

async function initializeEditor() {
	const source = document.querySelector("#code-editor-source");
	const parent = document.querySelector("#code-editor");
	const form = document.querySelector("#pasteForm");
	const status = document.querySelector("#editor-status");
	if(!source || !parent || !form || !status) return;

	const language = new Compartment();
	let setLanguage = async () => {};
	const submitKey = {
		key: "Mod-s",
		preventDefault: true,
		run: () => {
			if(form.requestSubmit) form.requestSubmit();
			else form.submit();
			return true;
		},
	};

	const initialLanguage = await languageExtension(selectedLanguage(source), source.value.length, status);
	const state = EditorState.create({
		doc: source.value,
		extensions: [
			lineNumbers(),
			highlightActiveLineGutter(),
			history(),
			foldGutter(),
			drawSelection(),
			dropCursor(),
			EditorState.allowMultipleSelections.of(true),
			indentOnInput(),
			bracketMatching(),
			rectangularSelection(),
			crosshairCursor(),
			highlightActiveLine(),
			highlightSelectionMatches(),
			keymap.of([submitKey, indentWithTab, ...defaultKeymap, ...searchKeymap, ...historyKeymap, ...foldKeymap]),
			syntaxHighlighting(brightHighlightStyle),
			spectreTheme,
			language.of(initialLanguage),
			EditorView.updateListener.of(update => {
				if(update.docChanged) {
					source.dispatchEvent(new CustomEvent("spectre:editor-change", {bubbles: true}));
					const crossedHighlightLimit = (update.startState.doc.length > syntaxHighlightLimit) !== (update.state.doc.length > syntaxHighlightLimit);
					if(crossedHighlightLimit) setLanguage(selectedLanguage(source));
				}
			}),
		],
	});

	const view = new EditorView({state, parent});
	const syncSource = () => { source.value = view.state.doc.toString(); };
	let languageRequest = 0;
	setLanguage = async name => {
		const request = ++languageRequest;
		const extension = await languageExtension(name, view.state.doc.length, status);
		if(request === languageRequest) view.dispatch({effects: language.reconfigure(extension)});
	};

	form.addEventListener("submit", syncSource, true);
	const languageSelector = document.querySelector("#langbox");
	if(languageSelector) languageSelector.addEventListener("change", () => setLanguage(languageSelector.value));

	window.SpectreCodeEditor = {
		focus: () => view.focus(),
		setLanguage,
		sync: syncSource,
		value: () => view.state.doc.toString(),
		view,
	};
	document.body.classList.add("codemirror-active");
	document.dispatchEvent(new CustomEvent("spectre:editor-ready"));
	view.focus();
}

if(document.readyState === "loading") {
	document.addEventListener("DOMContentLoaded", initializeEditor, {once: true});
} else {
	initializeEditor();
}
