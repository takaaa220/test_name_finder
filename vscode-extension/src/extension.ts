import * as vscode from "vscode";
import { execSync } from "child_process";
import { join } from "path";

export function activate(context: vscode.ExtensionContext) {
  let disposable = vscode.commands.registerCommand(
    "go-test-executor.exec",
    async () => {
      const editor = vscode.window.activeTextEditor;
      if (!editor) {
        return;
      }

      const document = editor.document;
      const filePath = document.fileName;
      const selection = editor.selection;

      const test = await getTest(context, filePath, selection).catch(() => {
        // TODO: output error on development environment
        // console.error(error)
        return undefined;
      });

      if (!test) {
        vscode.window.showErrorMessage("no tests found");
        return;
      }

      const { functionName, subTestName } = test;

      // see:
      // - https://github.com/golang/vscode-go/blob/d6fb20289a8484e57dc4fa21a2f44094de7f1a5b/src/goRunTestCodelens.ts#L134-L138
      // - https://github.com/golang/vscode-go/blob/master/src/goTest.ts#L234-L246
      if (subTestName) {
        vscode.window.showInformationMessage(
          `Start to run "${subTestName}" in "${functionName}"`
        );

        vscode.commands.executeCommand("go.subtest.cursor", {
          functionName,
          subTestName,
        });

        return;
      }

      vscode.window.showInformationMessage(
        `Start to run tests in "${functionName}"`
      );
      vscode.commands.executeCommand("go.test.cursor", {
        functionName,
      });
    }
  );

  context.subscriptions.push(disposable);
}

async function getTest(
  context: vscode.ExtensionContext,
  filePath: string,
  selection: vscode.Selection
): Promise<{ functionName: string; subTestName?: string }> {
  return new Promise((resolve, reject) => {
    try {
      const binaryPath = join(
        context.extensionPath,
        "out/bin/test-name-finder"
      );

      // TODO: fix
      const res = execSync(
        `${binaryPath} -f ${filePath} -l ${selection.start.line + 1} -s ${
          selection.start.character
        } -e ${selection.end.character}`
      );

      const testName = res.toString();

      const [functionName, subTestName] = testName
        .split("/")
        .map((s) => s.trim());

      resolve({ functionName, subTestName });
    } catch (error) {
      reject(error);
    }
  });
}

export function deactivate() {}
