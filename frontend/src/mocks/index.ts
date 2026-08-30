/* ============================================
   Mocks barrel — import all mock APIs from this package

   Usage:
     import { getPersonas, getDashboard, ... } from '@/mocks'

   When switching to real API: replace with actual HTTP calls
   ============================================ */

export { delay, listDelay, itemDelay, mutationDelay, heavyDelay, uploadDelay } from './delay'

export type {
  ApiResponse,
  PaginatedRequest,
  PaginatedResponse,
  PreferenceData,
  LoginRequest,
  LoginResponse,
  UserInfo,
  UpdateProfileRequest,
  ChangePasswordRequest,
  CheckDbRequest,
  CheckRedisRequest,
  InitRequest,
  InitResponse,
  SessionListItem,
  SessionDetail,
  MessageItem,
  ChatRequest,
  ChatResponse,
  ShareInfo,
  ShareLinkRequest,
  PersonaListItem,
  PersonaDetail,
  CreatePersonaRequest,
  TemplateItem,
  TemplateDetail,
  TemplateCategoryItem,
  PersonaCategoryItem,
  SimpleSkill,
  MarketSkill,
  MarketSkillDetail,
  UserSkill,
  UserSkillDetail,
  SkillCategory,
  AvailableModel,
  McpEndpoint,
  FileItem,
  FileListResponse,
  StorageUsage,
  DashboardData,
  TokenTrendItem,
  SparklineItem,
  AdminUserItem,
  AdminModelItem,
  ProviderItem,
  AdminMcpItem,
  McpRecommendItem,
  SystemSkillItem,
  SystemSkillDetail,
  AdminMarketSkillItem,
  AdminSkillCategoryItem,
  ClawHubSkillItem,
  AdminPersonaTemplateItem,
  AdminPersonaCategoryItem,
  SettingGroupData,
  AdminDashboardData,
  ActiveTrendItem,
} from './types'

/* ---------- Preference ---------- */
export { getPreference } from './preference'

/* ---------- Auth ---------- */
export { login, checkDb, checkRedis, initSystem, ping } from './auth'

/* ---------- User ---------- */
export { getProfile, updateProfile, changePassword, logout, MOCK_USER } from './user'

/* ---------- Dashboard ---------- */
export { getDashboard, getTokenTrend, MOCK_DASHBOARD, generateTrendData, TREND_7, TREND_30, TREND_90 } from './dashboard'

/* ---------- Chat / Sessions ---------- */
export {
  getSessions,
  getSessionDetail,
  updateSession,
  deleteSession,
  getSessionMessages,
  createChat,
  stopChat,
  compressChat,
  getCompressStatus,
  createShare,
  getShare,
  deleteShare,
  getMyShares,
  getModels as getChatModels,
  getMcpEndpoints as getChatMcpEndpoints,
  getSkills as getChatSkills,
  getPersonas as getChatPersonas,
  uploadFile as uploadChatFile,
  MOCK_MODELS as CHAT_MODELS,
  MOCK_MCP_ENDPOINTS as CHAT_MCP_ENDPOINTS,
  MOCK_SKILLS as CHAT_SKILLS,
  MOCK_PERSONAS as CHAT_PERSONAS,
  MOCK_SESSIONS,
  MOCK_STREAM_REPLY,
  reasoningSegments,
  streamingToolCall1,
  streamingToolCall2,
  streamingToolCall3,
} from './chat'
export type { ToolCall, RichMessage, RichSession } from './chat'

/* ---------- Personas ---------- */
export {
  getPersonas,
  getPersonaDetail,
  createPersona,
  updatePersona,
  deletePersona,
  getTemplates,
  getTemplateDetail,
  getTemplateCategories,
  getPersonaCategories,
  getModels as getPersonaModels,
  getMcps as getPersonaMcps,
  getSkills as getPersonaSkills,
  TEMPLATES,
  TEMPLATE_CATEGORIES,
  PERSONA_MODELS,
  PERSONA_MCPS,
  PERSONA_SKILLS,
  PERSONA_CATEGORIES,
} from './persona'

/* ---------- Skills ---------- */
export {
  getSimpleSkills,
  getMarketSkills,
  getMarketSkillDetail,
  getUserSkills,
  getUserSkillDetail,
  updateUserSkill,
  deleteUserSkill,
  refreshSkills,
  installSkill,
  retryInstall,
  getSkillCategories,
  generateSkillMd,
} from './skill'

/* ---------- Provider ---------- */
export { getAvailableModels, MOCK_MODELS } from './provider'


/* ---------- MCP ---------- */
export { getMcpEndpoints, MOCK_MCP_ENDPOINTS } from './mcp'

/* ---------- File Manager ---------- */
export {
  listFiles,
  mkdir,
  rename,
  deleteFiles,
  compress,
  decompress,
  getUsage,
} from './file-manager'

/* ---------- Admin ---------- */
export * as adminUsers from './admin/users'
export * as adminModels from './admin/models'
export * as adminMcp from './admin/mcp'
export * as adminSkills from './admin/skills'
export * as adminPersonas from './admin/personas'
export * as adminSystem from './admin/system'
