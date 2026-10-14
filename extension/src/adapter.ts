import { spawn, ChildProcess } from 'child_process';

export interface LaunchArgs {
  program: string;
  cwd?: string;
  args?: string[];
  console?: 'debugConsole' | 'integratedTerminal';
  openBrowser?: boolean;
  noDebug?: boolean;
}

type Listener = (m: any) => void;

export class MihanAdapter {
  private listeners: Listener[] = [];
  private seq = 1;
  private child: ChildProcess | null = null;
  private pid = 0;
  private poll: NodeJS.Timeout | null = null;
  private done = false;
  private pendingTerminal = new Map<number, any>();

  constructor(private command: string) {}

  onDidSendMessage = (l: Listener) => {
    this.listeners.push(l);
    return { dispose: () => { this.listeners = this.listeners.filter((x) => x !== l); } };
  };

  private send(m: any) {
    m.seq = this.seq++;
    for (const l of this.listeners) l(m);
  }

  private respond(req: any, body?: any, success = true, message?: string) {
    this.send({ type: 'response', request_seq: req.seq, command: req.command, success, body, message });
  }

  private event(event: string, body?: any) {
    this.send({ type: 'event', event, body });
  }

  private output(category: string, text: string) {
    this.event('output', { category, output: text });
  }

  commandLine(a: LaunchArgs): string[] {
    const out = ['run'];
    if (a.openBrowser) out.push('--open');
    out.push(a.program);
    return out.concat(a.args || []);
  }

  handleMessage(m: any) {
    if (m.type === 'response') {
      const req = this.pendingTerminal.get(m.request_seq);
      if (req) {
        this.pendingTerminal.delete(m.request_seq);
        if (!m.success) {
          this.output('stderr', `Could not open a terminal: ${m.message || ''}\n`);
          this.finish(1);
          return;
        }
        this.pid = (m.body && (m.body.shellProcessId || m.body.processId)) || 0;
        if (this.pid) {
          this.poll = setInterval(() => {
            try { process.kill(this.pid, 0); } catch { this.finish(0); }
          }, 700);
        }
      }
      return;
    }
    switch (m.command) {
      case 'initialize':
        this.respond(m, {
          supportsConfigurationDoneRequest: true,
          supportsTerminateRequest: true,
          supportsRestartRequest: false,
        });
        this.event('initialized');
        break;
      case 'configurationDone':
        this.respond(m);
        break;
      case 'launch':
        this.launch(m);
        break;
      case 'threads':
        this.respond(m, { threads: [{ id: 1, name: 'mihan' }] });
        break;
      case 'terminate':
      case 'disconnect':
        this.kill();
        this.respond(m);
        this.finish(null);
        break;
      default:
        this.respond(m, {});
    }
  }

  private launch(req: any) {
    const a: LaunchArgs = req.arguments || {};
    if (!a.program) {
      this.respond(req, undefined, false, 'There is no .mhn file to run.');
      this.finish(1);
      return;
    }
    const argv = this.commandLine(a);
    this.respond(req);
    this.output('console', `> mihan ${argv.join(' ')}\n`);
    if (a.console === 'integratedTerminal') {
      const seq = this.seq;
      this.pendingTerminal.set(seq, req);
      this.send({
        type: 'request',
        command: 'runInTerminal',
        arguments: { kind: 'integrated', title: 'Mihan', cwd: a.cwd || process.cwd(), args: [this.command, ...argv] },
      });
      return;
    }
    let child: ChildProcess;
    try {
      child = spawn(this.command, argv, { cwd: a.cwd || undefined, stdio: ['ignore', 'pipe', 'pipe'] });
    } catch (e) {
      this.output('stderr', `Could not start ${this.command}: ${e}\n`);
      this.finish(1);
      return;
    }
    this.child = child;
    child.stdout?.on('data', (d: Buffer) => this.output('stdout', d.toString('utf8')));
    child.stderr?.on('data', (d: Buffer) => this.output('stderr', d.toString('utf8')));
    child.on('error', (e) => {
      this.output('stderr', `Could not start ${this.command}: ${e.message}\nSet "mihan.path" to the full path of mihan.exe.\n`);
      this.finish(1);
    });
    child.on('close', (code) => this.finish(code ?? 0));
  }

  private kill() {
    const c = this.child;
    const pid = c?.pid || this.pid;
    if (!pid) return;
    if (process.platform === 'win32') {
      try { spawn('taskkill', ['/pid', String(pid), '/T', '/F']); } catch { }
    } else {
      try { c ? c.kill() : process.kill(pid); } catch { }
    }
  }

  private finish(code: number | null) {
    if (this.done) return;
    this.done = true;
    if (this.poll) clearInterval(this.poll);
    if (code !== null) {
      this.output('console', `\nmihan exited with code ${code}\n`);
      this.event('exited', { exitCode: code });
    }
    this.event('terminated');
  }

  dispose() {
    this.kill();
    if (this.poll) clearInterval(this.poll);
  }
}
