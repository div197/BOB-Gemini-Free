import re

with open('internal/server/playground.html', 'r') as f:
    code = f.read()

bad_logic = """        const readPromise = reader.read();
        const timeoutPromise = new Promise((_, reject) => setTimeout(() => reject(new Error('Stream timeout')), 15000));
        const {done, value} = await Promise.race([readPromise, timeoutPromise]);
        if (done) break;"""

good_logic = """        let timer;
        const readPromise = reader.read();
        const timeoutPromise = new Promise((_, reject) => {
          timer = setTimeout(() => reject(new Error('Stream timeout')), 15000);
        });
        const {done, value} = await Promise.race([readPromise, timeoutPromise]);
        clearTimeout(timer);
        if (done) break;"""

code = code.replace(bad_logic, good_logic)

with open('internal/server/playground.html', 'w') as f:
    f.write(code)
