import * as pty from "node-pty";
import xtermHeadless from "@xterm/headless";
const { Terminal } = xtermHeadless;

const ptyProcess = pty.spawn("ls", ["-la", "--color=always"], {
  name: "xterm-color",
  cols: 80,
  rows: 24,
  cwd: process.cwd(),
  env: process.env,
});

const term = new Terminal({
  allowProposedApi: true,
  cols: 80,
  rows: 24,
});

console.log("Writing 'Terminal Started' to the invisible screen...");

ptyProcess.onData((data) => {
  term.write(data);
});

setTimeout(() => {
    const buffer = term.buffer.active;
    console.log("\n[SCREEN SNAPSHOT]:");
  
  // Let's print the first 5 lines to prove it captured the file list
  for(let i = 0; i < 5; i++) {
    const line = buffer.getLine(i);
    if (line) {
      console.log(`"${line.translateToString(true)}"`);
    }
  }

  process.exit(0);
}, 1000);
