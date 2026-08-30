import {build} from "esbuild";
import {rm} from "node:fs/promises";

await rm("public/js/codemirror", {force: true, recursive: true});

await build({
	entryPoints: ["public/js/editor.js"],
	outdir: "public/js/codemirror",
	bundle: true,
	chunkNames: "chunks/[name]-[hash]",
	entryNames: "editor",
	format: "esm",
	minify: true,
	splitting: true,
	target: "es2020",
});
