<p align="center">
  <img src="https://raw.githubusercontent.com/8treenet/blog/9d6f2d861fbb7c5f4627a9a5b1a3472fb4236881/img/favicon.svg" alt="Raven" width="80" />
</p>

<h1 align="center">Raven</h1>

<p align="center">
  <strong>开源 · 自部署 · AI 工坊</strong><br/>
  把模型、工具、知识、技能和工作流装进同一个运行时——让 Agent 替你干活，而不只是陪你聊天。
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

## 它不只是聊天。它替你干活。

AI 聊天已经不新鲜了。真正难的是把它变成团队每天能用的工程系统。

Raven 是一套 **AI 工坊（AI Harness）**——每个团队成员有自己的独立工作空间，Agent 在里面**读文件、写代码、执行命令、调用 API、检索知识库**。它能接任务、拆步骤、调工具、出结果，而不只是生成一段文字等你复制粘贴。

```
你提需求 → Agent 理解目标 → 拆解成子任务 → 并行执行 → 整理结果交付
               ↓
     读文档 · 写代码 · 调接口 · 搜知识库 · 装依赖
```

**从「模型输出文本」到「Agent 交付成果」**，这是 Raven 做的事。

---

## 为什么选 Raven？

<table>
<tr>
<td width="50%">

### 🎯 团队优先，不是单人玩具

每人独立工作空间，团队共享项目和技能库。管理员统一管控模型配额、工具权限、数据访问。

</td>
<td width="50%">

### 🔧 能动嘴，更能动手

Agent 不止生成建议——它直接**读写文件、执行 Shell、调用 MCP 工具链**，把内部 API、数据库、私有服务全接进来。

</td>
</tr>
<tr>
<td>

### 🧩 技能市场，能力可复用

提示词、脚本、工作流打包成技能，团队一键安装、版本统一维护。一个人踩过的坑，全队下次直接跳过。

</td>
<td>

### 📚 知识库参与执行

制度、文档、业务资料进入 RAG 知识库，Agent 实时检索上下文。**知识不再是静态文档，而是 Agent 做决策的依据。**

</td>
</tr>
<tr>
<td>

### 🔌 插件 Hook，不 Fork 也能扩展

对话前后、工具调用、SSE 事件流——生命周期 Hook 让你注入自定义逻辑，不改核心代码也能深度定制。

</td>
<td>

### 🏠 数据自主，完全自部署

一台 Docker 命令就跑起来。模型数据不出你的服务器，支持 OpenAI、Claude、DeepSeek、Qwen、GLM 及兼容接口。

</td>
</tr>
</table>

---

## 快速开始

### 国内镜像（推荐）

```bash
docker pull docker.1panel.live/8treenet/raven:v0.2.0

docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -e RAVEN_CHINA_MIRROR=1 \
  docker.1panel.live/8treenet/raven:v0.2.0
```

### 持久化数据

```bash
docker run -d --restart=always --name raven-agent \
  -p 8000:8000 \
  -v /opt/raven_data:/raven/data \
  -e RAVEN_CHINA_MIRROR=1 \
  docker.1panel.live/8treenet/raven:v0.2.0
```

启动后访问 `http://localhost:8000`，按引导完成初始化，创建管理员账户即可使用。

---

<p align="center">
  <img src="https://github.com/8treenet/blog/blob/master/img/cnd.png?raw=true" alt="Raven 运营面板" width="90%" />
</p>

---

## 能力全景

| 能力 | 说明 |
|------|------|
| 🧠 **模型调度** | 不绑死一家厂商。任务来了自动选最合适的模型——贵的干精细活，便宜的干粗活。Token 花在刀刃上。 |
| ⚡ **Agent 编排** | 复杂任务自动拆解，多 Agent 并行推进，上下文自主收敛。像一个 AI 项目经理调度着一支微型工程团队——你只需交代目标，剩下的交给它。 |
| 🎯 **技能市场** | 好用的提示词、顺手的脚本、验证过的工作流——打包成技能，全队一键安装。一个人的经验，沉淀为团队的肌肉记忆。 |
| 👥 **团队协作** | 每人独立工作空间，Agent 在自己的沙箱里干活。团队共享区内，技能、文件、会话全打通——管理员掌控全局，成员各干各的又不打架。 |
| 📖 **知识库** | 把文档、规范、业务知识喂进去，Agent 回答时自动翻资料、找依据。不再凭记忆瞎编，每句话都有出处。 |
| 🛡️ **安全沙盒** | 命令执行全隔离——权限最小化、网络可控、文件系统独立。哪怕是第三方技能，也碰不到你的宿主环境。安全是地基，不是装修。 |
| 🔌 **MCP 工具链** | 内部 API、数据库、私有服务、命令行——全接进来。Agent 不光是动嘴建议，它直接上手查数据、调接口、执行动作。 |
| 🪝 **插件 Hook** | 对话前后、工具调用、事件推送——任意节点都能插入你的逻辑。深度定制 Agent 行为，不碰一行核心代码。 |
| 📊 **运营面板** | 谁用了多少 Token、哪个模型最费钱、团队活跃度怎么样——一张面板看全。让 AI 的投入产出算得清账。 |

---

## 小而美，不妥协

Raven 不是 KPI 项目，也永远不会是。

每一个功能、每一行代码，哪怕只是一个按钮，都得有它存在的理由——否则不上线。我们追求的是一个**锐利、可靠的工具**，而不是靠堆砌功能讨好评审委员会的大而全平台。

不膨胀。不功利。不做企业级虚荣。

我们也不玩概念、不炒热词。比起讲一个漂亮的故事，我们更在意**工程细节是否到位、产品是否真的顺手好用**。你打开 Raven，东西在哪儿、怎么用，应该一目了然——这才是我们理解的"好用"。

> 就只是一个把一件事做好的团队 Agent 平台。

---

## 路线图

- [ ] 核心代码开放
- [ ] 团队协作
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