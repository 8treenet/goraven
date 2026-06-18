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
  <a href="https://8tree.net">官网</a> ·
  <a href="https://preview.8tree.net">在线预览</a> ·
  <a href="https://discord.gg/derR7CBYDW">Discord</a>
</p>

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/cn_chat.png?raw=true" alt="聊天界面" width="90%" />
</p>

---

## 别再只做一个会聊天的 Agent

Raven 是一套自部署的 AI Agent Harness。它把模型、工具、知识、文件、技能和工作流装配到同一个多用户运行环境里，让 Agent 能接任务、调工具、读资料、写文件、沉淀能力，真正进入团队的日常工程链路。

基础 Agent 能力已经不稀缺，真正稀缺的是把它变成团队可持续使用的工程系统。Raven 提供的不是又一个对话框，而是一套 **AI Harness**——它负责把模型接入、任务编排、工具调用、知识检索、文件工作区、技能分发和运行治理连接起来。个人可以用它完成复杂任务，团队可以共享工作区、复用技能、沉淀最佳实践，组织可以在自己的基础设施上持续演进 AI 能力。

---

## 快速开始

### 国内镜像（推荐）

```bash
docker pull docker.1panel.live/8treenet/raven:v0.1.2

docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -e RAVEN_CHINA_MIRROR=1 \
  docker.1panel.live/8treenet/raven:v0.1.2
```

### Docker Hub

```bash
docker pull 8treenet/raven:latest

docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  8treenet/raven:latest
```

### 挂载数据盘（持久化）

```bash
docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -v /opt/raven_data:/raven/data \
  -e RAVEN_CHINA_MIRROR=1 \
  docker.1panel.live/8treenet/raven:v0.1.2
```

启动后访问 `http://localhost:8000`，按提示完成初始化，创建管理员账户即可使用。

---

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/cnd.png?raw=true" alt="仪表盘" width="90%" />
</p>

---

## 能力

- **模型调度** — 接入 OpenAI、Claude、DeepSeek、Qwen、GLM 及兼容接口，按任务选择模型，按成本和质量调度资源
- **长任务编排** — 主 Agent 理解目标，子 Agent 并行拆解执行，运维 Agent 整理上下文
- **技能市场** — 提示词、命令、脚本、工作流都可沉淀为技能，团队成员一键安装，版本集中维护
- **共享工作台** — 每个用户独立文件系统，团队共享空间和项目协作区，Agent 可读写文件、执行命令
- **知识库** — 制度、文档、业务资料进入 RAG 知识库，Agent 实时检索上下文，让知识参与执行
- **MCP 工具链** — 连接内部 API、数据库、私有服务和命令行工具，Agent 不只生成建议，还能执行动作
- **插件 Hook** — 生命周期 Hook 允许在对话前后、工具调用、事件流节点注入逻辑，无需 fork 核心代码
- **运营面板** — 使用量、模型消耗、用户活跃度集中呈现，观察团队使用情况并持续优化

---

## 路线图

- Core 代码整理并开放
- 多用户沙盒
- API
- 多人协作

---

## License

Apache-2.0
