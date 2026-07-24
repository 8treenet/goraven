<p align="center">
  <img src="https://raw.githubusercontent.com/8treenet/blog/9d6f2d861fbb7c5f4627a9a5b1a3472fb4236881/img/favicon.svg" alt="Raven" width="80" />
</p>

<h1 align="center">Raven</h1>

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
  <img src="https://github.com/8treenet/blog/blob/master/img/enchat.png?raw=true" alt="Raven Chat" width="90%" />
</p>

---

## What It Is

Raven gives every person on your team an independent Agent workspace. Not a chat window — a workstation. The Agent reads files, writes code, runs commands, calls APIs, searches your knowledge base, analyzes data, writes reports, builds charts. You assign the task; it breaks it down, works in parallel, and delivers the result.

What you get is the outcome, not a block of text you have to copy-paste.

---

## Why Raven

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
docker pull 8treenet/raven:latest

docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  8treenet/raven:latest
```

### Persistent Data

```bash
docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -v /opt/raven_data:/raven/data \
  8treenet/raven:latest
```

After startup, visit `http://localhost:8000` and follow the setup wizard to create your admin account.

> The container uses UTC by default. To set a different timezone, add `-e TZ=America/New_York`. Available timezones: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

---

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/cnd.png?raw=true" alt="Raven Dashboard" width="90%" />
</p>

---

## Architecture

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/architecture.svg?raw=true" alt="Raven Architecture" width="90%" />
</p>

---

## Design Principles

The model is the Way; the Agent is the instrument. The Way evolves on its own — the instrument need only be precise.

We don't do "memory and evolution" — if it can't be distilled into generalizable rules or real-time retrieval augmentation, it's just noise. The model's evolution is the model's own business; the Agent shouldn't overstep.

As a tool, the Agent has only three duties: deterministic workflows, precise tool invocation, and robust execution. Borrow the instrument to let wisdom arise naturally — a true instrument is one that fulfills its purpose to the fullest.

Open Raven, and where things are and how they work should be obvious without reading docs. That's our definition of "good."

---

## Roadmap

- [ ] `core` directory core code open-sourced (Available at 1k Stars)
- [ ] Security sandbox
- [ ] Open API

---

## Community

- 💬 [Discord](https://discord.gg/derR7CBYDW) — feedback, discussions, tips
- 🌐 [goraven.dev](https://goraven.dev) — Website
- 🚀 [preview.goraven.dev](https://preview.goraven.dev) — Live preview

---

## License

[Apache-2.0](LICENSE)