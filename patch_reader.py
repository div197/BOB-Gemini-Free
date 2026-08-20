import re

with open('internal/server/playground.html', 'r') as f:
    code = f.read()

bad_reader_exit = """      } catch (err) {
        console.warn("Stream interrupted:", err);
        break; // Break gracefully on timeout or disconnect
      }
    }
    // Flush any remaining partial buffer (safety net for truncated streams)"""

good_reader_exit = """      } catch (err) {
        console.warn("Stream interrupted:", err);
        break; // Break gracefully on timeout or disconnect
      }
    }
    try { await reader.cancel(); } catch (e) {}
    try { reader.releaseLock(); } catch (e) {}
    // Flush any remaining partial buffer (safety net for truncated streams)"""

code = code.replace(bad_reader_exit, good_reader_exit)

with open('internal/server/playground.html', 'w') as f:
    f.write(code)
