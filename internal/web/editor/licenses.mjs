// Prints the license of every package bundled into editor.js.
import { readFileSync, readdirSync, existsSync } from "node:fs";
import { join } from "node:path";

const bundled = ["@codemirror/commands", "@codemirror/language", "@codemirror/legacy-modes", "@codemirror/state",
	"@codemirror/view", "@lezer/common", "@lezer/highlight", "@marijn/find-cluster-break", "@replit/codemirror-vim",
	"crelt", "style-mod", "w3c-keyname"];
for (const name of bundled) {
	const dir = join("node_modules", name);
	if (!existsSync(dir)) continue;
	const file = readdirSync(dir).find((f) => /^licen[cs]e/i.test(f));
	const { version } = JSON.parse(readFileSync(join(dir, "package.json"), "utf8"));
	console.log(`== ${name}@${version}\n\n${file ? readFileSync(join(dir, file), "utf8").trim() : "(no license file)"}\n`);
}
