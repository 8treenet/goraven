/* ============================================
   Canonical API types — single source of truth
   Aligned with backend protocol /docs/protocol/api-docs
   ============================================ */

/* ---------- Generic ---------- */

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

export interface PaginatedResponse<T> {
  list: T[]
  totalPage: number
  totalCount: number
  page: number
  pageSize: number
}

export interface PaginationParams {
  page?: number
  pageSize?: number
}

/* ---------- User ---------- */

export interface User {
  userId: string
  username: string
  email: string
  role: number
  status: number
  nickname: string
  avatar: string
  created: string
}

export interface AdminUserItem extends User {
  sessionCount: number
  lastActiveTime: string | null
  dailyTokenLimit: number
  updated: string
}

/* ---------- Model ---------- */

export interface ModelInfo {
  aiModelId: number
  providerDisplayName: string
  displayName: string
  modelName: string
  icon: string
  contextLen: number
  isDefault: number
  isFlash: number
  isVisual: number
}

/* ---------- MCP ---------- */

export interface McpInfo {
  mcpId: number
  name: string
  displayName: string
  icon: string
  description: string
}

/* ---------- Persona ---------- */

export interface PersonaSimple {
  personaId: number
  name: string
  icon: string
}

export interface PersonaListItem {
  personaId: number
  name: string
  icon: string
  roleInfo: string
  categoryId: number
  categoryName: string
  categoryIcon: string
  modelName: string
  mcpIds: number[]
  skillIds: number[]
  mcpNames: string[]
  skillNames: string[]
}

export interface PersonaDetail extends PersonaSimple {
  roleInfo: string
  categoryId: number
  mcpIds: number[]
  skillIds: number[]
  aiModelId: number
  created: string
  updated: string
  categoryName: string
  categoryIcon: string
  modelName: string
  modelIcon: string
}

export interface PersonaCategory {
  categoryId: number
  name: string
  icon: string
}

export interface CreatePersonaRequest {
  name: string
  icon?: string
  roleInfo: string
  categoryId: number
  mcpIds?: number[]
  skillIds?: number[]
  aiModelId?: number
  templateId?: number
}

export interface UpdatePersonaRequest {
  name?: string
  icon?: string
  roleInfo?: string
  categoryId?: number
  mcpIds?: number[]
  skillIds?: number[]
  aiModelId?: number
}

export interface PersonaTemplateItem {
  templateId: number
  name: string
  icon: string
  description: string
  categoryId: number
  categoryName: string
  categoryIcon: string
}

export interface PersonaTemplateDetail extends PersonaTemplateItem {
  roleInfo: string
}

/* ---------- Skill ---------- */

export interface SkillSimple {
  userSkillId: number
  skillName: string
  description: string
  icon: string
  source: 'market' | 'custom' | 'share'
  categoryId: number
  categoryName: string
}

export interface MarketSkill {
  skillId: number
  name: string
  description: string
  icon: string
  source: string
  categoryId: number
  categoryName: string
  installedCount: number
  userInstalled: boolean
  updated: string
}

export interface UserSkill {
  userSkillId: number
  skillName: string
  description: string
  icon: string
  marketSkillId: number
  categoryId: number
  categoryName: string
  source: string
  installStatus: number
  installError?: string
  alwaysOn: number
  created: string
  updated: string
}

export interface MarketSkillDetail extends MarketSkill {
  content: string
}

export interface UserSkillDetail extends UserSkill {
  content: string
  isShared: boolean
}

export interface ShareSkill {
  shareId: number
  ownerId: string
  ownerName: string
  skillName: string
  description: string
  icon: string
  categoryId: number
  categoryName: string
  note: string
  installCount: number
  created: string
  updated: string
  canDelete: boolean
}

export interface ShareSkillDetail extends ShareSkill {
  content: string
}

export interface ShareSkillResult {
  shareId: number
}

export interface SkillCategory {
  categoryId: number
  name: string
  icon: string
}

export interface InstallSkillResult {
  userSkillId: number
}

export interface UserSkillStatus {
  userSkillId: number
  installStatus: number
  installError: string
}

export interface RefreshSkillsResult {
  added: number
  removed: number
}

/* ---------- Session / Chat ---------- */

export interface SharedProjectInfo {
  id: number
  creatorId: string
  creatorName: string
  projectName: string
  description: string
}

export interface SessionSimple {
  sessionId: string
  title: string
  status: number
  personaId: number
  project: string
  sharedProject?: SharedProjectInfo
  lastChatTime: string
  created: string
}

export interface SessionDetail extends SessionSimple {
  aiModelId: number
  contextTokens: number
  promptTokensCount: number
  completionTokensCount: number
  promptCachedTokens: number
  mcpIds: number[]
  skillIds: number[]
  modelName: string
  modelIcon: string
  personaName: string
  personaIcon: string
  contextLimit: number
}

export interface ReasoningItem {
  eventType: 'reasoning' | 'tool'
  content?: string
  tool?: {
    name: string
    displayName: string
    icon: string
    action: string
  }
}

export interface Message {
  msgId: string
  roundId: string
  contextState: number
  content: string
  reasoningContent: ReasoningItem[]
  roleType: 'user' | 'assistant' | 'summary'
  created: string
}

export interface ChatRequest {
  sessionId?: string
  content: string
  attachments: string[]
  aiModelId: number
  personaId?: number
  mcpIds: number[]
  skillIds: number[]
  reasoning: number
  project: string
  sharedProjectId?: number
}

export interface ChatResponse {
  sessionId: string
  session?: SessionDetail
}

/* ---------- Share ---------- */

export type ShareType = 'public' | 'internal'

export interface ShareInfo {
  shareId: string
  sessionId: string
  title: string
  shareType: ShareType
  expiresAt: string
  viewCount: number
  isExpired: boolean
  created: string
}

export interface CreateShareRequest {
  title?: string
  expiresIn?: string
  shareType?: ShareType
  domain?: string
}

export interface PublicShare {
  shareId: string
  title: string
  creator: string
  shareType: ShareType
  created: string
  expiresAt: string
  viewCount: number
  isExpired: boolean
}

/* ---------- Dashboard ---------- */

export interface DashboardOverview {
  todayTokens: number
  weekTokens: number
  totalTokens: number
  dailyTokenLimit: number
  totalSessions: number
  newSessions: number
  sparkline: Array<{ date: string; tokens: number }>
}

export interface TokenTrendItem {
  date: string
  promptTokens: number
  promptCachedTokens: number
  completionTokens: number
}

export interface TokenTrendRsp {
  items: TokenTrendItem[]
}

export interface ModelUsageRsp {
  items: ModelUsageItem[]
}

export interface ModelUsageItem {
  modelName: string
  tokenCount: number
  percentage: number
  promptTokens: number
  completionTokens: number
}

export interface RankItem {
  name: string
  count: number
}

export interface StorageStats {
  usedBytes: number
  freeBytes: number
  totalBytes: number
  items: Array<{ name: string; bytesSize: number; percentage: number }>
}

export interface DashboardData {
  overview: DashboardOverview
  skillUsageRank: RankItem[]
  mcpUsageRank: RankItem[]
  toolUsageRank: RankItem[]
  storageStats: StorageStats
}

/* ---------- Preference ---------- */

export interface PreferenceData {
  language: 'zh' | 'en'
}

/* ---------- Admin: Model ---------- */

export interface AdminModelItem {
  aiModelId: number
  providerDisplayName: string
  displayName: string
  providerId: string
  modelName: string
  icon: string
  apiKey: string
  apiKeyMasked: string
  baseUrl: string
  proxyUrl: string
  conversationHeaderKey: string
  contextLen: number
  extraFields: string
  isDefault: number
  isFlash: number
  isVisual: number
  status: number
  access: number
  remark: string
  created: string
  updated: string
}

export interface AdminModelDetail {
  aiModelId: number
  providerDisplayName: string
  displayName: string
  providerId: string
  modelName: string
  icon: string
  apiKey: string
  baseUrl: string
  proxyUrl: string
  conversationHeaderKey: string
  contextLen: number
  extraFields: string
  isDefault: number
  isFlash: number
  isVisual: number
  status: number
  access: number
  remark: string
  created: string
  updated: string
}

export interface ProviderItem {
  providerId: string
  providerDisplayNameZh: string
  providerDisplayNameEn: string
  icon: string
  defaultBaseUrl: string
  requireApiKey: boolean
  requireBaseUrl: boolean
}

/* ---------- Admin: MCP ---------- */

export interface AdminMcpItem {
  mcpId: number
  name: string
  displayName: string
  icon: string
  description: string
  transport: string
  httpUrl: string
  httpHeader: string
  httpProxyUrl: string
  stdioType: string
  stdioEnv: string
  stdioArgs: string
  status: number
  alwaysOn: number
  healthLatency: number
  healthCheckedAt: string | null
  remark: string
  created: string
  updated: string
}

/* ---------- Admin: Skill ---------- */

export type AdminSkillStatus = 0 | 1
export type AdminMarketSkillSource = 'clawhub' | 'custom_upload'
export type AdminInstallStatus = 0 | 1 | 2 | 3

export interface AdminSystemSkillItem {
  skillId: number
  name: string
  description: string
  status: AdminSkillStatus
  updated: string
}

export interface AdminSystemSkillDetail extends AdminSystemSkillItem {
  content: string
  created: string
}

export interface AdminMarketSkillItem {
  skillId: number
  name: string
  description: string
  icon: string
  source: AdminMarketSkillSource
  categoryId: number
  categoryName: string
  categoryIcon: string
  installedCount: number
  status: AdminSkillStatus
  sortOrder: number
  updated: string
}

export interface AdminMarketSkillDetail extends AdminMarketSkillItem {
  sourceUrl: string
  remark: string
  content: string
  created: string
}

export interface AdminMarketSkillUserItem {
  userId: string
  installStatus: AdminInstallStatus
  created: string
}

export interface AdminSkillCategoryItem {
  categoryId: number
  name: string
  icon: string
  isDefault: AdminSkillStatus
  skillCount: number
  updated: string
}

export interface AdminSkillCategorySimple {
  categoryId: number
  name: string
  icon: string
  isDefault: AdminSkillStatus
}

export interface AdminSkillCategoryDetail extends AdminSkillCategorySimple {
  created: string
  updated: string
}

export interface ClawHubSearchItem {
  slug: string
  displayName: string
  summary: string
  version: string
  score: number
  updatedAt: number
}

export interface ClawHubSearchResponse {
  results: ClawHubSearchItem[]
}

export interface ClawHubExploreItem {
  slug: string
  displayName: string
  summary: string
  stats: {
    comments: number
    downloads: number
    installsAllTime: number
    installsCurrent: number
    stars: number
    versions: number
  }
  createdAt: number
  updatedAt: number
}

export interface ClawHubExploreResponse {
  items: ClawHubExploreItem[]
  nextCursor: string
}

export interface ClawHubSkillDetail {
  slug: string
  name: string
  description: string
  content: string
  version: string
}

export type SystemSkillItem = AdminSystemSkillItem
export type MarketSkillItem = AdminMarketSkillItem

/* ---------- Admin: Persona ---------- */

export interface AdminPersonaTemplateItem {
  templateId: number
  name: string
  icon: string
  description: string
  roleInfo: string
  categoryId: number
  categoryName: string
  categoryIcon: string
  usageCount: number
  sortOrder: number
  created: string
  updated: string
}

export interface AdminPersonaCategoryItem {
  categoryId: number
  name: string
  icon: string
  isDefault: number
  templateCount: number
  updated: string
}

/* ---------- Admin: System ---------- */

export interface SettingItem {
  key: string
  value: string
  valueType: string
  defaultValue: string
  displayName: string
  description: string
  inputType: string
  min?: number
  max?: number
  placeholder?: string
  displayOrder: number
}

export interface SettingGroup {
  name: string
  displayName: string
  displayOrder: number
  settings: SettingItem[]
}

/* ---------- Admin: Dashboard ---------- */

export interface AdminDashboardOverview {
  activeUsers: number
  activeUsersDiff: number
  totalSessions: number
  newSessions: number
  weekTokens: number
  todayTokens: number
  enabledModels: number
  sparkline: Array<{ date: string; tokens: number }>
}

export interface UserTokenRankItem {
  userId: string
  username: string
  tokenCount: number
  percentage: number
}

export interface ActiveTrendItem {
  date: string
  count: number
}

export interface AdminDashboardData {
  overview: AdminDashboardOverview
  skillUsageRank: RankItem[]
  mcpUsageRank: RankItem[]
  toolUsageRank: RankItem[]
}

/* ---------- Install ---------- */

export interface CheckDbRequest {
  dbType: string
  dbAddr: string
  dbPort: number
  dbUser: string
  dbPass: string
  dbName: string
}

export interface CheckRedisRequest {
  redisAddr: string
  redisPort: number
  redisPass: string
  redisDB: number
}

export interface InitRequest {
  language: string
  domain: string
  username: string
  password: string
  email: string
  dbType: string
  dbAddr: string
  dbPort: number
  dbUser: string
  dbPass: string
  dbName: string
  cacheType: string
  redisAddr: string
  redisPort: number
  redisPass: string
  redisDB: number
}

export interface InitResponse {
  userId: string
  username: string
}

/* ---------- File Manager ---------- */

export interface FileItem {
  name: string
  path: string
  isDir: boolean
  size: number
  modTime: string
  isDefault?: boolean
}

export interface StorageUsage {
  totalSize: number
  usedSize: number
  fileCount: number
}

export interface ProfileEntry {
  key: string
  value: string
}

export interface ProfileListResponse {
  items: ProfileEntry[]
}

/* ---------- Team Project ---------- */

export interface TeamProjectItem {
  id: number
  creatorId: string
  creatorName: string
  creatorAvatar: string
  projectName: string
  description: string
  access: number // 0全员开放 1仅成员可见
  updatedAt: string
  isCreator: boolean
}

export interface TeamProjectListRsp {
  items: TeamProjectItem[]
}

export interface TeamProjectCreateRsp {
  id: number
}

/* ---------- Team Project Members ---------- */

export interface TeamProjectUserItem {
  userId: string
  username: string
  nickname: string
  avatar: string
}

export interface TeamProjectMembersRsp {
  creatorId: string
  memberIds: string[]
}

/* ---------- Admin Team Projects ---------- */

export interface AdminTeamProjectItem {
  id: number
  creatorId: string
  creatorName: string
  creatorAvatar: string
  projectName: string
  description: string
  access: number
  visitCount: number
  lastActiveAt: string | null
  locked: boolean
  lockedBy: string
  created: string
  updated: string
}

export interface AdminTeamProjectListReq {
  search?: string
  page?: number
  pageSize?: number
}

/* ---------- Captcha ---------- */

export interface CaptchaRsp {
  required: boolean
  image1?: string
  image2?: string
}

/* ---------- SSE Events ---------- */

export interface SSEConnectedEvent {
  type: 'connected'
  sessionId: string
}

export interface SSEReasoningEvent {
  type: 'reasoning'
  content: string
}

export interface SSEContentEvent {
  type: 'content'
  content: string
}

export interface SSEToolEvent {
  type: 'tool'
  tool: {
    name: string
    displayName: string
    icon: string
    action: string
  }
}

export interface SSERetryEvent {
  type: 'retry'
  retry: {
    attempt: number
    maxRetries: number
    error: string
  }
}

export interface SSEEndEvent {
  type: 'end'
}

export interface SSEContextEvent {
  type: 'context'
  context: {
    tokens: number
    limit: number
  }
}

export interface SSEHeartbeatEvent {
  type: 'heartbeat'
}

export type SSEEvent =
  | SSEConnectedEvent
  | SSEReasoningEvent
  | SSEContentEvent
  | SSEToolEvent
  | SSERetryEvent
  | SSEEndEvent
  | SSEContextEvent
  | SSEHeartbeatEvent

/* ---------- Automation ---------- */

/** 执行类型：1单次固定时间 2按间隔 3每天固定时间 4每周固定时间 */
export const AutomationExecType = {
  Once: 1,
  Interval: 2,
  Daily: 3,
  Weekly: 4,
} as const

/** 任务状态：0停用 1启用 2已完成 */
export const AutomationStatus = {
  Disabled: 0,
  Enabled: 1,
  Done: 2,
} as const

export interface AutomationTaskItem {
  id: number
  title: string
  userId: string
  execType: number
  runAt: string | null
  intervalMinutes: number
  fixedTime: string
  weekday: number
  mcpIds: string
  skillIds: string
  project: string
  sharedProjectId: number
  aiModelId: number
  personaId: number
  aiModelName: string
  personaName: string
  sharedProjectName: string
  mcpNames: string[]
  skillNames: string[]
  nextRunAt: string
  status: number
  deleted: number
  created: string
  updated: string
}

export interface AutomationTaskDetail extends AutomationTaskItem {
  requirement: string
}

export interface AutomationExecutionItem {
  id: number
  startedAt: string
  finishedAt: string
}

export interface AutomationQARsp {
  question: string
  answer: string
}
