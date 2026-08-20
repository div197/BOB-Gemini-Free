import re

for filepath in ['functions/v1/chat/completions.js', 'web/functions/v1/chat/completions.js']:
    with open(filepath, 'r') as f:
        code = f.read()

    # Fix isGoogleEnd
    bad_google_end = """            const chunk = decoder.decode(value, { stream: true });
            const isGoogleEnd = chunk.includes('["e",') || chunk.includes('["di",');

            const deltas = parser.feed(chunk);"""
            
    good_google_end = """            const chunk = decoder.decode(value, { stream: true });
            
            // Fix BardErrorInfo silent dropping
            if (chunk.includes("BardErrorInfo")) {
               throw new Error("Upstream Error: BardErrorInfo (Auth/RateLimit)");
            }

            const deltas = parser.feed(chunk);
            // Properly check for Google end marker to avoid false positives in code blocks
            const isGoogleEnd = chunk.includes('\\n[["e",') || chunk.includes('\\n[["di",') || chunk.startsWith('[["e",') || chunk.startsWith('[["di",');"""
            
    code = code.replace(bad_google_end, good_google_end)

    with open(filepath, 'w') as f:
        f.write(code)
