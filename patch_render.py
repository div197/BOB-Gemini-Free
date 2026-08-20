import re

with open('internal/server/playground.html', 'r') as f:
    code = f.read()

bad_render = """    let fallbackTimer = 0;

    function doRender() {"""

good_render = """    let fallbackTimer = 0;
    let streamRenderActive = true;

    function doRender() {
      if (!streamRenderActive) return;"""

code = code.replace(bad_render, good_render)

bad_final_render = """    // Clear throttles and do a synchronous final render
    if (fallbackTimer) { clearTimeout(fallbackTimer); fallbackTimer = 0; }
    pendingRender = false;"""

good_final_render = """    // Clear throttles and do a synchronous final render
    streamRenderActive = false;
    if (fallbackTimer) { clearTimeout(fallbackTimer); fallbackTimer = 0; }
    pendingRender = false;"""

code = code.replace(bad_final_render, good_final_render)

with open('internal/server/playground.html', 'w') as f:
    f.write(code)
