/**
 * BOB Gemini Free - Node.js OpenAI SDK Example
 * Run BOB Gemini Free gateway: ./bob-gemini-free --port 9610
 * Then run: npm install openai && node openai_chat.mjs
 */

import OpenAI from 'openai';

const openai = new OpenAI({
  baseURL: 'http://127.0.0.1:9610/v1',
  apiKey: 'none',
});

async function main() {
  console.log('--- 1. Synchronous Chat Completion ---');
  const completion = await openai.chat.completions.create({
    model: 'gemini-3.7-flash',
    messages: [{ role: 'user', content: 'Explain event loops in Node.js in 2 sentences.' }],
  });
  console.log(completion.choices[0].message.content);

  console.log('\n--- 2. Real-Time Streaming ---');
  const stream = await openai.chat.completions.create({
    model: 'gemini-3.7-flash-thinking',
    messages: [{ role: 'user', content: 'Count from 1 to 5 with a short explanation for each.' }],
    stream: true,
  });

  for await (const chunk of stream) {
    const delta = chunk.choices[0]?.delta;
    if (delta?.reasoning_content) {
      process.stdout.write(`[Thinking]: ${delta.reasoning_content}`);
    }
    if (delta?.content) {
      process.stdout.write(delta.content);
    }
  }
  console.log('\n\nDone!');
}

main().catch(console.error);
