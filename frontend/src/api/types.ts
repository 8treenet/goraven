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
  isCompress: number
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
  source: 'market' | 'custom'
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

export interface SessionSimple {
  sessionId: string
  title: string
  status: number
  personaId: number
  project: string
  lastChatTime: string
  created: string
}

export interface SessionDetail extends SessionSimple {
  aiModelId: number
  contextTokens: number
  promptTokensCount: number
  completionTokensCount: number
  mcpIds: number[]
  skillIds: number[]
  modelName: string
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
}

export interface ChatResponse {
  sessionId: string
  session?: SessionDetail
}

/* ---------- Share ---------- */

export interface ShareInfo {
  shareId: string
  sessionId: string
  title: string
  expiresAt: string
  viewCount: number
  isExpired: boolean
  created: string
}

export interface CreateShareRequest {
  title?: string
  expiresIn?: string
}

export interface PublicShare {
  shareId: string
  title: string
  creator: string
  created: string
  expiresAt: string
  viewCount: number
  isExpired: boolean
  messages: Message[]
}

/* ---------- Dashboard ---------- */

export interface DashboardOverview {
  todayTokens: number
  weekTokens: number
  totalTokens: number
  totalSessions: number
  newSessions: number
  sparkline: Array<{ date: string; tokens: number }>
}

export interface TokenTrendItem {
  date: string
  promptTokens: number
  completionTokens: number
}

export interface TokenTrendRsp {
  items: TokenTrendItem[]
}

export interface ModelUsageItem {
  modelName: string
  tokenCount: number
  percentage: number
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
  tokenTrend: TokenTrendItem[]
  modelUsage: ModelUsageItem[]
  skillUsageRank: RankItem[]
  mcpUsageRank: RankItem[]
  toolUsageRank: RankItem[]
  storageStats: StorageStats
}

/* ---------- Preference ---------- */

export interface PreferenceData {
  language: 'zh' | 'en'
  domain: string
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
  contextLen: number
  extraFields: string
  isDefault: number
  isCompress: number
  isVisual: number
  status: number
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
  contextLen: number
  extraFields: string
  isDefault: number
  isCompress: number
  isVisual: number
  status: number
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
  tokenTrend: TokenTrendItem[]
  modelUsage: ModelUsageItem[]
  userTokenRank: UserTokenRankItem[]
  activeTrend: ActiveTrendItem[]
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
