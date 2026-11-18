"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = require("vscode");
const child_process_1 = require("child_process");
let collection;
function activate(ctx) {
    collection = vscode.languages.createDiagnosticCollection("piclang");
    ctx.subscriptions.push(collection);
    ctx.subscriptions.push(vscode.workspace.onDidOpenTextDocument((d) => {
        if (d.languageId === "piclang")
            checkDocument(d);
    }), vscode.workspace.onDidSaveTextDocument((d) => {
        if (d.languageId !== "piclang")
            return;
        const cfg = vscode.workspace.getConfiguration("piclang");
        if (cfg.get("checkOnSave", true))
            checkDocument(d);
    }), vscode.workspace.onDidCloseTextDocument((d) => {
        if (d.languageId === "piclang")
            collection.delete(d.uri);
    }));
    ctx.subscriptions.push(vscode.commands.registerCommand("piclang.check", () => {
        const ed = vscode.window.activeTextEditor;
        if (ed)
            checkDocument(ed.document);
    }), vscode.commands.registerCommand("piclang.format", async () => {
        const ed = vscode.window.activeTextEditor;
        if (ed)
            await formatDocument(ed.document);
    }));
}
function checkDocument(doc) {
    const cfg = vscode.workspace.getConfiguration("piclang");
    const picf = cfg.get("picfPath", "picf");
    const file = doc.uri.fsPath;
    (0, child_process_1.execFile)(picf, ["check", "--json", file], { timeout: 15000 }, (_err, stdout, stderr) => {
        let diags = [];
        try {
            diags = JSON.parse(stdout || "[]");
        }
        catch (e) {
            const msg = (stderr || String(e)).trim() || "picf: output inválido";
            collection.set(doc.uri, [
                new vscode.Diagnostic(new vscode.Range(0, 0, 0, 1), msg, vscode.DiagnosticSeverity.Error),
            ]);
            return;
        }
        applyDiagnostics(doc, diags);
    });
}
function applyDiagnostics(doc, diags) {
    const out = diags.map((d) => {
        const line = Math.max(0, (d.line || 1) - 1);
        const col = Math.max(0, (d.col || 1) - 1);
        const endLine = d.endLine ? d.endLine - 1 : line;
        const endCol = d.endCol ? d.endCol - 1 : col + 1;
        const range = new vscode.Range(line, col, endLine, Math.max(endCol, col + 1));
        const sev = d.severity === "warning"
            ? vscode.DiagnosticSeverity.Warning
            : vscode.DiagnosticSeverity.Error;
        const diag = new vscode.Diagnostic(range, d.message, sev);
        diag.source = "picf";
        return diag;
    });
    collection.set(doc.uri, out);
}
function formatDocument(doc) {
    const cfg = vscode.workspace.getConfiguration("piclang");
    const picf = cfg.get("picfPath", "picf");
    const file = doc.uri.fsPath;
    return new Promise((resolve) => {
        (0, child_process_1.execFile)(picf, ["fmt", file], { timeout: 15000 }, async (err, stdout, stderr) => {
            if (err || !stdout) {
                vscode.window.showErrorMessage(`picf fmt: ${stderr || err}`);
                return resolve();
            }
            const edit = new vscode.WorkspaceEdit();
            const full = new vscode.Range(doc.positionAt(0), doc.positionAt(doc.getText().length));
            edit.replace(doc.uri, full, stdout);
            await vscode.workspace.applyEdit(edit);
            resolve();
        });
    });
}
function deactivate() {
    if (collection)
        collection.dispose();
}
//# sourceMappingURL=extension.js.map