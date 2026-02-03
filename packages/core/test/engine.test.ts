import { describe, it, expect } from "vitest";
import { createSession } from '../src/index'

describe('Bannin Engine', ()=>{
    it('should capture output from the terminal session', async ()=>{
        const session = createSession();

        session.write('ls -la --color=always\r');

        await session.waitForText('package.json');

        const lines = session.getSnapshot(5);
        lines?.forEach((line) => console.log(`> "${line}"`));

        session.saveRecording('recording.json');

        expect(lines.length).toBeGreaterThan(0);

        session.kill();
    })

})