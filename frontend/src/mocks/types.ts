/* ============================================
   Shared types — aligned with /docs/protocol/api-docs
   ============================================ */

/* ---------- Generic API response ---------- */

export interface ApiResponse<T = unknown> {
  code: number
  msg: string
  data: T
}

export interface PaginatedRequest {
  page?: number
  pageSize?: number
}

export interface PaginatedResponse<T> {
  list: T[]
  totalPage: number
  totalCount: number
  page: number
  pageSize: number
}

/* ---------- Auth & User ---------- */

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  accessToken: string
}

export interface UserInfo {
  userId: string
  username: string
  email: string
  role: number
  status: number
  nickname: string
  avatar: string
  created: string
}

export interface UpdateProfileRequest {
  nickname?: string
  email?: string
  avatar?: string
}

export interface ChangePasswordRequest {
  currentPassword: string
  newPassword: string
}

/* ---------- Preference ---------- */

export interface PreferenceData {
  language: 'zh' | 'en'
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

/* ---------- Session / Chat ---------- */

export interface SessionListItem {
  sessionId: string
  title: string
  status: number
  personaId: number
  lastChatTime: string
  created: string
}

export interface SessionDetail {
  sessionId: string
  title: string
  status: number
  personaId: number
  aiModelId: number
  contextTokens: number
  promptTokensCount: number
  completionTokensCount: number
  mcpIds: number[]
  skillIds: number[]
  lastChatTime: string
  created: string
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

export interface MessageItem {
  msgId: string
  roundId: string
  contextState: number
  content: string
  reasoningContent: ReasoningItem[]
  roleType: 'user' | 'assistant' | 'summary' | 'tool'
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
}

export interface ChatResponse {
  sessionId: string
}

export interface ShareInfo {
  shareId: string
  sessionId: string
  title: string
  expiresAt: string
  viewCount: number
  isExpired: boolean
  created: string
}

export interface ShareLinkRequest {
  title?: string
  expiresIn?: string
}

/* ---------- Persona ---------- */

export interface PersonaListItem {
  personaId: number
  name: string
  icon: string
  categoryName: string
  modelName: string
  roleInfo: string
  mcpNames: string[]
  skillNames: string[]
}

export interface PersonaDetail {
  personaId: number
  name: string
  icon: string
  roleInfo: string
  categoryId: number
  categoryName: string
  categoryIcon: string
  mcpIds: number[]
  mcpNames: { id: number; name: string; icon: string }[]
  skillIds: number[]
  skillNames: { id: number; name: string; icon: string }[]
  aiModelId: number
  modelName: string
  modelIcon: string
  created: string
  updated: string
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

export interface TemplateItem {
  templateId: number
  name: string
  icon: string
  description: string
  categoryId: number
  categoryName: string
  categoryIcon: string
}

export interface TemplateDetail extends TemplateItem {
  roleInfo: string
}

export interface TemplateCategoryItem {
  categoryId: number
  name: string
  icon: string
}

export interface PersonaCategoryItem {
  categoryId: number
  name: string
  icon: string
}

/* ---------- Skill ---------- */

export interface SimpleSkill {
  userSkillId: number
  skillName: string
  description: string
  icon: string
  source: string
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

export interface MarketSkillDetail extends MarketSkill {
  content: string
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
  created: string
  updated: string
}

export interface UserSkillDetail extends UserSkill {
  content: string
}

export interface SkillCategory {
  categoryId: number
  name: string
  icon: string
}

/* ---------- Provider / Model ---------- */

export interface AvailableModel {
  aiModelId: number
  providerDisplayName: string
  displayName: string
  modelName: string
  icon: string
  contextLen: number
  isDefault: boolean
}

/* ---------- MCP ---------- */

export interface McpEndpoint {
  mcpId: number
  name: string
  displayName: string
  icon: string
  description: string
}

/* ---------- File Manager ---------- */

export interface FileItem {
  name: string
  isDir: boolean
  size: number
  modTime: string
  isDefault?: boolean
}

export interface FileListResponse {
  items: FileItem[]
}

export interface StorageUsage {
  totalSize: number
  usedSize: number
  fileCount: number
}

/* ---------- Dashboard ---------- */

export interface SparklineItem {
  date: string
  tokens: number
}

export interface TokenTrendItem {
  date: string
  promptTokens: number
  completionTokens: number
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

export interface StorageBreakdownItem {
  name: string
  bytesSize: number
  percentage: number
}

export interface DashboardOverview {
  todayTokens: number
  weekTokens: number
  totalTokens: number
  totalSessions: number
  newSessions: number
  sparkline: SparklineItem[]
}

export interface DashboardStorageStats {
  usedBytes: number
  freeBytes: number
  totalBytes: number
  items: StorageBreakdownItem[]
}

export interface DashboardData {
  overview: DashboardOverview
  tokenTrend: TokenTrendItem[]
  modelUsage: ModelUsageItem[]
  skillUsageRank: RankItem[]
  mcpUsageRank: RankItem[]
  toolUsageRank: RankItem[]
  storageStats: DashboardStorageStats
}

/* ---------- Admin ---------- */

export interface AdminUserItem {
  userId: string
  username: string
  nickname: string
  email: string
  avatar: string
  role: number
  status: number
  sessionCount: number
  lastActiveTime: string
  created: string
}

export interface AdminModelItem {
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

export interface RecommendModelItem {
  id: string
  object: string
}

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
  healthCheckedAt: string
  remark: string
  created: string
  updated: string
}

export interface McpRecommendItem {
  name: string
  displayName: string
  icon: string
  description: string
  transport: string
  httpUrl: string
  stdioType: string
  stdioArgs: string
  installed: boolean
  mcpId?: number
  mcpStatus?: number
}

export interface AdminSkillCategoryItem {
  categoryId: number
  name: string
  icon: string
  isDefault: number
  created: string
  updated: string
}

export interface SystemSkillItem {
  systemSkillId: number
  name: string
  description: string
  icon: string
  status: number
  created: string
  updated: string
}

export interface SystemSkillDetail extends SystemSkillItem {
  content: string
}

export interface AdminMarketSkillItem {
  marketSkillId: number
  name: string
  description: string
  icon: string
  source: string
  categoryId: number
  categoryName: string
  status: number
  sortOrder: number
  installedCount: number
  remark: string
  created: string
  updated: string
}

export interface ClawHubSkillItem {
  slug: string
  name: string
  description: string
  icon: string
  source: string
  score: number
  downloads: number
  installs: number
  stars: number
}

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
  created: string
  updated: string
}

export interface SettingItem {
  key: string
  value: string | number | boolean
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

export interface SettingGroupData {
  name: string
  displayName: string
  displayOrder: number
  settings: SettingItem[]
}

export interface AdminDashboardOverview {
  activeUsers: number
  activeUsersDiff: number
  totalSessions: number
  newSessions: number
  weekTokens: number
  todayTokens: number
  enabledModels: number
  sparkline: SparklineItem[]
}

export interface ActiveTrendItem {
  date: string
  count: number
}

export interface AdminDashboardData {
  overview: AdminDashboardOverview
  tokenTrend: TokenTrendItem[]
  modelUsage: ModelUsageItem[]
  userTokenRank: RankItem[]
  activeTrend: ActiveTrendItem[]
  skillUsageRank: RankItem[]
  mcpUsageRank: RankItem[]
  toolUsageRank: RankItem[]
}
