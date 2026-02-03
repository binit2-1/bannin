import * as pty from "node-pty";
import * as os from "node:os";
import * as fs from "node:fs/promises"; 
import xtermHeadles from "@xterm/headless";
import { getHeapSnapshot } from "node:v8";
import { text } from "node:stream/consumers";
const { Terminal } = xtermHeadles;

type TerminalInstance = import("@xterm/headless").Terminal;

export function createSession() {
  const shell = os.platform() === "win32" ? "powershell.exe" : "bash";

  type RecordingEvent = {
    timeOffset: number,
    content: string
  }

  const events: RecordingEvent[] = []
  const startTime = Date.now();

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
    events.push({
      timeOffset: Date.now() - startTime,
      content: data
    });
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

    saveRecording: (filepath: string) =>{
      const recordingJSON = JSON.stringify(events, null, 2)
      fs.writeFile(filepath, recordingJSON);
      console.log(`Recording Saved to ${filepath}`);
    },

    waitForText: async ( text: string, timeout = 5000 ) => {

      const searchStartTime = Date.now();
      while(Date.now() - searchStartTime < timeout){
        const snapshot = term.buffer.active;
        for(let i=0; i< snapshot.length; i++){
          const line = snapshot.getLine(i);
          if(line && line.translateToString(true).includes(text)){
            return;
          }
        }
        await new Promise((resolve) => setTimeout(resolve, 100));
      }
      throw new Error(`Timeout waiting for text: ${text}`);
    },

    kill: () => {
      ptyProcess.kill();
    },
  };
}
