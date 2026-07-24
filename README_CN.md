<p align="center">
  <img src="https://raw.githubusercontent.com/8treenet/blog/9d6f2d861fbb7c5f4627a9a5b1a3472fb4236881/img/favicon.svg" alt="Raven" width="80" />
</p>

<h1 align="center">Raven</h1>

<p align="center">
  <strong>开源 · 自部署 · AI 工坊</strong><br/>
  模型、工具、知识、技能、工作流，一个运行时全装进去。Agent 不只是聊天，是真的能干活。
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=white" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-6-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript" />
  <img src="https://img.shields.io/badge/Vite-8-646CFF?style=flat-square&logo=vite&logoColor=white" alt="Vite" />
  <img src="https://img.shields.io/badge/Tailwind_CSS-4-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white" alt="Tailwind CSS" />
  <img src="https://img.shields.io/badge/Docker-就绪-2496ED?style=flat-square&logo=docker&logoColor=white" alt="Docker" />
  <img src="https://img.shields.io/badge/License-Apache%202.0-green?style=flat-square" alt="License" />
</p>

<p align="center">
  <a href="https://8tree.net">官网</a> ·
  <a href="https://preview.8tree.net">在线体验</a> ·
  <a href="https://discord.gg/derR7CBYDW">Discord</a> ·
  <a href="./README.md">English</a>
</p>

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/cn_chat.png?raw=true" alt="Raven 聊天界面" width="90%" />
</p>

---

## 它是什么

Raven 给团队里每个人一个独立的 Agent 工作空间。不是聊天框，是工位。Agent 在里面读文件、写代码、跑命令、调接口、查知识库、分析数据、写报告、做图表。你交代任务，它自己拆步骤、并行推进、交付结果。

你拿到的是成果，不是一段需要复制粘贴的文本。

---

## 为什么选 Raven

<table>
<tr>
<td width="50%">

### 🎯 团队优先

每人独立工作空间，团队共享项目和技能库。管理员统一管控模型配额、工具权限、数据访问。不是给一个人用的玩具。

</td>
<td width="50%">

### 🔧 真能动手

Agent 直接读写文件、执行 Shell、调用 MCP 工具链。内部 API、数据库、私有服务都能接进来，不只是给你一段建议文本。

</td>
</tr>
<tr>
<td>

### 🧩 技能市场

提示词、脚本、工作流打包成技能，团队一键安装、统一维护。一个人踩过的坑，全队不用再踩一遍。

</td>
<td>

### 📚 知识库接入执行

制度、文档、业务资料进 RAG 知识库，Agent 执行时实时检索。回答有依据，不凭记忆瞎编。

</td>
</tr>
<tr>
<td>

### 🔌 插件 Hook

对话前后、工具调用、SSE 事件流，生命周期各节点都能插入自定义逻辑。不用 Fork 核心代码也能深度定制。

</td>
<td>

### 🏠 完全自部署

一条 Docker 命令跑起来。数据不出你的服务器，支持 OpenAI、Claude、DeepSeek、Qwen、GLM 及任何兼容接口。

</td>
</tr>
</table>

---

## 快速开始

### 国内镜像（推荐）

```bash
docker pull docker.1panel.live/8treenet/raven:v0.3.1

docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -e RAVEN_CHINA_MIRROR=1 \
  docker.1panel.live/8treenet/raven:v0.3.1
```

### 持久化数据

```bash
docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -v /opt/raven_data:/raven/data \
  -e RAVEN_CHINA_MIRROR=1 \
  docker.1panel.live/8treenet/raven:v0.3.1
```

启动后访问 `http://localhost:8000`，按引导完成初始化，创建管理员账户即可使用。

---

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/cnd.png?raw=true" alt="Raven 运营面板" width="90%" />
</p>

---

## 架构

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/architecture.svg?raw=true" alt="Raven Architecture" width="90%" />
</p>

---

## 设计原则

模型为道，Agent 为器。道可进化，器唯工善。

我们不做什么“记忆与进化”——若不能转化为可泛化的规则或实时的检索增强，那只是刻舟求剑的噪音。模型的进化是模型自己的事，Agent 不该越俎代庖。

Agent 作为工具，本分只有三件事：流程的确定性、工具调用的精度、执行的鲁棒性。假器以自生慧，真器以能尽物用。

打开 Raven，东西在哪、怎么用，不需要看文档就能明白。这是我们理解的“好用”。

---

## 路线图

- [ ] `core` 目录核心代码开放（1k Star 时开放）
- [ ] 安全沙盒
- [ ] 开放 API

---

## 社区

- 💬 [Discord](https://discord.gg/derR7CBYDW) — 问题反馈、功能讨论、经验分享
- 🌐 [8tree.net](https://8tree.net) — 官网
- 🚀 [preview.8tree.net](https://preview.8tree.net) — 在线体验

---

## License

[Apache-2.0](LICENSE)