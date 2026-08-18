# BOB Gemini Free — Complete Documentation Suite

Welcome to the comprehensive documentation suite for **BOB Gemini Free** (*Break Ordinary Boundaries*) by **ABCsteps** ([abcsteps.com](https://abcsteps.com/)) and **Divyanshu Singh Chouhan** ([@div197](https://github.com/div197)).

---

## 📚 Table of Contents

### 1. Getting Started
- [**Quickstart Guide**](./1-getting-started/quickstart.md): 30-second installation & execution for macOS, Linux, and Windows.
- [**Zero-Dependency Standalone Binaries**](./1-getting-started/zero-dependency-binary.md): Running with zero Go, Python, or runtime dependencies.
- [**Docker & OrbStack Deployment**](./1-getting-started/docker-and-orbstack.md): Containerization, healthchecks, and Compose setup.

### 2. Authentication & Session Routing
- [**1-Click Interactive Login Window**](./2-authentication-and-routing/1-click-login.md): Zero-friction session extraction via `--login`.
- [**Manual DevTools Cookie Extraction**](./2-authentication-and-routing/devtools-manual-paths.md): Visual extraction paths for headless servers.
- [**Multi-Account Cookie Pool & Auto-Rotation**](./2-authentication-and-routing/multi-account-pool.md): High-concurrency round-robin load-balancing and 429 failover.
- [**Multi-Account Profile Handling (`auth_user`)**](./2-authentication-and-routing/auth-user-profiles.md): Routing to Work vs Personal Google profiles.

### 3. AI Client Integrations
- [**Connecting Cursor, Windsurf, & Continue.dev**](./3-ai-client-integrations/cursor-and-windsurf.md): Modern AI coding IDE setup.
- [**Claude Code CLI (Anthropic Protocol)**](./3-ai-client-integrations/claude-code-cli.md): Anthropic Messages SSE streaming & reasoning blocks.
- [**OpenAI Codex CLI (Responses API)**](./3-ai-client-integrations/openai-codex-cli.md): Terminal coding agent integration.
- [**Deep Agentic Tools & Open-Source Agent Frameworks**](./3-ai-client-integrations/grok-build-and-custom-agents.md): Grok Build, OpenHands, Roo Code, Cline, Aider, and Goose.

### 4. API Reference
- [**OpenAI Standard Endpoints**](./4-api-reference/openai-endpoints.md): `/v1/chat/completions`, `/v1/models`, reasoning effort.
- [**Anthropic Messages Protocol**](./4-api-reference/anthropic-endpoints.md): `/v1/messages`, thinking blocks lifecycle.
- [**Google Native Gemini v1beta**](./4-api-reference/google-v1beta-endpoints.md): `/v1beta/models`, `/v1beta/models/{model}:generateContent`.
- [**Image Generation: Imagen 3 & Gemini Nano Banana**](./4-api-reference/imagen-3-and-nano-banana.md): `/v1/images/generations`, photorealistic & native synthesis.
- [**Health, Diagnostics & Benchmarking**](./4-api-reference/health-and-diagnostics.md): Health monitoring, 13-point diagnostic kit, stress benchmarks.

### 5. Embedded Go SDK
- [**Embedded Go Library Guide (`pkg/gateway`)**](./5-embedded-sdk/go-library-guide.md): Embedding the gateway directly inside Go microservices.

---

## 🙏 Acknowledgements & Research Foundations

BOB Gemini Free stands on the collective wisdom and engineering breakthroughs of the global AI and open-source communities:

1. **Google Research & DeepMind**: For publishing the foundational Transformer architecture (*"Attention Is All You Need"*, Vaswani et al., 2017) and for engineering the state-of-the-art Gemini 3.7 Flash, Flash Thinking, 3.1 Pro, and Imagen 3 models with generous public web accessibility.
2. **OpenAI & Anthropic**: For establishing the open API standards, Messages schemas, reasoning block conventions, and coding agent CLI patterns that unite modern developer workflows.
3. **The Go Language Team & Chromium Engineers**: For the systems-level foundations (Go standard library concurrency, zero-dependency static compilation, and Chrome DevTools Protocol) enabling high-performance, local-first, zero-friction execution.
4. **The Global Open-Source Community**: The creators and maintainers of OpenAI Codex CLI, Claude Code CLI, Cursor, Grok Build, OpenHands, Roo Code, Cline, Aider, and the global indie hacker ecosystem pushing the frontiers of software engineering.
5. **ABCsteps Technologies (Jodhpur, Rajasthan)**: For championing truthful, first-principles AI engineering education, open learning foundations, and the **Break Ordinary Boundaries (BOB)** developer empowerment mission.
