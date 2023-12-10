import * as vscode from "vscode";

export function activate(context: vscode.ExtensionContext) {
  vscode.window.showInformationMessage("Hello World from go-test-name-finder!");

  let disposable = vscode.commands.registerCommand("extension.findTest", () => {
    // TODO: 選択された位置の行数、カラム数(開始、終了)、ファイル名を取得する
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
      return;
    }

    const document = editor.document;
    const selection = editor.selection;
    const filePath = document.fileName;

    vscode.window.showInformationMessage(filePath);
  });

  context.subscriptions.push(disposable);
}

export function deactivate() {}
