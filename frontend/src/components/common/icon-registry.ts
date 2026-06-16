import type { LucideIcon } from 'lucide-react'
import { t } from '@/i18n'
import {
  // 通用 General
  Bot, Star, Sparkles, Zap, Target, Compass, Gauge, ShieldCheck, Eye, Fingerprint, KeyRound, Crown,
  Puzzle, Settings, Rocket, LayoutGrid,
  // 开发 Development
  Code, Terminal, GitBranch, GitCommit, Bug, Wrench, Hammer, Package, Blocks, Webhook, Cpu, Binary,
  GitPullRequest, GitMerge, Code2, SquareTerminal,
  // 沟通 Communication
  MessageSquare, MessageCircle, Mail, Globe, Languages, Phone, Megaphone, Radio, Bell, AtSign, Share2, Send,
  Inbox, Reply, Forward, MessagesSquare,
  // 数据 Data
  Database, BarChart, BarChart3, Table2, PieChart, TrendingUp, Activity, Sigma, Hash, Scan, FileSearch, Filter,
  LineChart, CircleDot, Crosshair, ArrowUpDown,
  // 文件 Files
  FileText, FileCode, File, FolderOpen, Clipboard, Archive, BookOpen, NotebookPen, FileCheck, Files, FileSpreadsheet, Paperclip,
  Folder, FilePlus, Download, FileArchive,
  // 运维 Infrastructure
  Server, Cloud, Container, HardDrive, Network, Shield, Lock, Radar, ShieldAlert, Workflow, CloudCog, Router,
  Monitor, Plug, Cog, GlobeLock,
  // 创意 Creative
  Palette, PenLine, Pencil, Lightbulb, Brush, Music, Headphones, Camera, Image, Film, WandSparkles, Frame,
  Quote, Feather, Stamp, Ruler,
  // 商业 Business
  Calculator, GraduationCap, Heart, Briefcase, Wallet, Receipt, Scale, Landmark, ChartSpline, BadgeDollarSign, Handshake, UserCheck,
  Building2, Ticket, Flag, Trophy,
  // 写作 Writing
  Text, TextCursor, FilePen, BookMarked, BookCopy, FileSignature, Heading, Type, Bold, Italic, Underline, SpellCheck,
  CaseSensitive, CaseUpper, CaseLower, TextQuote,
  // 金融 Finance
  Banknote, Coins, DollarSign, TrendingDown, CircleDollarSign, CreditCard, PiggyBank, Percent, CircleEqual, ArrowDownRight, ArrowUpRight, Split,
  WalletCards, Sheet, BadgeCent, BadgeEuro,
  // 社交 Social
  Users, UserPlus, UserMinus, UserX, UserRound, UserRoundPlus, UserRoundCheck, UserRoundSearch, ThumbsUp, ThumbsDown, Link, Unlink,
  Rss, Contact, ContactRound, UsersRound,
  // 时间 Time
  Clock, Calendar, CalendarDays, CalendarRange, CalendarCheck, CalendarClock, AlarmClock, Timer, TimerReset, History, Hourglass, Watch,
  BellRing, CalendarPlus, CalendarOff, ClockArrowUp,
} from 'lucide-react'

/* ============================================
   Icon Name Type
   ============================================ */

export type IconName =
  // 通用 General
  | 'bot' | 'star' | 'sparkles' | 'zap' | 'target' | 'compass' | 'gauge' | 'shield-check' | 'eye' | 'fingerprint' | 'key-round' | 'crown'
  | 'puzzle' | 'settings' | 'rocket' | 'layout-grid'
  // 开发 Development
  | 'code' | 'terminal' | 'git-branch' | 'git-commit' | 'bug' | 'wrench' | 'hammer' | 'package' | 'blocks' | 'webhook' | 'cpu' | 'binary'
  | 'git-pull-request' | 'git-merge' | 'code-2' | 'square-terminal'
  // 沟通 Communication
  | 'message-square' | 'message-circle' | 'mail' | 'globe' | 'languages' | 'phone' | 'megaphone' | 'radio' | 'bell' | 'at-sign' | 'share-2' | 'send'
  | 'inbox' | 'reply' | 'forward' | 'messages-square'
  // 数据 Data
  | 'database' | 'bar-chart' | 'bar-chart-3' | 'table-2' | 'pie-chart' | 'trending-up' | 'activity' | 'sigma' | 'hash' | 'scan' | 'file-search' | 'filter'
  | 'line-chart' | 'circle-dot' | 'crosshair' | 'arrow-up-down'
  // 文件 Files
  | 'file-text' | 'file-code' | 'file' | 'folder-open' | 'clipboard' | 'archive' | 'book-open' | 'notebook-pen' | 'file-check' | 'files' | 'file-spreadsheet' | 'paperclip'
  | 'folder' | 'file-plus' | 'download' | 'file-archive'
  // 运维 Infrastructure
  | 'server' | 'cloud' | 'container' | 'hard-drive' | 'network' | 'shield' | 'lock' | 'radar' | 'shield-alert' | 'workflow' | 'cloud-cog' | 'router'
  | 'monitor' | 'plug' | 'cog' | 'globe-lock'
  // 创意 Creative
  | 'palette' | 'pen-line' | 'pencil' | 'lightbulb' | 'brush' | 'music' | 'headphones' | 'camera' | 'image' | 'film' | 'wand-sparkles' | 'frame'
  | 'quote' | 'feather' | 'stamp' | 'ruler'
  // 商业 Business
  | 'calculator' | 'graduation-cap' | 'heart' | 'briefcase' | 'wallet' | 'receipt' | 'scale' | 'landmark' | 'chart-spline' | 'badge-dollar-sign' | 'handshake' | 'user-check'
  | 'building-2' | 'ticket' | 'flag' | 'trophy'
  // 写作 Writing
  | 'text' | 'text-cursor' | 'file-pen' | 'book-marked' | 'book-copy' | 'file-signature' | 'heading' | 'type' | 'bold' | 'italic' | 'underline' | 'spell-check'
  | 'case-sensitive' | 'case-upper' | 'case-lower' | 'text-quote'
  // 金融 Finance
  | 'banknote' | 'coins' | 'dollar-sign' | 'trending-down' | 'circle-dollar-sign' | 'credit-card' | 'piggy-bank' | 'percent' | 'circle-equal' | 'arrow-down-right' | 'arrow-up-right' | 'split'
  | 'wallet-cards' | 'sheet' | 'badge-cent' | 'badge-euro'
  // 社交 Social
  | 'users' | 'user-plus' | 'user-minus' | 'user-x' | 'user-round' | 'user-round-plus' | 'user-round-check' | 'user-round-search' | 'thumbs-up' | 'thumbs-down' | 'link' | 'unlink'
  | 'rss' | 'contact' | 'contact-round' | 'users-round'
  // 时间 Time
  | 'clock' | 'calendar' | 'calendar-days' | 'calendar-range' | 'calendar-check' | 'calendar-clock' | 'alarm-clock' | 'timer' | 'timer-reset' | 'history' | 'hourglass' | 'watch'
  | 'bell-ring' | 'calendar-plus' | 'calendar-off' | 'clock-arrow-up'

/* ============================================
   Category Definition
   ============================================ */

export interface IconCategory {
  key: string
  label: string
  icons: IconName[]
}

export const ICON_CATEGORIES: IconCategory[] = [
  {
    key: 'general',
    label: t('icon.category.general'),
    icons: ['bot', 'star', 'sparkles', 'zap', 'target', 'compass', 'gauge', 'shield-check', 'eye', 'fingerprint', 'key-round', 'crown', 'puzzle', 'settings', 'rocket', 'layout-grid'],
  },
  {
    key: 'development',
    label: t('icon.category.development'),
    icons: ['code', 'terminal', 'git-branch', 'git-commit', 'bug', 'wrench', 'hammer', 'package', 'blocks', 'webhook', 'cpu', 'binary', 'git-pull-request', 'git-merge', 'code-2', 'square-terminal'],
  },
  {
    key: 'communication',
    label: t('icon.category.communication'),
    icons: ['message-square', 'message-circle', 'mail', 'globe', 'languages', 'phone', 'megaphone', 'radio', 'bell', 'at-sign', 'share-2', 'send', 'inbox', 'reply', 'forward', 'messages-square'],
  },
  {
    key: 'data',
    label: t('icon.category.data'),
    icons: ['database', 'bar-chart', 'bar-chart-3', 'table-2', 'pie-chart', 'trending-up', 'activity', 'sigma', 'hash', 'scan', 'file-search', 'filter', 'line-chart', 'circle-dot', 'crosshair', 'arrow-up-down'],
  },
  {
    key: 'files',
    label: t('icon.category.files'),
    icons: ['file-text', 'file-code', 'file', 'folder-open', 'clipboard', 'archive', 'book-open', 'notebook-pen', 'file-check', 'files', 'file-spreadsheet', 'paperclip', 'folder', 'file-plus', 'download', 'file-archive'],
  },
  {
    key: 'infrastructure',
    label: t('icon.category.infrastructure'),
    icons: ['server', 'cloud', 'container', 'hard-drive', 'network', 'shield', 'lock', 'radar', 'shield-alert', 'workflow', 'cloud-cog', 'router', 'monitor', 'plug', 'cog', 'globe-lock'],
  },
  {
    key: 'creative',
    label: t('icon.category.creative'),
    icons: ['palette', 'pen-line', 'pencil', 'lightbulb', 'brush', 'music', 'headphones', 'camera', 'image', 'film', 'wand-sparkles', 'frame', 'quote', 'feather', 'stamp', 'ruler'],
  },
  {
    key: 'business',
    label: t('icon.category.business'),
    icons: ['calculator', 'graduation-cap', 'heart', 'briefcase', 'wallet', 'receipt', 'scale', 'landmark', 'chart-spline', 'badge-dollar-sign', 'handshake', 'user-check', 'building-2', 'ticket', 'flag', 'trophy'],
  },
  {
    key: 'writing',
    label: t('icon.category.writing'),
    icons: ['text', 'text-cursor', 'file-pen', 'book-marked', 'book-copy', 'file-signature', 'heading', 'type', 'bold', 'italic', 'underline', 'spell-check', 'case-sensitive', 'case-upper', 'case-lower', 'text-quote'],
  },
  {
    key: 'finance',
    label: t('icon.category.finance'),
    icons: ['banknote', 'coins', 'dollar-sign', 'trending-down', 'circle-dollar-sign', 'credit-card', 'piggy-bank', 'percent', 'circle-equal', 'arrow-down-right', 'arrow-up-right', 'split', 'wallet-cards', 'sheet', 'badge-cent', 'badge-euro'],
  },
  {
    key: 'social',
    label: t('icon.category.social'),
    icons: ['users', 'user-plus', 'user-minus', 'user-x', 'user-round', 'user-round-plus', 'user-round-check', 'user-round-search', 'thumbs-up', 'thumbs-down', 'link', 'unlink', 'rss', 'contact', 'contact-round', 'users-round'],
  },
  {
    key: 'time',
    label: t('icon.category.time'),
    icons: ['clock', 'calendar', 'calendar-days', 'calendar-range', 'calendar-check', 'calendar-clock', 'alarm-clock', 'timer', 'timer-reset', 'history', 'hourglass', 'watch', 'bell-ring', 'calendar-plus', 'calendar-off', 'clock-arrow-up'],
  },
]

/* ============================================
   Icon Registry (name → component)
   ============================================ */

const ICON_MAP: Record<IconName, LucideIcon> = {
  // 通用 General
  bot: Bot,
  star: Star,
  sparkles: Sparkles,
  zap: Zap,
  target: Target,
  compass: Compass,
  gauge: Gauge,
  'shield-check': ShieldCheck,
  eye: Eye,
  fingerprint: Fingerprint,
  'key-round': KeyRound,
  crown: Crown,
  puzzle: Puzzle,
  settings: Settings,
  rocket: Rocket,
  'layout-grid': LayoutGrid,
  // 开发 Development
  code: Code,
  terminal: Terminal,
  'git-branch': GitBranch,
  'git-commit': GitCommit,
  bug: Bug,
  wrench: Wrench,
  hammer: Hammer,
  package: Package,
  blocks: Blocks,
  webhook: Webhook,
  cpu: Cpu,
  binary: Binary,
  'git-pull-request': GitPullRequest,
  'git-merge': GitMerge,
  'code-2': Code2,
  'square-terminal': SquareTerminal,
  // 沟通 Communication
  'message-square': MessageSquare,
  'message-circle': MessageCircle,
  mail: Mail,
  globe: Globe,
  languages: Languages,
  phone: Phone,
  megaphone: Megaphone,
  radio: Radio,
  bell: Bell,
  'at-sign': AtSign,
  'share-2': Share2,
  send: Send,
  inbox: Inbox,
  reply: Reply,
  forward: Forward,
  'messages-square': MessagesSquare,
  // 数据 Data
  database: Database,
  'bar-chart': BarChart,
  'bar-chart-3': BarChart3,
  'table-2': Table2,
  'pie-chart': PieChart,
  'trending-up': TrendingUp,
  activity: Activity,
  sigma: Sigma,
  hash: Hash,
  scan: Scan,
  'file-search': FileSearch,
  filter: Filter,
  'line-chart': LineChart,
  'circle-dot': CircleDot,
  crosshair: Crosshair,
  'arrow-up-down': ArrowUpDown,
  // 文件 Files
  'file-text': FileText,
  'file-code': FileCode,
  file: File,
  'folder-open': FolderOpen,
  clipboard: Clipboard,
  archive: Archive,
  'book-open': BookOpen,
  'notebook-pen': NotebookPen,
  'file-check': FileCheck,
  files: Files,
  'file-spreadsheet': FileSpreadsheet,
  paperclip: Paperclip,
  folder: Folder,
  'file-plus': FilePlus,
  download: Download,
  'file-archive': FileArchive,
  // 运维 Infrastructure
  server: Server,
  cloud: Cloud,
  container: Container,
  'hard-drive': HardDrive,
  network: Network,
  shield: Shield,
  lock: Lock,
  radar: Radar,
  'shield-alert': ShieldAlert,
  workflow: Workflow,
  'cloud-cog': CloudCog,
  router: Router,
  monitor: Monitor,
  plug: Plug,
  cog: Cog,
  'globe-lock': GlobeLock,
  // 创意 Creative
  palette: Palette,
  'pen-line': PenLine,
  pencil: Pencil,
  lightbulb: Lightbulb,
  brush: Brush,
  music: Music,
  headphones: Headphones,
  camera: Camera,
  image: Image,
  film: Film,
  'wand-sparkles': WandSparkles,
  frame: Frame,
  quote: Quote,
  feather: Feather,
  stamp: Stamp,
  ruler: Ruler,
  // 商业 Business
  calculator: Calculator,
  'graduation-cap': GraduationCap,
  heart: Heart,
  briefcase: Briefcase,
  wallet: Wallet,
  receipt: Receipt,
  scale: Scale,
  landmark: Landmark,
  'chart-spline': ChartSpline,
  'badge-dollar-sign': BadgeDollarSign,
  handshake: Handshake,
  'user-check': UserCheck,
  'building-2': Building2,
  ticket: Ticket,
  flag: Flag,
  trophy: Trophy,
  // 写作 Writing
  text: Text,
  'text-cursor': TextCursor,
  'file-pen': FilePen,
  'book-marked': BookMarked,
  'book-copy': BookCopy,
  'file-signature': FileSignature,
  heading: Heading,
  type: Type,
  bold: Bold,
  italic: Italic,
  underline: Underline,
  'spell-check': SpellCheck,
  'case-sensitive': CaseSensitive,
  'case-upper': CaseUpper,
  'case-lower': CaseLower,
  'text-quote': TextQuote,
  // 金融 Finance
  banknote: Banknote,
  coins: Coins,
  'dollar-sign': DollarSign,
  'trending-down': TrendingDown,
  'circle-dollar-sign': CircleDollarSign,
  'credit-card': CreditCard,
  'piggy-bank': PiggyBank,
  percent: Percent,
  'circle-equal': CircleEqual,
  'arrow-down-right': ArrowDownRight,
  'arrow-up-right': ArrowUpRight,
  split: Split,
  'wallet-cards': WalletCards,
  sheet: Sheet,
  'badge-cent': BadgeCent,
  'badge-euro': BadgeEuro,
  // 社交 Social
  users: Users,
  'user-plus': UserPlus,
  'user-minus': UserMinus,
  'user-x': UserX,
  'user-round': UserRound,
  'user-round-plus': UserRoundPlus,
  'user-round-check': UserRoundCheck,
  'user-round-search': UserRoundSearch,
  'thumbs-up': ThumbsUp,
  'thumbs-down': ThumbsDown,
  link: Link,
  unlink: Unlink,
  rss: Rss,
  contact: Contact,
  'contact-round': ContactRound,
  'users-round': UsersRound,
  // 时间 Time
  clock: Clock,
  calendar: Calendar,
  'calendar-days': CalendarDays,
  'calendar-range': CalendarRange,
  'calendar-check': CalendarCheck,
  'calendar-clock': CalendarClock,
  'alarm-clock': AlarmClock,
  timer: Timer,
  'timer-reset': TimerReset,
  history: History,
  hourglass: Hourglass,
  watch: Watch,
  'bell-ring': BellRing,
  'calendar-plus': CalendarPlus,
  'calendar-off': CalendarOff,
  'clock-arrow-up': ClockArrowUp,
}

/** Look up a Lucide component by icon name. Returns Bot as fallback. */
export function getIconComponent(name: string): LucideIcon {
  return (ICON_MAP as Record<string, LucideIcon>)[name] ?? Bot
}

/** All icon names in a flat array. */
export const ALL_ICON_NAMES: IconName[] = ICON_CATEGORIES.flatMap(c => c.icons)

/** Default icon name used as fallback. */
export const DEFAULT_ICON: IconName = 'bot'
