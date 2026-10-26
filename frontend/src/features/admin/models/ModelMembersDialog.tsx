import { useState, useCallback, useEffect } from 'react'
import { toast } from 'sonner'
import { ChevronLeft, ChevronRight, Loader2, UserPlus, X } from 'lucide-react'
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
import { adminModelsApi, adminUsersApi } from '@/api'
import type { AdminModelItem, AdminUserItem } from '@/api'

const PAGE_SIZE = 10

interface ModelMembersDialogProps {
  model: AdminModelItem | null
  onClose: () => void
  onSaved: () => void
}

export function ModelMembersDialog({ model, onClose, onSaved }: ModelMembersDialogProps) {
  const t = useT()
  const [access, setAccess] = useState(0)
  const [memberIds, setMemberIds] = useState<Set<string>>(new Set())
  const [originalMemberIds, setOriginalMemberIds] = useState<Set<string>>(new Set())
  const [users, setUsers] = useState<AdminUserItem[]>([])
  const [userPage, setUserPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)
  const [loadingMembers, setLoadingMembers] = useState(false)
  const [loadingUsers, setLoadingUsers] = useState(false)
  const [saving, setSaving] = useState(false)

  // 已知用户信息缓存（userId -> item）
  const [userMap, setUserMap] = useState<Map<string, AdminUserItem>>(new Map())

  const loadUsers = useCallback((page: number) => {
    setLoadingUsers(true)
    adminUsersApi.getUsers({ page, pageSize: PAGE_SIZE })
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
    if (!model) return
    setAccess(model.access ?? 0)
    setLoadingMembers(true)
    adminModelsApi.getModelMembers(model.aiModelId)
      .then((data) => {
        const ids = new Set(data.memberIds || [])
        setMemberIds(ids)
        setOriginalMemberIds(new Set(ids))
      })
      .catch((err: Error) => {
        toast.error(err.message)
      })
      .finally(() => setLoadingMembers(false))
    loadUsers(1)
  }, [model, loadUsers])

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
    if (!model) return
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
    if (access !== model.access) {
      tasks.push(adminModelsApi.updateModelAccess(model.aiModelId, access))
    }
    if (addUserIds.length > 0 || removeUserIds.length > 0) {
      tasks.push(adminModelsApi.updateModelMembers(model.aiModelId, addUserIds, removeUserIds))
    }

    if (tasks.length === 0) {
      setSaving(false)
      onClose()
      return
    }

    Promise.all(tasks)
      .then(() => {
        toast.success(translate('adminModels.membersUpdated'))
        onClose()
        onSaved()
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.failed'))
      })
      .finally(() => setSaving(false))
  }, [model, access, memberIds, originalMemberIds, onClose, onSaved])

  const displayName = useCallback(
    (userId: string) => {
      const u = userMap.get(userId)
      return u?.nickname || u?.username || userId
    },
    [userMap],
  )

  const memberList = Array.from(memberIds)

  return (
    <Dialog open={model !== null} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t('adminModels.membersDialogTitle')}</DialogTitle>
          <DialogDescription>
            {t('adminModels.membersDialogDesc').replace('{name}', model?.displayName || '')}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {/* 访问权限单选 */}
          <div>
            <label className="text-xs text-text-muted">{t('adminModels.accessLabel')}</label>
            <div className="mt-2 flex flex-col gap-2">
              <label className="flex cursor-pointer items-center gap-2.5 rounded-md border border-border px-3 py-2 transition-colors hover:bg-bg-hover has-[:checked]:border-interactive has-[:checked]:bg-interactive/5">
                <input
                  type="radio"
                  name="model-access"
                  checked={access === 0}
                  onChange={() => setAccess(0)}
                  className="accent-interactive"
                />
                <span className="text-sm text-text-1">{t('adminModels.accessAll')}</span>
                <span className="ml-auto text-xs text-text-muted">{t('adminModels.accessAllHint')}</span>
              </label>
              <label className="flex cursor-pointer items-center gap-2.5 rounded-md border border-border px-3 py-2 transition-colors hover:bg-bg-hover has-[:checked]:border-interactive has-[:checked]:bg-interactive/5">
                <input
                  type="radio"
                  name="model-access"
                  checked={access === 1}
                  onChange={() => setAccess(1)}
                  className="accent-interactive"
                />
                <span className="text-sm text-text-1">{t('adminModels.accessMembers')}</span>
                <span className="ml-auto text-xs text-text-muted">{t('adminModels.accessMembersHint')}</span>
              </label>
            </div>
          </div>

          {/* 成员编辑区域（仅成员可见时显示） */}
          {access === 1 && (
            <div className="space-y-3">
              {/* 当前成员 */}
              <div>
                <label className="text-xs text-text-muted">{t('adminModels.currentMembers')}</label>
                {loadingMembers ? (
                  <div className="flex items-center justify-center py-3">
                    <Loader2 className="size-4 animate-spin text-text-muted" />
                  </div>
                ) : (
                  <div className="mt-1.5 flex flex-wrap gap-1.5">
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
                      <span className="text-xs text-text-muted">{t('adminModels.noMembersHint')}</span>
                    )}
                  </div>
                )}
              </div>

              {/* 用户分页列表 */}
              <div>
                <label className="text-xs text-text-muted">{t('adminModels.addMembers')}</label>
                <div className="mt-1.5 max-h-[240px] overflow-auto rounded-md border border-border">
                  {loadingUsers ? (
                    <div className="flex items-center justify-center py-6">
                      <Loader2 className="size-4 animate-spin text-text-muted" />
                    </div>
                  ) : users.length === 0 ? (
                    <div className="py-4 text-center text-xs text-text-muted">{t('adminModels.noUsers')}</div>
                  ) : (
                    users.map((user) => {
                      const isMember = memberIds.has(user.userId)
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
                          {isMember ? (
                            <span className="shrink-0 text-xs text-interactive">{t('adminModels.joined')}</span>
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
