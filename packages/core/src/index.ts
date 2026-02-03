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
  term.write("Terminal started");
});

setTimeout(() => {
  const buffer = term.buffer.active;
  const topRow = buffer.getLine(0);

  if (topRow) {
    console.log("\n[SCREEN SNAPSHOT]:");
    // .translateToString(true) trims the empty black space
    console.log(`Row 0 says: "${topRow.translateToString(true)}"`);
  } else {
    console.log("Error: Screen is empty!");
  }
}, 1000);
