import re

with open('internal/server/playground.html', 'r') as f:
    code = f.read()

# Remove the existing try {
code = code.replace("  try {\n    const baseUrl = getGatewayBaseUrl();", "    const baseUrl = getGatewayBaseUrl();")

# Add try { right after isGenerating = true;
code = code.replace("  isGenerating = true;", "  isGenerating = true;\n  try {")

with open('internal/server/playground.html', 'w') as f:
    f.write(code)
