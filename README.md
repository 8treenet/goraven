<p align="center">
  <img src="https://raw.githubusercontent.com/8treenet/blog/9d6f2d861fbb7c5f4627a9a5b1a3472fb4236881/img/favicon.svg" alt="Raven" width="80" />
</p>

<h1 align="center">Raven</h1>

<p align="center">
  <strong>Open Source · Self-Hosted · AI Harness</strong><br/>
  One runtime. Models, tools, knowledge, skills, and workflows — assembled so your Agent delivers results, not just replies.
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

## It doesn't just chat. It gets things done.

AI chat is everywhere. What's hard is turning it into an engineering system your team can actually rely on.

Raven is an **AI Harness** — every team member gets their own workspace, and inside it the Agent **reads files, writes code, runs commands, calls APIs, and searches your knowledge base**. It takes on tasks, breaks them down, picks the right tools, and delivers results — instead of spitting out text for you to copy-paste.

```
You describe the goal → Agent understands it → Decomposes into subtasks → Runs in parallel → Delivers results
                                ↓
              Reads docs · Writes code · Calls APIs · Searches knowledge · Installs deps
```

**From "model outputs text" to "agent delivers outcomes."** That's what Raven does.

---

## Why Raven?

<table>
<tr>
<td width="50%">

### 🎯 Team-first, not a solo toy

Every member gets an isolated workspace. The team shares projects and skill libraries. Admins control model quotas, tool permissions, and data access — all in one place.

</td>
<td width="50%">

### 🔧 Talks less. Does more.

Agents don't stop at suggestions — they **read and write files, run shell commands, call MCP tools**, connecting your internal APIs, databases, and private services directly.

</td>
</tr>
<tr>
<td>

### 🧩 Skill marketplace — reuse what works

Prompts, scripts, and workflows packaged as installable skills. One-click install, automatic dependency resolution, centralized versioning. One person figures it out — the whole team skips the hard part.

</td>
<td>

### 📚 Knowledge that participates

Policies, docs, and business data fed into RAG. Agents retrieve context in real time. **Knowledge stops being static documentation and becomes the basis for every decision.**

</td>
</tr>
<tr>
<td>

### 🔌 Plugin hooks — extend without forking

Before/after conversations, during tool calls, on event streams — lifecycle hooks let you inject custom logic anywhere. Deep customization without touching a line of core code.

</td>
<td>

### 🏠 Your data, your server

One Docker command and you're running. Model data never leaves your infrastructure. Works with OpenAI, Claude, DeepSeek, Qwen, GLM, and compatible APIs.

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

## Capabilities

| Capability | Description |
|------------|-------------|
| 🧠 **Model Orchestration** | No vendor lock-in. Tasks get routed to the right model — expensive ones for precision work, cheap ones for the heavy lifting. Every token counts. |
| ⚡ **Agent Orchestration** | Complex tasks auto-decomposed, multiple agents running in parallel, context converging on its own. Like an AI project manager running a micro engineering team — you set the goal, it handles the rest. |
| 🎯 **Skill Marketplace** | Battle-tested prompts, handy scripts, proven workflows — packaged as skills, one-click install for the whole team. One person's experience becomes the team's muscle memory. |
| 👥 **Team Collaboration** | Everyone gets their own workspace, agents operate in sandboxed environments. Team spaces share skills, files, and sessions — admins keep control, members stay in their lane, no one steps on each other. |
| 📖 **Knowledge Base** | Feed in docs, specs, and business know-how. When your agent answers, it pulls real sources — no more hallucinating from memory. Every claim has a receipt. |
| 🛡️ **Security Sandbox** | Fully isolated command execution — minimal permissions, network controls, independent filesystems. Even third-party skills can't touch your host. Security is the foundation, not afterthought. |
| 🔌 **MCP Toolchain** | Internal APIs, databases, private services, CLI tools — connect them all. Your agent doesn't just suggest — it queries data, invokes services, and takes action. |
| 🪝 **Plugin Hooks** | Inject logic at any lifecycle point — before/after conversations, on tool calls, on event streams. Deep customization without touching a line of core code. |
| 📊 **Operations Dashboard** | Who's using what, which model costs the most, how active is your team — one panel to see it all. AI spending you can actually account for. |

---

## Small. Beautiful. Uncompromising.

Raven is not a KPI project, and it never will be.

Every feature, every line of code, even a single button — it earns its place, or it doesn't ship. We build a **sharp, reliable tool**, not a bloated platform padded with checkboxes to please a roadmap committee.

No feature creep. No enterprise vanity.

We don't chase buzzwords or ride hype cycles. What matters to us is **engineering quality and whether the product actually feels good to use**. You open Raven, and where things are and how they work should be obvious at a glance — that's our definition of "good."

> Just a team agent platform that does one thing well.

---

## Roadmap

- [ ] Core code open-sourced
- [ ] Team collaboration
- [ ] Security sandbox
- [ ] Open API

---

## Community

- 💬 [Discord](https://discord.gg/derR7CBYDW) — feedback, discussions, tips
- 🌐 [8tree.net](https://8tree.net) — Website
- 🚀 [preview.8tree.net](https://preview.8tree.net) — Live preview

---

## License

[Apache-2.0](LICENSE)
