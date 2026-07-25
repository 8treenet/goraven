package seed

import "strings"

var PresetSkillNames = map[string]bool{
	"goraven-install-skill": true,
	"goraven-chart":         true,
	"goraven-guide":         true,
	"goraven-runtime":       true,
	"goraven-features":      true,
	"goraven-user-ui":       true,
	"goraven-admin-ui":      true,
	"goraven-troubleshoot":  true,
}

func ParseSkillDescription(content string) string {
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	foundFirst := false
	for _, line := range lines {
		if line == "---" {
			if !foundFirst {
				foundFirst = true
				inFrontmatter = true
			} else {
				break
			}
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "description:") {
			desc := strings.TrimPrefix(line, "description:")
			desc = strings.TrimSpace(desc)
			return desc
		}
	}
	return ""
}

const SystemSkillInstall = `---
name: goraven-install-skill
description: "当用户要求安装/下载/添加技能时（任意来源，包括商店CLI），必须先调用此技能获取目录规范和配置流程，否则路径错误或配置遗漏。"
---

# 技能安装

当用户提出任何与**安装技能**相关的请求时，**必须先调用本技能**，再执行任何其他操作。即使使用商店 CLI（如 skillhub install），也必须先调用本技能确认目录结构。

**适用场景（包括但不限于）：**
- 从 SkillHub / clawhub / 其他商店安装技能
- 要求"安装 xxx 技能"、"下载 xxx 技能"
- 手动创建或部署技能文件
- 配置技能的 API Key 等环境变量
- 从 git 仓库克隆技能
- 从压缩包解压技能

**不适用的场景：**
- 安装外部系统工具（如 Python 包、系统命令等），这属于依赖安装而非技能安装

## 安装流程

### 1. 确认目标路径

<根路径> 已在系统提示词中给出。技能安装到 <根路径>/skills/<技能名>/ 目录：

<根路径>/skills/<技能名>/
├── SKILL.md（必需）
├── _meta.json（可选）
├── scripts/（可选）
├── references/（可选）
└── assets/（可选）

### 2. 检查是否已存在

安装前先确认 skills/<技能名>/ 目录是否已存在。如已存在，询问用户是否覆盖更新。

### 3. 安装方式

无论通过哪种方式获取技能文件，最终都要确保技能目录结构正确：

- 从商店安装（skillhub install）：执行商店 CLI 命令后，将技能文件同步到用户 skills 目录
- 手动创建：使用 mkdir + write_file 创建
- 从压缩包/仓库获取：解压或 clone 后复制到用户 skills 目录

### 4. 配置环境变量

如果技能需要 API Key 等环境变量，写入 <根路径>/.profile（格式与规则同系统提示词中的环境变量规则），变量值必须从用户处获取，禁止编造。

### 5. 通知用户

安装完成后告知用户技能已可用，可在技能中心页面点击刷新按钮查看。

## 规则

- 所有路径使用以 <根路径> 开头的绝对路径
- 技能名必须唯一，下载前先确认目标目录是否已存在
- skills/ 目录下已有的技能文件夹不要修改
- 商店 CLI（如 clawhub install）默认装到全局目录，必须将技能文件实际拷贝到 <根路径>/skills/<技能名>/，禁止使用软链接
`

const SystemSkillInstallEn = `---
name: goraven-install-skill
description: "When users ask to install, download, or add skills from any source, including marketplace CLIs, invoke this skill first to get the directory rules and configuration flow."
---

# Skill Installation

When a user makes any request related to **installing skills**, you **must invoke this skill first** before taking any other action. Even when using a marketplace CLI such as skillhub install, invoke this skill first to confirm the directory structure.

**Applicable scenarios include, but are not limited to:**
- Installing skills from SkillHub, clawhub, or other marketplaces
- Requests such as "install the xxx skill" or "download the xxx skill"
- Manually creating or deploying skill files
- Configuring environment variables such as API keys for skills
- Cloning skills from git repositories
- Extracting skills from archives

**Not applicable:**
- Installing external system tools such as Python packages or shell commands. These are dependency installations, not skill installations.

## Installation Flow

### 1. Confirm the Target Path

The <root path> is provided in the system prompt. Install skills under <root path>/skills/<skill name>/:

<root path>/skills/<skill name>/
├── SKILL.md (required)
├── _meta.json (optional)
├── scripts/ (optional)
├── references/ (optional)
└── assets/ (optional)

### 2. Check Whether It Already Exists

Before installing, confirm whether the skills/<skill name>/ directory already exists. If it exists, ask the user whether to overwrite or update it.

### 3. Installation Methods

No matter how the skill files are obtained, the final skill directory structure must be correct:

- Marketplace install (skillhub install): after running the marketplace CLI command, sync the skill files into the user's skills directory
- Manual creation: use mkdir and write_file to create the files
- Archive or repository source: extract or clone first, then copy the files into the user's skills directory

### 4. Configure Environment Variables

If the skill needs environment variables such as API keys, write them to <root path>/.profile using the format and rules from the system prompt. Values must come from the user. Never invent them.

### 5. Notify the User

After installation, tell the user the skill is available and can be viewed by clicking the refresh button on the Skill Center page.

## Rules

- Use absolute paths that start with <root path>
- Skill names must be unique. Check whether the target directory exists before downloading
- Do not modify existing skill folders under skills/
- Marketplace CLIs (e.g. clawhub install) install to a global directory by default. Skill files must be physically copied into <root path>/skills/<skill name>/. Soft links are forbidden
`

const SystemSkillChart = `---
name: goraven-chart
description: 当用户要求数据分析、统计对比、趋势展示时，使用<goraven-chart>标签生成可视化图表
---

# 统计图表生成

使用 <goraven-chart> 标签在回复中嵌入数据可视化图表。支持柱状图、折线图、饼图、面积图、散点图。

## 标签格式

<goraven-chart
  type="bar|line|pie|area|scatter"
  title="图表标题（可选）"
  x="['标签1','标签2','标签3']"
  labels="['标签1','标签2','标签3']"
  y1="[数值1,数值2,数值3]"
  y1name="系列名"
  y2="[数值1,数值2,数值3]"
  y2name="系列名"
  y3="[数值1,数值2,数值3]"
  y3name="系列名"
  height="280"
/>

## 属性说明

- **type**: 必填，图表类型。bar（柱状图）、line（折线图）、pie（饼图）、area（面积图）、scatter（散点图）
- **title**: 可选，图表标题
- **x**: X轴标签数组，JSON格式。饼图不使用此属性
- **labels**: 饼图的扇区标签数组，JSON格式。非饼图可使用x代替
- **y1/y2/y3**: 数值数组，JSON格式。y1必填，y2和y3用于多系列对比
- **y1name/y2name/y3name**: 对应系列的名称，显示在图例中
- **height**: 可选，图表高度（像素），默认280

## 使用场景

- bar（柱状图）—— 分类对比，如季度营收、部门预算
- line（折线图）—— 趋势变化，如CPU使用率、用户增长
- pie（饼图）—— 占比分布，如错误类型分布、市场份额
- area（面积图）—— 体量趋势，如流量变化、存储用量
- scatter（散点图）—— 相关性分析，如请求量与延迟关系

## 示例

柱状图（多系列对比）:
<goraven-chart type="bar" title="季度营收对比" x="['Q1','Q2','Q3','Q4']" y1="[120,200,150,260]" y1name="2025" y2="[80,140,100,180]" y2name="2024" height="280" />

折线图（单系列趋势）:
<goraven-chart type="line" title="CPU使用率" x="['10:00','10:30','11:00','11:30','12:00']" y1="[23,45,67,89,56]" y1name="使用率(%)" height="280" />

饼图（占比分布）:
<goraven-chart type="pie" title="错误分布" labels="['超时','连接拒绝','参数错误','其他']" y1="[45,30,18,7]" height="280" />

面积图（流量趋势）:
<goraven-chart type="area" title="流量趋势" x="['Mon','Tue','Wed','Thu','Fri','Sat','Sun']" y1="[1200,1900,1700,2100,2400,1800,900]" y1name="请求数" height="280" />

## 规则

- 所有数组使用JSON格式，如 [120, 200, 150]
- 标签数组使用单引号包裹字符串，如 ['Q1','Q2','Q3']
- <goraven-chart> 标签放在回复内容的尾部
- 数值应基于实际数据，不要编造不存在的统计数字
- 单个回复最多3个图表，避免信息过载
`

const SystemSkillChartEn = `---
name: goraven-chart
description: Use <goraven-chart> tags to generate visual charts when users ask for data analysis, statistical comparison, or trend presentation
---

# Statistical Chart Generation

Use <goraven-chart> tags to embed data visualization charts in replies. Supported chart types are bar, line, pie, area, and scatter.

## Tag Format

<goraven-chart
  type="bar|line|pie|area|scatter"
  title="Chart title (optional)"
  x="['Label 1','Label 2','Label 3']"
  labels="['Label 1','Label 2','Label 3']"
  y1="[Value 1,Value 2,Value 3]"
  y1name="Series name"
  y2="[Value 1,Value 2,Value 3]"
  y2name="Series name"
  y3="[Value 1,Value 2,Value 3]"
  y3name="Series name"
  height="280"
/>

## Attributes

- **type**: required chart type. bar, line, pie, area, or scatter
- **title**: optional chart title
- **x**: X-axis label array in JSON format. Pie charts do not use this attribute
- **labels**: sector label array for pie charts in JSON format. Non-pie charts can use x instead
- **y1/y2/y3**: numeric arrays in JSON format. y1 is required; y2 and y3 are for multi-series comparisons
- **y1name/y2name/y3name**: series names shown in the legend
- **height**: optional chart height in pixels. Default is 280

## Use Cases

- bar: category comparisons, such as quarterly revenue or department budgets
- line: trends over time, such as CPU usage or user growth
- pie: proportions and distributions, such as error type distribution or market share
- area: volume trends, such as traffic changes or storage usage
- scatter: correlation analysis, such as request volume versus latency

## Examples

Bar chart (multi-series comparison):
<goraven-chart type="bar" title="Quarterly Revenue Comparison" x="['Q1','Q2','Q3','Q4']" y1="[120,200,150,260]" y1name="2025" y2="[80,140,100,180]" y2name="2024" height="280" />

Line chart (single-series trend):
<goraven-chart type="line" title="CPU Usage" x="['10:00','10:30','11:00','11:30','12:00']" y1="[23,45,67,89,56]" y1name="Usage (%)" height="280" />

Pie chart (proportion distribution):
<goraven-chart type="pie" title="Error Distribution" labels="['Timeout','Connection Refused','Invalid Parameter','Other']" y1="[45,30,18,7]" height="280" />

Area chart (traffic trend):
<goraven-chart type="area" title="Traffic Trend" x="['Mon','Tue','Wed','Thu','Fri','Sat','Sun']" y1="[1200,1900,1700,2100,2400,1800,900]" y1name="Requests" height="280" />

## Rules

- Use JSON format for all arrays, such as [120, 200, 150]
- Wrap strings in label arrays with single quotes, such as ['Q1','Q2','Q3']
- Put <goraven-chart> tags at the end of the reply
- Numeric values must be based on real data. Do not invent statistics that do not exist
- Use at most 3 charts in a single reply to avoid information overload
`
