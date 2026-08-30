<p align="center">
  <img src="https://raw.githubusercontent.com/8treenet/blog/9d6f2d861fbb7c5f4627a9a5b1a3472fb4236881/img/favicon.svg" alt="GoRaven" width="80" />
</p>

<h1 align="center">GoRaven</h1>

<p align="center">
  <strong>Open Source · Self-Hosted · AI Harness</strong><br/>
  Models, tools, knowledge, skills, workflows — all in one runtime. The Agent doesn't just chat. It actually gets work done.
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=white" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-6-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Vite-8-646CFF?style=flat-square&logo=vite&logoColor=white" alt="Vite" />
  <img src="https://img.shields.io/badge/Tailwind_CSS-4-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/Docker-Ready-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/License-Apache%202.0-green?style=flat-square" alt="License" />
</p>

<p align="center">
  <a href="https://goraven.dev">Website</a> ·
  <a href="https://preview.goraven.dev">Live Preview</a> ·
  <a href="https://discord.gg/derR7CBYDW">Discord</a> ·
  <a href="./README_CN.md">中文</a>
</p>

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/enchat.png?raw=true" alt="GoRaven Chat" width="90%" />
</p>

---

## What It Is

GoRaven gives every person on your team an independent Agent workspace. Not a chat window — a workstation. The Agent reads files, writes code, runs commands, calls APIs, searches your knowledge base, analyzes data, writes reports, builds charts. You assign the task; it breaks it down, works in parallel, and delivers the result.

What you get is the outcome, not a block of text you have to copy-paste.

---

## Why GoRaven

<table>
<tr>
<td width="50%">

### 🎯 Team-first

Everyone gets an isolated workspace. The team shares projects and skill libraries. Admins control model quotas, tool permissions, and data access. Not a single-player toy.

</td>
<td width="50%">

### 🔧 Actually does things

Agents directly read/write files, run Shell, call MCP tool chains. Internal APIs, databases, private services — plug them all in. Not just another suggestion paragraph.

</td>
</tr>
<tr>
<td>

### 🧩 Skill Marketplace

Prompts, scripts, workflows packaged as skills. One-click install, centralized maintenance. One person hits a wall — the whole team skips it next time.

</td>
<td>

### 📚 Knowledge in the loop

Policies, docs, business materials go into RAG. Agents retrieve context in real time during execution. Answers have sources, not hallucinations.

</td>
</tr>
<tr>
<td>

### 🔌 Plugin Hooks

Before/after conversations, tool calls, SSE event streams — inject custom logic at any lifecycle point. Deep customization without forking core code.

</td>
<td>

### 🏠 Fully self-hosted

One Docker command and it's running. Data never leaves your server. Works with OpenAI, Claude, DeepSeek, Qwen, GLM, or any compatible API.

</td>
</tr>
</table>

---

## Quick Start

```bash
docker pull 8treenet/goraven:latest

docker run -d --restart=always --name goraven-agent \
  -p 8000:8000 \
  8treenet/goraven:latest
```

### Persistent Data

```bash
docker run -d --restart=always --name goraven-agent \
  -p 8000:8000 \
  -v /opt/goraven:/goraven/data \
  8treenet/goraven:latest
```

After startup, visit `http://localhost:8000` and follow the setup wizard to create your admin account.

> The container uses UTC by default. To set a different timezone, add `-e TZ=America/New_York`. Available timezones: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

---

## ☁️ One-Click Deploy

[![Deploy on RepoCloud](https://d16t0pc4846x52.cloudfront.net/deploylobe.svg)](https://repocloud.io/details/GoRaven/)

---

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/end.png?raw=true" alt="GoRaven Dashboard" width="90%" />
</p>

---

## Architecture

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/architecture.svg?raw=true" alt="GoRaven Architecture" width="90%" />
</p>

---

## Design Principles

The model is the Way; the Agent is the instrument. The Way evolves on its own — the instrument need only be precise.

We don't do "memory and evolution" — if it can't be distilled into generalizable rules or real-time retrieval augmentation, it's just noise. The model's evolution is the model's own business; the Agent shouldn't overstep.

As a tool, the Agent has only three duties: deterministic workflows, precise tool invocation, and robust execution. Borrow the instrument to let wisdom arise naturally — a true instrument is one that fulfills its purpose to the fullest.

**Observability First**: Collaboration isn't about delegating tasks — it's about aligning understanding. Human-agent communication is far more fragile than human-human communication: you can't see what it's thinking, and it can't guess what you want. Observability of the execution process must be pursued and solved at the product level — every round of observation should become the reserve for the next round of decisions.

**OS First**: Models have knowledge and reasoning, but no eyes and no hands. For an Agent to get work done, it must heavily leverage operating system capabilities. But an Agent can't come with a full OS — so integrating runtime environments and dependencies must be the first priority. This isn't something code or algorithms alone can solve. A good Agent is, at its core, an Agent OS — except you can't actually build an operating system from scratch.

**Engineering First**: Agents are too flexible — tool combinations explode fast, and tweaking a pruning rule even slightly sends the whole execution flow off track. Without solid engineering, a few iterations in and nobody dares to touch it. The time spent on engineering isn't about shipping one more feature today. It's about making sure the project is still worth working on next year.

---

## Build & Run

Building from source is only recommended when you need to modify the code:

```bash
cd frontend && pnpm build
go run main.go
```

For non-development needs, use the Docker image instead — it bundles the full runtime (Go, Node.js, Python, Git, and more); building from source requires setting up these environments yourself.

---

## Community

- 💬 [Discord](https://discord.gg/derR7CBYDW) — feedback, discussions, tips
- 🌐 [goraven.dev](https://goraven.dev) — Website
- 🚀 [preview.goraven.dev](https://preview.goraven.dev) — Live preview

---

## License

[Apache-2.0](LICENSE)