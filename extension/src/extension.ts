

import * as fs from 'fs';
import * as path from 'path';
import * as vscode from 'vscode';
import { spawn, ChildProcessWithoutNullStreams } from 'child_process';

let trace: vscode.OutputChannel;

export function activate(context: vscode.ExtensionContext) {
  trace = vscode.window.createOutputChannel('Mihan');
  context.subscriptions.push(trace);

  const client = new MihanClient(context, trace);
  context.subscriptions.push(client);

  context.subscriptions.push(
    vscode.commands.registerCommand('mihan.restartLanguageServer', () => client.restart()),
    vscode.commands.registerCommand('mihan.showOutput', () => trace.show(true)),
    vscode.commands.registerCommand('mihan.newProgram', () => newProgram()),
  );

}

class MihanClient {
  private proc: ChildProcessWithoutNullStreams | null = null;
  private buffer = Buffer.alloc(0);
  private nextId = 1;
  private pending = new Map<number, (value: any) => void>();
  private disposables: vscode.Disposable[] = [];
  private providers: vscode.Disposable[] = [];
  private status: vscode.StatusBarItem;
  private starting: Promise<void> | null = null;
  private restarts = 0;

  constructor(private context: vscode.ExtensionContext, private trace: vscode.OutputChannel) {
    this.status = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Right, 100);
    this.status.text = '$(beaker) Mihan';
    this.status.tooltip = 'The Mihan language server';
    this.status.show();
    this.disposables.push(this.status);
    this.disposables.push(
      vscode.workspace.onDidOpenTextDocument((doc) => this.opened(doc)),
      vscode.workspace.onDidChangeTextDocument((e) => this.changed(e.document)),
      vscode.workspace.onDidCloseTextDocument((doc) => {
        if (doc.languageId !== 'mihan' || !this.proc) return;
        this.notify('textDocument/didClose', { textDocument: { uri: doc.uri.toString() } });
      }),
    );

    this.restart();
  }

  dispose() {
    for (const d of this.disposables) d.dispose();
    this.stop();
  }

  restart() {
    if (this.starting) return this.starting;
    this.starting = this.start().finally(() => {
      this.starting = null;
    });
    return this.starting;
  }

  private async start(): Promise<void> {
    this.stop();
    this.restarts++;
    this.status.text = '$(sync~spin) Mihan';
    this.status.tooltip = 'Starting the Mihan language server';

    const command = mihanCommand(this.context);
    this.write(`starting: ${command} lsp`);

    let child;
    try {
      child = spawn(command, ['lsp'], { stdio: ['pipe', 'pipe', 'pipe'] });
    } catch (e) {
      this.fail(command, `${e}`);
      return;
    }
    this.proc = child;
    child.stdout.on('data', (chunk: Buffer) => this.receive(chunk));
    child.stderr.on('data', (chunk: Buffer) => {
      this.write(`mihan-ls said: ${chunk.toString().trimEnd()}`);
    });
    child.on('error', (e: Error) => {
      this.fail(command, e.message);
    });
    child.on('exit', (code: number | null, signal: string | null) => {
      this.proc = null;

      if (code !== 0 && code !== null) {
        this.write(`mihan-ls stopped with code ${code}${signal ? ` (${signal})` : ''}`);
        this.status.text = '$(warning) Mihan';
        this.status.tooltip = 'The Mihan language server stopped. "Mihan: Show the language server\'s messages" says why.';
        vscode.window.showWarningMessage(
          'The Mihan language server stopped. Run "Mihan: Show the language server\'s messages" to see why.',
        );
      } else {
        this.status.text = '$(beaker) Mihan';
      }
    });

    const root = vscode.workspace.workspaceFolders?.[0]?.uri.toString();
    const result = await this.request('initialize', {
      processId: process.pid,
      rootUri: root,
      capabilities: {},
      workspaceFolders: null,
    });
    this.write(`initialize answered: ${JSON.stringify(result)}`);
    this.notify('initialized', {});

    this.status.text = '$(beaker) Mihan';
    this.status.tooltip = 'The Mihan language server is ready';

    for (const doc of vscode.workspace.textDocuments) {
      this.opened(doc);
    }

    this.register();
  }

  private opened(doc: vscode.TextDocument) {
    if (doc.languageId !== 'mihan' || !this.proc) return;
    this.notify('textDocument/didOpen', {
      textDocument: {
        uri: doc.uri.toString(),
        languageId: 'mihan',
        version: doc.version,
        text: doc.getText(),
      },
    });
  }

  private changed(doc: vscode.TextDocument) {
    if (doc.languageId !== 'mihan' || !this.proc) return;
    this.notify('textDocument/didChange', {
      textDocument: { uri: doc.uri.toString() },
      contentChanges: [{ text: doc.getText() }],
    });
  }

  private fail(command: string, why: string) {
    this.write(`cannot start ${command}: ${why}`);
    this.status.text = '$(error) Mihan';
    this.status.tooltip = 'The Mihan language server did not start';
    vscode.window
      .showErrorMessage(
        `Mihan could not start its language server using "${command}".`,
        'Show messages',
        'Open settings',
      )
      .then((choice) => {
        if (choice === 'Show messages') this.trace.show(true);
        if (choice === 'Open settings') {
          vscode.commands.executeCommand('workbench.action.openSettings', 'mihan.path');
        }
      });
  }

  private stop() {
    if (!this.proc) return;
    const p = this.proc;
    this.proc = null;
    try {
      p.kill();
    } catch {

    }
  }

  private register() {
    this.providers.forEach((d) => d.dispose());
    this.providers = [];
    const add = (d: vscode.Disposable) => {
      this.providers.push(d);
      this.disposables.push(d);
    };

    add(
      vscode.languages.registerCompletionItemProvider('mihan', {
        provideCompletionItems: async (document, position) => {
          this.changed(document);
          const r = await this.request('textDocument/completion', {
            textDocument: { uri: document.uri.toString() },
            position: toLspPosition(position),
          });
          return toCompletions(r);
        },
      }, '.', '(', ',', ' '));

    add(
      vscode.languages.registerDocumentFormattingEditProvider('mihan', {
        provideDocumentFormattingEdits: async (document) => {
          this.changed(document);
          const r = await this.request('textDocument/formatting', {
            textDocument: {
              uri: document.uri.toString(),
              languageId: 'mihan',
              version: document.version,
              text: document.getText(),
            },
            options: { tabSize: 4, insertSpaces: true },
          });
          if (!r || r.length === 0) return [];
          return r.map((e: any) => vscode.TextEdit.replace(toRange(e.range), e.newText));
        },
      }),
    );

    add(vscode.languages.registerHoverProvider('mihan', {
        provideHover: async (document, position) => {
          this.changed(document);
          const r = await this.request('textDocument/hover', {
            textDocument: { uri: document.uri.toString() },
            position: toLspPosition(position),
          });
          if (!r) return null;
          const value = typeof r.contents === 'object' ? r.contents.value : r.contents;
          return new vscode.Hover(new vscode.MarkdownString(value ?? ''));
        },
      }),
    );

    add(
      vscode.languages.registerSignatureHelpProvider('mihan', {
        provideSignatureHelp: async (document, position) => {
          this.changed(document);
          const r = await this.request('textDocument/signatureHelp', {
            textDocument: { uri: document.uri.toString() },
            position: toLspPosition(position),
          });
          if (!r || !r.signatures || r.signatures.length === 0) return null;
          const help = new vscode.SignatureHelp();
          help.signatures = r.signatures.map((sig: any) => {
            const info = new vscode.SignatureInformation(
              sig.label,
              sig.documentation ? new vscode.MarkdownString(sig.documentation.value) : undefined,
            );
            info.parameters = (sig.parameters || []).map((p: any) => new vscode.ParameterInformation(p.label));
            return info;
          });
          help.activeSignature = r.activeSignature || 0;
          help.activeParameter = r.activeParameter || 0;
          return help;
        },
      }, { triggerCharacters: ['(', ','], retriggerCharacters: [','] }),
    );

    add(vscode.languages.registerDefinitionProvider('mihan', {
      provideDefinition: async (document, position) => {
        this.changed(document);
        const r = await this.request('textDocument/definition', {
          textDocument: { uri: document.uri.toString() },
          position: toLspPosition(position),
        });
        if (!r || r.length === 0) return null;
        return r.map((l: any) => new vscode.Location(vscode.Uri.parse(l.uri), toRange(l.range)));
      },
    }));

    add(
      vscode.languages.registerDocumentSymbolProvider('mihan', {
        provideDocumentSymbols: async (document) => {
          this.changed(document);
          const r = await this.request('textDocument/documentSymbol', {
            textDocument: { uri: document.uri.toString() },
          });
          if (!r || r.length === 0) return [];
          return r.map((s: any) => toSymbol(s, document.uri));
        },
      }),
    );
  }

  private receive(chunk: Buffer) {
    this.buffer = Buffer.concat([this.buffer, chunk]);
    for (;;) {
      const headerEnd = this.buffer.indexOf('\r\n\r\n');
      if (headerEnd < 0) return;
      const header = this.buffer.subarray(0, headerEnd).toString('ascii');
      const match = /Content-Length:\s*(\d+)/i.exec(header);
      if (!match) {
        this.write('a message arrived with no length; dropping what was buffered');
        this.buffer = Buffer.alloc(0);
        return;
      }
      const length = parseInt(match[1], 10);
      const start = headerEnd + 4;
      if (this.buffer.length < start + length) return;
      const body = this.buffer.subarray(start, start + length).toString('utf8');
      this.buffer = this.buffer.subarray(start + length);
      this.dispatch(JSON.parse(body));
    }
  }

  private dispatch(message: any) {
    if (typeof message.id === 'number') {
      const resolve = this.pending.get(message.id);
      if (resolve) {
        this.pending.delete(message.id);
        resolve(message.result);
      }
      return;
    }
    if (message.method === 'textDocument/publishDiagnostics') {
      const p = message.params;
      const diags = (p.diagnostics ?? []).map((d: any) => {
        const severity =
          d.severity === 1
            ? vscode.DiagnosticSeverity.Error
            : d.severity === 2
              ? vscode.DiagnosticSeverity.Warning
              : vscode.DiagnosticSeverity.Information;
        const problem = new vscode.Diagnostic(toRange(d.range), d.message, severity);
        problem.source = d.source || 'mihan';
        if (d.code) problem.code = d.code;
        return problem;
      });
      this.problems.set(p.uri, diags);
      const collection = vscode.languages.createDiagnosticCollection('mihan');
      collection.set(vscode.Uri.parse(p.uri), diags);
      this.diagnosticCollection?.dispose();
      this.diagnosticCollection = collection;
    }
  }

  private problems = new Map<string, readonly vscode.Diagnostic[]>();
  private diagnosticCollection: vscode.DiagnosticCollection | null = null;

  private request(method: string, params: any): Promise<any> {
    const id = this.nextId++;
    const message = JSON.stringify({ jsonrpc: '2.0', id, method, params });
    this.write(`-> ${method} ${JSON.stringify(params)}`);
    return new Promise((resolve) => {

      const timer = setTimeout(() => {
        if (this.pending.delete(id)) resolve(null);
      }, 5000);
      this.pending.set(id, (value: any) => {
        clearTimeout(timer);
        resolve(value);
      });
      if (!this.proc) {
        this.pending.delete(id);
        clearTimeout(timer);
        resolve(null);
        return;
      }
      const frame = `Content-Length: ${Buffer.byteLength(message, 'utf8')}\r\n\r\n${message}`;
      this.proc.stdin.write(frame);
    });
  }

  private notify(method: string, params: any) {
    const message = JSON.stringify({ jsonrpc: '2.0', method, params });
    this.write(`-> ${method} ${JSON.stringify(params)}`);
    const frame = `Content-Length: ${Buffer.byteLength(message, 'utf8')}\r\n\r\n${message}`;
    this.proc?.stdin.write(frame);
  }

  private write(line: string) {
    const setting = vscode.workspace.getConfiguration('mihan').get<string>('trace.server') || 'off';
    if (setting !== 'off') this.trace.appendLine(line);
  }
}

async function newProgram() {
  const name = await vscode.window.showInputBox({
    prompt: 'What is your program called?',
    placeHolder: 'salam',
    validateInput: (v) => (v && /^[a-zA-Z_][a-zA-Z0-9_]*$/.test(v) ? undefined : 'A name is a word, with letters and no marks.'),
  });
  if (!name) return;
  const folder = vscode.workspace.workspaceFolders?.[0];
  if (!folder) {
    vscode.window.showErrorMessage('Open a folder first, so the program has somewhere to go.');
    return;
  }
  const uri = vscode.Uri.joinPath(folder.uri, name, 'main.mhn');
  const doc = await vscode.workspace.openTextDocument(uri);
  const edit = new vscode.WorkspaceEdit();
  edit.insert(
    uri,
    new vscode.Position(0, 0),
    `# ${name} — a Mihan program.` + '\n\n' +
      `#khoroji` + '\n' +
      `#| salam, donya!` + '\n\n' +
      `chap_khat "salam, donya!"` + '\n',
  );
  await vscode.workspace.applyEdit(edit);
  await vscode.window.showTextDocument(doc);
}

function mihanCommand(context: vscode.ExtensionContext): string {
  const configured = vscode.workspace.getConfiguration('mihan').get<string>('path') || 'mihan';
  if (configured !== 'mihan') return configured;
  const exe = process.platform === 'win32' ? 'mihan.exe' : 'mihan';

  const bundled = path.join(context.extensionPath, 'server', exe);
  const built = path.join(context.extensionPath, '..', 'bin', exe);
  if (fs.existsSync(built)) return built;
  if (onPath(exe)) return 'mihan';
  if (fs.existsSync(bundled)) return bundled;
  return 'mihan';
}

function toLspPosition(p: vscode.Position) {
  return { line: p.line, character: p.character };
}

function toRange(r: any): vscode.Range {
  return new vscode.Range(r.start.line, r.start.character, r.end.line, r.end.character);
}

function toCompletions(r: any): vscode.CompletionItem[] {
  if (!r || !r.items) return [];
  return r.items.map((it: any) => {
    const item = new vscode.CompletionItem(it.label, toCompletionKind(it.kind));
    if (it.detail) item.detail = it.detail;
    if (it.filterText) item.filterText = it.filterText;
    if (it.insertTextFormat === 2 && typeof it.insertText === 'string') {

      item.insertText = new vscode.SnippetString(it.insertText);
    } else if (it.insertText) {
      item.insertText = it.insertText;
    }
    if (it.sortText) item.sortText = it.sortText;
    if (it.documentation?.value) {
      item.documentation = new vscode.MarkdownString(it.documentation.value);
    }
    return item;
  });
}

function toCompletionKind(kind: number): vscode.CompletionItemKind {
  const k = vscode.CompletionItemKind;
  switch (kind) {
    case 1: return k.Text;
    case 2: return k.Method;
    case 3: return k.Function;
    case 5: return k.Field;
    case 6: return k.Variable;
    case 7: return k.Class;
    case 14: return k.Keyword;
    case 15: return k.Snippet;
    case 21: return k.Constant;
    default: return k.Text;
  }
}

function toSymbol(s: any, uri: vscode.Uri): vscode.DocumentSymbol {
  const range = toRange(s.location.range);
  const symbol = new vscode.DocumentSymbol(
    s.name,
    '',
    toSymbolKind(s.kind),
    range,
    range,
  );
  symbol.children = [];
  return symbol;
}

function toSymbolKind(kind: number): vscode.SymbolKind {
  const k = vscode.SymbolKind;
  switch (kind) {
    case 1: return k.File;
    case 5: return k.Class;
    case 6: return k.Method;
    case 10: return k.Enum;
    case 11: return k.Interface;
    case 12: return k.Function;
    default: return k.Variable;
  }
}

function onPath(exe: string): boolean {
  const dirs = (process.env.PATH || '').split(path.delimiter);
  return dirs.some((d) => d && fs.existsSync(path.join(d, exe)));
}
