import { useState, useCallback, useEffect } from 'react'
import { toast } from 'sonner'
import { ChevronLeft, ChevronRight, Crown, Loader2, UserPlus, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useT, t as translate } from '@/i18n'
import {
  listTeamProjectUsers,
  getTeamProjectMembers,
  updateTeamProjectMembers,
  updateTeamProjectAccess,
} from '@/api/team-projects'
import type { TeamProjectItem, TeamProjectUserItem } from '@/api/types'

const PAGE_SIZE = 10

interface MembersDialogProps {
  project: TeamProjectItem | null
  onClose: () => void
  onSaved: () => void
}

export function MembersDialog({ project, onClose, onSaved }: MembersDialogProps) {
  const t = useT()
  const [access, setAccess] = useState(0)
  const [creatorId, setCreatorId] = useState('')
  const [memberIds, setMemberIds] = useState<Set<string>>(new Set())
  const [originalMemberIds, setOriginalMemberIds] = useState<Set<string>>(new Set())
  const [users, setUsers] = useState<TeamProjectUserItem[]>([])
  const [userPage, setUserPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loadingMembers, setLoadingMembers] = useState(false)
  const [loadingUsers, setLoadingUsers] = useState(false)
  const [saving, setSaving] = useState(false)

  // 已知用户信息缓存（userId -> item）
  const [userMap, setUserMap] = useState<Map<string, TeamProjectUserItem>>(new Map())

  const loadUsers = useCallback((page: number) => {
    setLoadingUsers(true)
    listTeamProjectUsers(page, PAGE_SIZE)
      .then((data) => {
        setUsers(data.list || [])
        setTotalPages(Math.max(1, data.totalPage))
        setUserPage(page)
        setUserMap((prev) => {
          const next = new Map(prev)
          for (const u of data.list || []) {
            next.set(u.userId, u)
          }
          return next
        })
      })
      .catch(() => {})
      .finally(() => setLoadingUsers(false))
  }, [])

  useEffect(() => {
    if (!project) return
    setAccess(project.access ?? 0)
    setLoadingMembers(true)
    getTeamProjectMembers(project.id)
      .then((data) => {
        setCreatorId(data.creatorId)
        const ids = new Set(data.memberIds || [])
        setMemberIds(ids)
        setOriginalMemberIds(new Set(ids))
      })
      .catch((err: Error) => {
        toast.error(err.message)
      })
      .finally(() => setLoadingMembers(false))
    loadUsers(1)
  }, [project, loadUsers])

  const addMember = useCallback((userId: string) => {
    setMemberIds((prev) => new Set(prev).add(userId))
  }, [])

  const removeMember = useCallback((userId: string) => {
    setMemberIds((prev) => {
      const next = new Set(prev)
      next.delete(userId)
      return next
    })
  }, [])

  const handleSave = useCallback(() => {
    if (!project) return
    setSaving(true)

    const addUserIds: string[] = []
    const removeUserIds: string[] = []
    for (const id of memberIds) {
      if (!originalMemberIds.has(id)) addUserIds.push(id)
    }
    for (const id of originalMemberIds) {
      if (!memberIds.has(id)) removeUserIds.push(id)
    }

    const tasks: Promise<unknown>[] = []
    if (access !== project.access) {
      tasks.push(updateTeamProjectAccess(project.id, access))
    }
    if (addUserIds.length > 0 || removeUserIds.length > 0) {
      tasks.push(updateTeamProjectMembers(project.id, addUserIds, removeUserIds))
    }

    if (tasks.length === 0) {
      setSaving(false)
      onClose()
      return
    }

    Promise.all(tasks)
      .then(() => {
        toast.success(translate('files.membersUpdated'))
        onClose()
        onSaved()
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.failed'))
      })
      .finally(() => setSaving(false))
  }, [project, access, memberIds, originalMemberIds, onClose, onSaved])

  const displayName = useCallback(
    (userId: string) => {
      const u = userMap.get(userId)
      return u?.nickname || u?.username || userId
    },
    [userMap],
  )

  const memberList = Array.from(memberIds)

  return (
    <Dialog open={project !== null} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t('files.membersDialogTitle')}</DialogTitle>
          <DialogDescription>
            {t('files.membersDialogDesc').replace('{name}', project?.projectName || '')}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* 访问权限单选 */}
          <div>
            <label className="text-xs text-text-muted">{t('files.accessLabel')}</label>
            <div className="mt-2 flex flex-col gap-2">
              <label className="flex cursor-pointer items-center gap-2.5 rounded-md border border-border px-3 py-2 transition-colors hover:bg-bg-hover has-[:checked]:border-interactive has-[:checked]:bg-interactive/5">
                <input
                  type="radio"
                  name="project-access"
                  checked={access === 0}
                  onChange={() => setAccess(0)}
                  className="accent-interactive"
                />
                <span className="text-sm text-text-1">{t('files.accessAll')}</span>
                <span className="ml-auto text-xs text-text-muted">{t('files.accessAllHint')}</span>
              </label>
              <label className="flex cursor-pointer items-center gap-2.5 rounded-md border border-border px-3 py-2 transition-colors hover:bg-bg-hover has-[:checked]:border-interactive has-[:checked]:bg-interactive/5">
                <input
                  type="radio"
                  name="project-access"
                  checked={access === 1}
                  onChange={() => setAccess(1)}
                  className="accent-interactive"
                />
                <span className="text-sm text-text-1">{t('files.accessMembers')}</span>
                <span className="ml-auto text-xs text-text-muted">{t('files.accessMembersHint')}</span>
              </label>
            </div>
          </div>

          {/* 成员编辑区域（仅成员可见时显示） */}
          {access === 1 && (
            <div className="space-y-3">
              {/* 当前成员 */}
              <div>
                <label className="text-xs text-text-muted">{t('files.currentMembers')}</label>
                {loadingMembers ? (
                  <div className="flex items-center justify-center py-3">
                    <Loader2 className="size-4 animate-spin text-text-muted" />
                  </div>
                ) : (
                  <div className="mt-1.5 flex flex-wrap gap-1.5">
                    {/* 创建者 */}
                    <span className="inline-flex items-center gap-1 rounded-md border border-border bg-bg-layer-2 px-2 py-1 text-xs text-text-1">
                      <Crown className="size-3 text-highlight" />
                      {displayName(creatorId)}
                      <span className="text-text-muted">({t('files.owner')})</span>
                    </span>
                    {memberList.map((uid) => (
                      <span
                        key={uid}
                        className="inline-flex items-center gap-1 rounded-md border border-border bg-bg-layer-2 px-2 py-1 text-xs text-text-1"
                      >
                        {displayName(uid)}
                        <button
                          onClick={() => removeMember(uid)}
                          className="ml-0.5 rounded-sm text-text-3 transition-colors hover:text-destructive"
                        >
                          <X className="size-3" />
                        </button>
                      </span>
                    ))}
                    {memberList.length === 0 && (
                      <span className="text-xs text-text-muted">{t('files.noMembersHint')}</span>
                    )}
                  </div>
                )}
              </div>

              {/* 用户分页列表 */}
              <div>
                <label className="text-xs text-text-muted">{t('files.addMembers')}</label>
                <div className="mt-1.5 max-h-[240px] overflow-auto rounded-md border border-border">
                  {loadingUsers ? (
                    <div className="flex items-center justify-center py-6">
                      <Loader2 className="size-4 animate-spin text-text-muted" />
                    </div>
                  ) : users.length === 0 ? (
                    <div className="py-4 text-center text-xs text-text-muted">{t('files.noUsers')}</div>
                  ) : (
                    users.map((user) => {
                      const isCreator = user.userId === creatorId
                      const isMember = memberIds.has(user.userId)
                      const disabled = isCreator || isMember
                      return (
                        <div
                          key={user.userId}
                          className="flex items-center gap-2.5 border-b border-border px-3 py-2 last:border-b-0"
                        >
                          {user.avatar ? (
                            <img src={user.avatar} alt="" className="size-6 shrink-0 rounded-sm object-cover" />
                          ) : (
                            <div className="inline-flex size-6 shrink-0 items-center justify-center rounded-sm bg-interactive text-[11px] font-medium text-white">
                              {(user.nickname || user.username).charAt(0).toUpperCase()}
                            </div>
                          )}
                          <div className="min-w-0 flex-1">
                            <span className="block truncate text-sm text-text-1">
                              {user.nickname || user.username}
                            </span>
                            {user.nickname && (
                              <span className="block truncate text-xs text-text-muted">@{user.username}</span>
                            )}
                          </div>
                          {isCreator ? (
                            <span className="shrink-0 text-xs text-text-muted">{t('files.owner')}</span>
                          ) : disabled ? (
                            <span className="shrink-0 text-xs text-interactive">{t('files.joined')}</span>
                          ) : (
                            <button
                              onClick={() => addMember(user.userId)}
                              className="inline-flex shrink-0 items-center gap-1 rounded-md px-2 py-1 text-xs text-interactive transition-colors hover:bg-interactive/10"
                            >
                              <UserPlus className="size-3" />
                              {t('common.add')}
                            </button>
                          )}
                        </div>
                      )
                    })
                  )}
                </div>

                {/* 分页 */}
                {totalPages > 1 && (
                  <div className="mt-2 flex items-center justify-center gap-2">
                    <button
                      onClick={() => loadUsers(userPage - 1)}
                      disabled={userPage <= 1 || loadingUsers}
                      className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
                    >
                      <ChevronLeft className="size-4" />
                    </button>
                    <span className="text-xs text-text-muted tabular-nums">
                      {userPage} / {totalPages}
                    </span>
                    <button
                      onClick={() => loadUsers(userPage + 1)}
                      disabled={userPage >= totalPages || loadingUsers}
                      className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
                    >
                      <ChevronRight className="size-4" />
                    </button>
                  </div>
                )}
              </div>
            </div>
          )}

          <div className={cn('flex justify-end gap-2', access === 1 && 'pt-1')}>
            <Button variant="ghost" size="default" onClick={onClose}>
              {t('common.cancel')}
            </Button>
            <Button variant="default" size="default" onClick={handleSave} disabled={saving}>
              {saving ? <Loader2 className="size-4 animate-spin" /> : t('common.save')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
