import * as pty from "node-pty";
import * as os from "node:os";
import xtermHeadles from "@xterm/headless";
const { Terminal } = xtermHeadles;

type TerminalInstance = import("@xterm/headless").Terminal;

export function createSession() {
  const shell = os.platform() === "win32" ? "powershell.exe" : "bash";

  const ptyProcess = pty.spawn(shell, [], {
    name: "xterm-color",
    cols: 80,
    rows: 24,
    cwd: process.cwd(),
    env: process.env,
  });

  const term: TerminalInstance = new Terminal({
    allowProposedApi: true,
    cols: 80,
    rows: 24,
  });

  ptyProcess.onData((data) => {
    term.write(data);
  });

  return {
    write: (input: string) => {
      ptyProcess.write(input);
    },

    wait: (ms: number) => new Promise((resolve) => setTimeout(resolve, ms)),

    getSnapshot: (lines: number) => {
      const buffer = term.buffer.active;
      const snapshot: string[] = [];

      for (let i = 0; i < lines; i++) {
        const line = buffer.getLine(i);
        if (line) {
          snapshot.push(line.translateToString(true));
        }
      }
      return snapshot;
    },

    kill: () => {
      ptyProcess.kill();
    },
  };
}

if (process.argv[1] === import.meta.filename) {
  (async () => {
    const session = createSession();

    session.write("ls -la --color=always\r");

    await session.wait(1000);

    const lines = session.getSnapshot(5);
    lines?.forEach((line) => console.log(`> "${line}"`));

    session.kill();
  })();
}
