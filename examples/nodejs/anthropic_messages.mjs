/**
 * BOB Gemini Free - Node.js Anthropic SDK Example
 * Run BOB Gemini Free gateway: ./bob-gemini-free --port 8081
 * Then run: npm install @anthropic-ai/sdk && node anthropic_messages.mjs
 */

import Anthropic from '@anthropic-ai/sdk';

const anthropic = new Anthropic({
  baseURL: 'http://127.0.0.1:8081',
  apiKey: 'none',
});

async function main() {
  console.log('--- 1. Anthropic Messages API ---');
  const message = await anthropic.messages.create({
    model: 'claude-3-7-sonnet',
    max_tokens: 1024,
    messages: [{ role: 'user', content: 'What is the speed of light in vacuum?' }],
  });
  for (const block of message.content) {
    if (block.type === 'text') {
      console.log(block.text);
    }
  }

  console.log('\n--- 2. Real-Time Streaming ---');
  const stream = await anthropic.messages.stream({
    model: 'claude-3-7-sonnet',
    max_tokens: 1024,
    messages: [{ role: 'user', content: 'List 3 practical uses of WebSockets.' }],
  });

  stream.on('text', (text) => {
    process.stdout.write(text);
  });

  await stream.finalMessage();
  console.log('\n\nDone!');
}

main().catch(console.error);
