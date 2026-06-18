<p align="center">
  <img src="https://raw.githubusercontent.com/8treenet/blog/9d6f2d861fbb7c5f4627a9a5b1a3472fb4236881/img/favicon.svg" alt="Raven" width="80" />
</p>

<h1 align="center">Raven</h1>

<p align="center">open source · self-hosted · ai harness</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=white" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-6-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Vite-8-646CFF?style=flat-square&logo=vite&logoColor=white" alt="Vite" />
  <img src="https://img.shields.io/badge/Tailwind_CSS-4-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/Docker-ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/License-Apache%202.0-green?style=flat-square" alt="License" />
</p>

<p align="center">
  <a href="https://goraven.dev">Website</a> ·
  <a href="https://preview.goraven.dev">Live Preview</a> ·
  <a href="https://discord.gg/derR7CBYDW">Discord</a> ·
  <a href="./README_CN.md">中文</a>
</p>

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/enchat.png?raw=true" alt="Chat" width="90%" />
</p>

---

## Stop Building Chatbots. Start Building an AI Harness.

Raven is a self-hosted, multi-user AI Agent harness. It assembles models, tools, knowledge, files, skills, and workflows into a single runtime environment where agents can take on tasks, invoke tools, read documents, write code, and compound capabilities across sessions.

Basic agent functionality is no longer scarce. What's scarce is an engineering system that a team can actually rely on. Raven is not another chat interface. It connects model access, task orchestration, tool invocation, knowledge retrieval, file workspaces, skill distribution, and operational governance into one cohesive harness.

---

## Quick Start

```bash
docker pull 8treenet/raven:latest

docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  8treenet/raven:latest
```

### With persistent data

```bash
docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -v /opt/raven_data:/raven/data \
  8treenet/raven:latest
```

After startup, visit `http://localhost:8000` and follow the initialization flow to create your admin account.

---

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/cnd.png?raw=true" alt="Dashboard" width="90%" />
</p>

---

## Capabilities

- **Model orchestration** — Connect OpenAI, Claude, DeepSeek, Qwen, GLM, and compatible APIs. Choose models by task, allocate by cost and quality, no vendor lock-in.
- **Long-running tasks** — Main agent understands goals, sub-agents decompose and execute in parallel, system agent manages context. Complex work breaks free from single-turn constraints.
- **Skill marketplace** — Prompts, commands, scripts, and workflows packaged as reusable skills. One-click install with automatic dependency resolution and centralized versioning.
- **Shared workspace** — Per-user isolated filesystem with team-shared spaces and project collaboration areas. Agents read, write, and execute commands in context.
- **Knowledge base** — Policies, docs, and business data ingested into RAG. Agents retrieve context in real time during planning, coding, and Q&A.
- **MCP toolchain** — Connect internal APIs, databases, private services, and CLI tools via MCP. Agents don't just suggest, they query data, invoke services, and trigger actions.
- **Plugin hooks** — Lifecycle hooks inject logic at conversation start/end, tool calls, and SSE events. Customize agent behavior without forking core code.
- **Operations dashboard** — Usage metrics, model consumption, and user activity in one view. Observe how your team uses AI and continuously optimize the harness.

---

## Roadmap

- Core code cleanup and open-source
- Multi-user sandbox
- API
- Multi-user collaboration

---

## License

Apache-2.0
