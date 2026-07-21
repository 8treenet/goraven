import { useState, useCallback, useRef, useEffect, useMemo } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import { Popover } from 'radix-ui'
import {
  PanelLeftClose,
  PanelLeftOpen,
  MessageSquarePlus,
  LayoutDashboard,
  FolderOpen,
  Puzzle,
  Palette,
  Users,
  Brain,
  Plug,
  Settings,
  Info,
  User,
  LogOut,
  ChevronDown,
  ChevronRight,
  Ellipsis,
  Pencil,
  Archive,
  Check,
  X,
  Share2,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useSidebarStore } from '@/stores/sidebar-store'
import { useUserStore } from '@/stores/user-store'
import type { CurrentUser } from '@/stores/user-store'
import { useChatStore } from '@/stores/chat-store'
import { useT, type TranslationKey } from '@/i18n'
import { logout } from '@/api/auth'
import { sessionsApi } from '@/api'
import type { SessionSimple } from '@/api/types'

/* ============================================
   Guest user fallback
   ============================================ */

const GUEST_USER: CurrentUser = {
  userId: '',
  username: 'guest',
  nickname: 'Guest',
  avatar: '',
  role: 0,
  email: '',
}

type TimeGroup = 'today' | '7days' | '30days'

const TIME_GROUP_ORDER: TimeGroup[] = ['today', '7days', '30days']

function getTimeGroup(lastChatTime: string): TimeGroup | null {
  const now = new Date()
  const date = new Date(lastChatTime)

  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  if (date >= todayStart) return 'today'

  const daysDiff = Math.floor((todayStart.getTime() - date.getTime()) / (1000 * 60 * 60 * 24))
  if (daysDiff < 7) return '7days'
  if (daysDiff < 30) return '30days'
  return null
}

const TIME_GROUP_I18N: Record<TimeGroup, TranslationKey> = {
  today: 'sidebar.today',
  '7days': 'sidebar.within7Days',
  '30days': 'sidebar.within30Days',
}

/* ============================================
   Component
   ============================================ */

export function Sidebar({ variant = 'desktop' }: { variant?: 'desktop' | 'mobile' }) {
  const location = useLocation()
  const collapsed = useSidebarStore((s) => s.collapsed)
  const mobileOpen = useSidebarStore((s) => s.mobileOpen)
  const closeMobile = useSidebarStore((s) => s.closeMobile)
  const adminMode = location.pathname.startsWith('/admin')
  const currentUser = useUserStore((s) => s.currentUser)
  const user = currentUser ?? GUEST_USER

  useEffect(() => {
    if (variant === 'mobile') closeMobile()
  }, [location.pathname, closeMobile, variant])

  if (variant === 'mobile') {
    return (
      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-50 flex w-64 flex-col border-r border-border bg-bg-layer-1 transition-transform duration-200 ease-out md:hidden',
          mobileOpen ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <SidebarBrand variant="mobile" />
        {!adminMode && <SidebarNewChat variant="mobile" />}
        <div className="no-scrollbar flex-1 overflow-y-auto px-2">
          {adminMode ? <AdminMenu collapsed={false} /> : <MobileUserMenu />}
        </div>
        <SidebarUserArea user={user} collapsed={false} variant={variant} />
      </aside>
    )
  }

  return (
    <aside
      className={cn(
        'flex h-screen flex-col border-r border-border bg-bg-layer-1 transition-all duration-200 ease-out',
        collapsed ? 'w-16' : 'w-64',
      )}
    >
      <SidebarBrand variant="desktop" />
      {!adminMode && <SidebarNewChat variant="desktop" />}
      <div className="no-scrollbar flex-1 overflow-y-auto px-2">
        {adminMode ? <AdminMenu /> : <UserMenu />}
      </div>
      <SidebarUserArea user={user} variant={variant} />
    </aside>
  )
}

/* ============================================
   Brand
   ============================================ */

function SidebarBrand({ variant = 'desktop' }: { variant?: 'desktop' | 'mobile' }) {
  const collapsed = useSidebarStore((s) => s.collapsed)
  const toggleCollapsed = useSidebarStore((s) => s.toggleCollapsed)
  const closeMobile = useSidebarStore((s) => s.closeMobile)
  const t = useT()

  return (
    <div className="flex h-10 items-center justify-between px-2 mb-2">
      <a
        href="https://goraven.dev"
        target="_blank"
        rel="noopener noreferrer"
        className="flex items-center gap-0 min-w-0 pl-2 transition-opacity hover:opacity-70"
      >
        <img
          src="/favicon.svg"
          alt="Raven"
          className="h-7 w-7 shrink-0 opacity-80"
        />
        {!(variant === 'desktop' && collapsed) && (
          <span className="truncate text-base font-semibold text-text-1">
            Raven
          </span>
        )}
      </a>
      {variant === 'mobile' ? (
        <button
          onClick={closeMobile}
          aria-label={t('sidebar.close')}
          className="shrink-0 text-text-muted transition-colors hover:text-text-3"
        >
          <X className="size-4" />
        </button>
      ) : (
        <button
          onClick={toggleCollapsed}
          className="shrink-0 text-interactive transition-colors hover:text-interactive-hover"
        >
          {collapsed ? (
            <PanelLeftOpen className="size-4" />
          ) : (
            <PanelLeftClose className="size-4" />
          )}
        </button>
      )}
    </div>
  )
}

/* ============================================
   New Chat Button
   ============================================ */

function SidebarNewChat({ variant = 'desktop' }: { variant?: 'desktop' | 'mobile' }) {
  const t = useT()
  const collapsed = useSidebarStore((s) => s.collapsed)
  const closeMobile = useSidebarStore((s) => s.closeMobile)
  const navigate = useNavigate()

  const handleClick = () => {
    navigate('/chat')
    if (variant === 'mobile') closeMobile()
  }

  return (
    <div className="px-2 py-2">
      <button
        onClick={handleClick}
        className={cn(
          'flex w-full items-center gap-2 rounded-md py-1.5 text-sm text-text-1 transition-colors hover:bg-bg-hover',
          (variant === 'desktop' && collapsed) ? 'justify-center px-0' : 'px-3',
        )}
      >
        <MessageSquarePlus className="size-4 shrink-0 text-amber-500" />
        {!(variant === 'desktop' && collapsed) && <span>{t('sidebar.newChat')}</span>}
      </button>
    </div>
  )
}

/* ============================================
   User Menu (user mode)
   ============================================ */

function UserMenu() {
  const t = useT()
  const collapsed = useSidebarStore((s) => s.collapsed)
  const refreshSessions = useChatStore((s) => s.refreshSessions)
  const sessionHasMore = useChatStore((s) => s.sessionHasMore)
  const loadMoreSessions = useChatStore((s) => s.loadMoreSessions)
  const sentinelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    refreshSessions()
  }, [refreshSessions])

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && sessionHasMore) {
          loadMoreSessions()
        }
      },
      { threshold: 0.1 },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [sessionHasMore, loadMoreSessions])

  const storeSessions = useChatStore((s) => s.sessions)
  const sessions: SessionSimple[] = storeSessions.map((s) => ({
    sessionId: s.id,
    title: s.title,
    status: s.status,
    personaId: s.personaId ?? 0,
    project: s.project,
    lastChatTime: s.lastChatTime,
    created: s.lastChatTime,
  }))

  const groupedSessions = useMemo(() => {
    const groups: Record<TimeGroup, SessionSimple[]> = {
      today: [], '7days': [], '30days': [],
    }
    for (const s of sessions) {
      const group = getTimeGroup(s.lastChatTime)
      if (group) groups[group].push(s)
    }
    return TIME_GROUP_ORDER
      .filter((key) => groups[key].length > 0)
      .map((key) => ({ key, sessions: groups[key] }))
  }, [sessions])

  const handleRename = useCallback(async (sessionId: string, title: string) => {
    useChatStore.setState((s) => ({
      sessions: s.sessions.map((sess) => sess.id === sessionId ? { ...sess, title } : sess),
    }))
    try {
      await sessionsApi.updateSession(sessionId, { title })
    } catch {
      refreshSessions()
    }
  }, [refreshSessions])

  const handleDelete = useCallback(async (sessionId: string) => {
    useChatStore.setState((s) => ({
      sessions: s.sessions.filter((sess) => sess.id !== sessionId),
    }))
    try {
      await sessionsApi.deleteSession(sessionId)
    } catch {
      refreshSessions()
    }
  }, [refreshSessions])

  return (
    <>
      <SidebarGroup
        label={t('sidebar.workspace')}
        groupKey="workspace"
        collapsed={collapsed}
      >
        <NavItem
          to="/dashboard"
          icon={LayoutDashboard}
          label={t('sidebar.dashboard')}
          collapsed={collapsed}
        />
        <NavItem
          to="/files"
          icon={FolderOpen}
          label={t('sidebar.files')}
          collapsed={collapsed}
        />
        <NavItem
          to="/skills"
          icon={Puzzle}
          label={t('sidebar.skills')}
          collapsed={collapsed}
        />
        <NavItem
          to="/personas"
          icon={Palette}
          label={t('sidebar.personas')}
          collapsed={collapsed}
        />
      </SidebarGroup>

      <div className="my-2 border-t border-border" />

      {collapsed ? (
        <div className="space-y-0.5 py-2">
          {sessions.map((s) => (
            <SessionItem key={s.sessionId} session={s} collapsed={collapsed} onRename={handleRename} onDelete={handleDelete} />
          ))}
          <div ref={sentinelRef} className="h-1" />
        </div>
      ) : (
        <div className="py-2 space-y-2">
          {groupedSessions.length === 0 && (
            <p className="px-2 text-xs text-text-muted">{t('sidebar.noSessions')}</p>
          )}
          {groupedSessions.map((group) => (
            <div key={group.key}>
              <p className="px-2 py-0.5 text-xs text-text-muted">{t(TIME_GROUP_I18N[group.key])}</p>
              <div className="mt-0.5 space-y-[3px]">
                {group.sessions.map((s) => (
                  <SessionItem key={s.sessionId} session={s} collapsed={collapsed} onRename={handleRename} onDelete={handleDelete} />
                ))}
              </div>
            </div>
          ))}
          <div ref={sentinelRef} className="h-1" />
        </div>
      )}
    </>
  )
}

/* ============================================
   Mobile User Menu (sessions only, no workspace nav)
   ============================================ */

function MobileUserMenu() {
  const t = useT()
  const closeMobile = useSidebarStore((s) => s.closeMobile)
  const refreshSessions = useChatStore((s) => s.refreshSessions)
  const sessionHasMore = useChatStore((s) => s.sessionHasMore)
  const loadMoreSessions = useChatStore((s) => s.loadMoreSessions)
  const sentinelRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    refreshSessions()
  }, [refreshSessions])

  useEffect(() => {
    const sentinel = sentinelRef.current
    if (!sentinel) return
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && sessionHasMore) {
          loadMoreSessions()
        }
      },
      { threshold: 0.1 },
    )
    observer.observe(sentinel)
    return () => observer.disconnect()
  }, [sessionHasMore, loadMoreSessions])

  const storeSessions = useChatStore((s) => s.sessions)
  const sessions: SessionSimple[] = storeSessions.map((s) => ({
    sessionId: s.id,
    title: s.title,
    status: s.status,
    personaId: s.personaId ?? 0,
    project: s.project,
    lastChatTime: s.lastChatTime,
    created: s.lastChatTime,
  }))

  const groupedSessions = useMemo(() => {
    const groups: Record<TimeGroup, SessionSimple[]> = {
      today: [], '7days': [], '30days': [],
    }
    for (const s of sessions) {
      const group = getTimeGroup(s.lastChatTime)
      if (group) groups[group].push(s)
    }
    return TIME_GROUP_ORDER
      .filter((key) => groups[key].length > 0)
      .map((key) => ({ key, sessions: groups[key] }))
  }, [sessions])

  const handleRename = useCallback(async (sessionId: string, title: string) => {
    useChatStore.setState((s) => ({
      sessions: s.sessions.map((sess) => sess.id === sessionId ? { ...sess, title } : sess),
    }))
    try {
      await sessionsApi.updateSession(sessionId, { title })
    } catch {
      refreshSessions()
    }
  }, [refreshSessions])

  const handleDelete = useCallback(async (sessionId: string) => {
    useChatStore.setState((s) => ({
      sessions: s.sessions.filter((sess) => sess.id !== sessionId),
    }))
    try {
      await sessionsApi.deleteSession(sessionId)
    } catch {
      refreshSessions()
    }
  }, [refreshSessions])

  return (
    <div className="py-2 space-y-2">
      {groupedSessions.length === 0 && (
        <p className="px-2 text-xs text-text-muted">{t('sidebar.noSessions')}</p>
      )}
      {groupedSessions.map((group) => (
        <div key={group.key}>
          <p className="px-2 py-0.5 text-xs text-text-muted">{t(TIME_GROUP_I18N[group.key])}</p>
          <div className="mt-0.5 space-y-[3px]">
            {group.sessions.map((s) => (
              <MobileSessionItem key={s.sessionId} session={s} onRename={handleRename} onDelete={handleDelete} onClose={closeMobile} />
            ))}
          </div>
        </div>
      ))}
      <div ref={sentinelRef} className="h-1" />
    </div>
  )
}

function MobileSessionItem({
  session,
  onRename,
  onDelete,
  onClose,
}: {
  session: SessionSimple
  onRename: (sessionId: string, title: string) => void
  onDelete: (sessionId: string) => void
  onClose: () => void
}) {
  const navigate = useNavigate()
  const handleNavigate = () => {
    navigate(`/chat/${session.sessionId}`)
    onClose()
  }
  return (
    <SessionItem
      session={session}
      collapsed={false}
      onRename={onRename}
      onDelete={onDelete}
      onNavigateOverride={handleNavigate}
    />
  )
}

/* ============================================
   Admin Menu
   ============================================ */

function AdminMenu({ collapsed: collapsedProp }: { collapsed?: boolean }) {
  const t = useT()
  const storeCollapsed = useSidebarStore((s) => s.collapsed)
  const collapsed = collapsedProp ?? storeCollapsed

  return (
    <>
      <SidebarGroup
        label={t('sidebar.overview')}
        groupKey="admin-overview"
        collapsed={collapsed}
      >
        <NavItem to="/admin" icon={LayoutDashboard} label={t('sidebar.dashboard')} collapsed={collapsed} />
        <NavItem to="/admin/systemInfo" icon={Info} label={t('sidebar.systemInfo')} collapsed={collapsed} />
      </SidebarGroup>

      <SidebarGroup
        label={t('sidebar.systemAdmin')}
        groupKey="admin-system"
        collapsed={collapsed}
      >
        <NavItem to="/admin/users" icon={Users} label={t('sidebar.users')} collapsed={collapsed} />
        <NavItem to="/admin/settings" icon={Settings} label={t('sidebar.settings')} collapsed={collapsed} />
        <NavItem to="/admin/shared-projects" icon={Share2} label={t('sidebar.sharedProjects')} collapsed={collapsed} />
      </SidebarGroup>

      <SidebarGroup
        label={t('sidebar.config')}
        groupKey="admin-config"
        collapsed={collapsed}
      >
        <NavItem to="/admin/models" icon={Brain} label={t('sidebar.models')} collapsed={collapsed} />
        <NavItem to="/admin/mcp" icon={Plug} label={t('sidebar.mcp')} collapsed={collapsed} />
        <NavItem to="/admin/skills" icon={Puzzle} label={t('sidebar.skillAdmin')} collapsed={collapsed} />
        <NavItem to="/admin/persona-templates" icon={Palette} label={t('sidebar.personaTemplates')} collapsed={collapsed} />
      </SidebarGroup>
    </>
  )
}

/* ============================================
   Collapsible Group
   ============================================ */

function SidebarGroup({
  label,
  groupKey,
  collapsed: sidebarCollapsed,
  children,
}: {
  label: string
  groupKey: string
  collapsed: boolean
  children: React.ReactNode
}) {
  const expandedGroups = useSidebarStore((s) => s.expandedGroups)
  const toggleGroup = useSidebarStore((s) => s.toggleGroup)
  const isExpanded = expandedGroups.includes(groupKey)

  if (sidebarCollapsed) {
    return <div className="space-y-0.5 py-2">{children}</div>
  }

  return (
    <div className="py-2">
      <button
        onClick={() => toggleGroup(groupKey)}
        className="flex w-full items-center gap-1 px-2 py-0.5 text-xs text-text-3 transition-colors hover:text-text-2"
      >
        {isExpanded ? (
          <ChevronDown className="size-3 shrink-0" />
        ) : (
          <ChevronRight className="size-3 shrink-0" />
        )}
        <span className="truncate">{label}</span>
      </button>
      {isExpanded && <div className="mt-0.5 space-y-0.5">{children}</div>}
    </div>
  )
}

/* ============================================
   Nav Item
   ============================================ */

function NavItem({
  to,
  icon: Icon,
  label,
  collapsed,
  indent,
}: {
  to: string
  icon: React.ComponentType<{ className?: string }>
  label: string
  collapsed: boolean
  indent?: boolean
}) {
  const navigate = useNavigate()
  const location = useLocation()
  const isActive = location.pathname === to

  return (
    <button
      onClick={() => navigate(to)}
      className={cn(
        'flex w-full items-center gap-2 rounded-md py-1.5 text-sm text-text-1 transition-colors',
        collapsed ? 'justify-center px-0' : indent ? 'pl-7 pr-2' : 'px-2',
        isActive
          ? 'bg-interactive-soft font-medium text-interactive'
          : 'hover:bg-bg-hover',
      )}
    >
      <Icon className={cn('size-4 shrink-0', isActive ? 'text-interactive' : 'text-text-3')} />
      {!collapsed && <span className="truncate">{label}</span>}
    </button>
  )
}

/* ============================================
   Session Item
   ============================================ */

function SessionItem({
  session,
  collapsed,
  onRename,
  onDelete,
  onNavigateOverride,
}: {
  session: SessionSimple
  collapsed: boolean
  onRename: (sessionId: string, title: string) => void
  onDelete: (sessionId: string) => void
  onNavigateOverride?: () => void
}) {
  const t = useT()
  const navigate = useNavigate()
  const location = useLocation()
  const isActive = location.pathname === `/chat/${session.sessionId}`
  const [menuOpen, setMenuOpen] = useState(false)
  const [renaming, setRenaming] = useState(false)
  const [editTitle, setEditTitle] = useState(session.title)
  const inputRef = useRef<HTMLInputElement>(null)

  const handleStartRename = useCallback(() => {
    setMenuOpen(false)
    setRenaming(true)
    setEditTitle(session.title)
    setTimeout(() => inputRef.current?.focus(), 0)
    setTimeout(() => inputRef.current?.select(), 10)
  }, [session.title])

  const handleConfirmRename = useCallback(() => {
    const trimmed = editTitle.trim()
    if (trimmed && trimmed !== session.title) {
      onRename(session.sessionId, trimmed)
    }
    setRenaming(false)
  }, [editTitle, session.sessionId, session.title, onRename])

  const handleCancelRename = useCallback(() => {
    setRenaming(false)
  }, [])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') handleConfirmRename()
      if (e.key === 'Escape') handleCancelRename()
    },
    [handleConfirmRename, handleCancelRename],
  )

  const handleDelete = useCallback(() => {
    setMenuOpen(false)
    onDelete(session.sessionId)
  }, [session.sessionId, onDelete])

  if (collapsed) {
    return (
      <button
        onClick={onNavigateOverride ?? (() => navigate(`/chat/${session.sessionId}`))}
        title={session.title}
        className={cn(
          'flex w-full items-center justify-center rounded-md py-1.5 text-sm text-text-2 transition-colors hover:bg-bg-hover',
          isActive && 'bg-interactive-soft font-medium text-interactive',
        )}
      >
        {session.title.charAt(0)}
      </button>
    )
  }

  if (renaming) {
    return (
      <div className="flex items-center gap-1 rounded-md py-1.5 px-2 bg-bg-layer-2">
        <input
          ref={inputRef}
          value={editTitle}
          onChange={(e) => setEditTitle(e.target.value)}
          onKeyDown={handleKeyDown}
          onBlur={handleCancelRename}
          className="min-w-0 flex-1 bg-transparent text-sm text-text-1 outline-none"
        />
        <button
          onClick={handleConfirmRename}
          className="shrink-0 text-text-3 hover:text-text-1"
        >
          <Check className="size-3.5" />
        </button>
        <button
          onClick={handleCancelRename}
          className="shrink-0 text-text-3 hover:text-text-1"
        >
          <X className="size-3.5" />
        </button>
      </div>
    )
  }

  return (
    <div
      className={cn(
        'group flex items-center rounded-md text-text-1 transition-colors',
        isActive
          ? 'bg-interactive-soft text-interactive'
          : 'hover:bg-bg-hover',
      )}
    >
      <button
        onClick={onNavigateOverride ?? (() => navigate(`/chat/${session.sessionId}`))}
        className="flex min-w-0 flex-1 py-1.5 pl-3 pr-2"
      >
        <span className="overflow-hidden text-nowrap text-sm">{session.title}</span>
      </button>

      <Popover.Root open={menuOpen} onOpenChange={setMenuOpen}>
        <Popover.Trigger asChild>
          <button
            onClick={(e) => e.stopPropagation()}
            className={cn(
              'shrink-0 py-1.5 pl-1 pr-2 text-text-3 transition-opacity hover:text-text-1',
              menuOpen ? 'opacity-100' : 'opacity-0 group-hover:opacity-100',
            )}
          >
            <Ellipsis className="size-3.5" />
          </button>
        </Popover.Trigger>

        <Popover.Portal>
          <Popover.Content
            side="right"
            sideOffset={4}
            className="z-50 w-36 rounded-lg border border-border bg-bg-layer-2 p-1 shadow-pop outline-none"
          >
            <PopoverItem
              icon={Pencil}
              label={t('sidebar.rename')}
              onClick={handleStartRename}
            />
            <PopoverItem
              icon={Archive}
              label={t('sidebar.archive')}
              onClick={handleDelete}
            />
          </Popover.Content>
        </Popover.Portal>
      </Popover.Root>
    </div>
  )
}

/* ============================================
   User Area
   ============================================ */

function SidebarUserArea({
  user,
  collapsed: collapsedProp,
  variant = 'desktop',
}: {
  user: CurrentUser
  collapsed?: boolean
  variant?: 'desktop' | 'mobile'
}) {
  const t = useT()
  const location = useLocation()
  const storeCollapsed = useSidebarStore((s) => s.collapsed)
  const collapsed = collapsedProp ?? storeCollapsed
  const adminMode = location.pathname.startsWith('/admin')
  const isAdmin = useUserStore((s) => s.isAdmin)
  const [open, setOpen] = useState(false)
  const navigate = useNavigate()

  const handleProfile = useCallback(() => {
    setOpen(false)
    navigate('/profile')
  }, [navigate])

  const handleAdminToggle = useCallback(() => {
    setOpen(false)
    if (adminMode) {
      navigate('/chat')
    } else {
      navigate('/admin')
    }
  }, [navigate, adminMode])

  const handleLogout = useCallback(() => {
    setOpen(false)
    logout().finally(() => {
      useUserStore.getState().clearAuth()
      navigate('/login', { replace: true })
    })
  }, [navigate])

  const displayName = user.nickname || user.username
  const initial = displayName.charAt(0).toUpperCase()

  return (
    <div className="border-t border-border px-2 py-2">
      <Popover.Root open={open} onOpenChange={setOpen}>
        <Popover.Trigger asChild>
          <button
            className={cn(
              'flex w-full items-center gap-2 rounded-md py-1.5 text-sm text-text-2 transition-colors hover:bg-bg-layer-2 hover:text-text-1',
              collapsed ? 'justify-center px-0' : 'px-2',
            )}
          >
            {user.avatar ? (
              <img src={user.avatar} alt={displayName} className="h-6 w-6 shrink-0 rounded-full object-cover" />
            ) : (
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-interactive text-xs font-semibold text-white">
                {initial}
              </span>
            )}
            {!collapsed && (
              <>
                <span className="truncate text-text-2">{displayName}</span>
                <ChevronRight className="ml-auto size-3 shrink-0 text-interactive" />
              </>
            )}
          </button>
        </Popover.Trigger>

        <Popover.Portal>
          <Popover.Content
            side="top"
            sideOffset={8}
            align={collapsed ? 'start' : 'center'}
            className="z-50 w-52 rounded-lg border border-border bg-bg-layer-2 p-1 shadow-pop outline-none"
          >
            <div className="flex items-center gap-2 px-2 py-2">
              {user.avatar ? (
                <img src={user.avatar} alt={displayName} className="h-8 w-8 shrink-0 rounded-full object-cover" />
              ) : (
                <span className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-interactive text-sm font-semibold text-white">
                  {initial}
                </span>
              )}
              <div className="min-w-0">
                <p className="truncate text-sm text-text-1">{displayName}</p>
                <p className="truncate text-xs text-text-3">@{user.username}</p>
              </div>
              {isAdmin && (
                <span className="ml-auto shrink-0 rounded bg-highlight px-1.5 py-0.5 text-xs text-highlight-fg">
                  {t('sidebar.admin')}
                </span>
              )}
            </div>

            <div className="my-1 h-px bg-border" />

            <PopoverItem icon={User} label={t('sidebar.profile')} onClick={handleProfile} />

            {isAdmin && variant !== 'mobile' && (
              <PopoverItem
                icon={adminMode ? ChevronDown : Settings}
                label={adminMode ? t('sidebar.userMode') : t('sidebar.adminMode')}
                onClick={handleAdminToggle}
                highlight={adminMode}
              />
            )}

            <div className="my-1 h-px bg-border" />

            <PopoverItem
              icon={LogOut}
              label={t('sidebar.logout')}
              onClick={handleLogout}
              danger
            />
          </Popover.Content>
        </Popover.Portal>
      </Popover.Root>
    </div>
  )
}

function PopoverItem({
  icon: Icon,
  label,
  onClick,
  highlight,
  danger,
}: {
  icon: React.ComponentType<{ className?: string }>
  label: string
  onClick: () => void
  highlight?: boolean
  danger?: boolean
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors',
        highlight
          ? 'bg-bg-layer-3 text-text-1'
          : danger
            ? 'text-text-2 hover:bg-bg-layer-3 hover:text-text-3'
            : 'text-text-1 hover:bg-bg-layer-3',
      )}
    >
      <Icon className="size-4 shrink-0" />
      <span>{label}</span>
    </button>
  )
}
