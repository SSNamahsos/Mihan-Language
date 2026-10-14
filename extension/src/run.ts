import * as fs from 'fs';
import * as path from 'path';
import * as vscode from 'vscode';
import { MihanAdapter } from './adapter';

const inputWords = /\b(khanak|input|read)\s*[( "]/;
const siteWords = /\b(serva_bgo|serve|baza_bgo|open_window)\s*[( ]/;

function sourceOf(file: string): string {
  const open = vscode.workspace.textDocuments.find((d) => d.uri.fsPath === file);
  if (open) return open.getText();
  try { return fs.readFileSync(file, 'utf8'); } catch { return ''; }
}

function folderOf(file: string): string {
  const f = vscode.workspace.getWorkspaceFolder(vscode.Uri.file(file));
  return f ? f.uri.fsPath : path.dirname(file);
}

function activeMihanFile(): string | undefined {
  const ed = vscode.window.activeTextEditor;
  if (ed && (ed.document.languageId === 'mihan' || ed.document.fileName.endsWith('.mhn'))) return ed.document.fileName;
  return undefined;
}

export function registerRun(context: vscode.ExtensionContext, command: () => string) {
  let terminal: vscode.Terminal | undefined;

  const provider: vscode.DebugConfigurationProvider = {
    provideDebugConfigurations: () => [{
      type: 'mihan', request: 'launch', name: 'Mihan: run this file', program: '${file}', cwd: '${workspaceFolder}',
    }],
    resolveDebugConfiguration: async (_folder, config) => {
      if (!config.type && !config.request && !config.name) {
        const file = activeMihanFile();
        if (!file) {
          vscode.window.showInformationMessage('Open a .mhn file to run it.');
          return undefined;
        }
        config.type = 'mihan';
        config.request = 'launch';
        config.name = 'Mihan: run this file';
        config.program = file;
      }
      if (!config.program || config.program === '${file}') {
        const file = activeMihanFile();
        if (!file) {
          vscode.window.showInformationMessage('Open a .mhn file to run it.');
          return undefined;
        }
        config.program = file;
      }
      const ed = vscode.window.activeTextEditor;
      if (ed && ed.document.isDirty && ed.document.fileName === config.program) await ed.document.save();
      return config;
    },
    resolveDebugConfigurationWithSubstitutedVariables: (_folder, config) => {
      const program = String(config.program);
      if (!config.cwd || String(config.cwd).includes('${')) config.cwd = folderOf(program);
      const text = sourceOf(program);
      if (!config.console) config.console = inputWords.test(text) ? 'integratedTerminal' : 'debugConsole';
      if (config.openBrowser === undefined) config.openBrowser = siteWords.test(text);
      return config;
    },
  };

  context.subscriptions.push(
    vscode.debug.registerDebugConfigurationProvider('mihan', provider),
    vscode.debug.registerDebugConfigurationProvider('mihan', {
      provideDebugConfigurations: () => provider.provideDebugConfigurations!(undefined),
    }, vscode.DebugConfigurationProviderTriggerKind.Dynamic),
    vscode.debug.registerDebugAdapterDescriptorFactory('mihan', {
      createDebugAdapterDescriptor: () => new vscode.DebugAdapterInlineImplementation(new MihanAdapter(command()) as any),
    }),
    vscode.window.onDidCloseTerminal((t) => { if (t === terminal) terminal = undefined; }),
    vscode.commands.registerCommand('mihan.runFile', async (uri?: vscode.Uri) => {
      const file = uri?.fsPath || activeMihanFile();
      if (!file) {
        vscode.window.showInformationMessage('Open a .mhn file to run it.');
        return;
      }
      const doc = vscode.workspace.textDocuments.find((d) => d.uri.fsPath === file);
      if (doc?.isDirty) await doc.save();
      const cwd = folderOf(file);
      if (!terminal) terminal = vscode.window.createTerminal({ name: 'Mihan', cwd });
      terminal.show(true);
      const exe = command();
      const q = (s: string) => (/[\s&()]/.test(s) ? `"${s}"` : s);
      const open = siteWords.test(sourceOf(file)) ? ' --open' : '';
      const call = process.platform === 'win32' && /^"/.test(q(exe)) ? '& ' : '';
      terminal.sendText(`${call}${q(exe)} run${open} ${q(file)}`);
    }),
    vscode.commands.registerCommand('mihan.debugFile', () => {
      const file = activeMihanFile();
      if (!file) return;
      vscode.debug.startDebugging(vscode.workspace.getWorkspaceFolder(vscode.Uri.file(file)), {
        type: 'mihan', request: 'launch', name: 'Mihan: run this file', program: file,
      });
    }),
  );
}
