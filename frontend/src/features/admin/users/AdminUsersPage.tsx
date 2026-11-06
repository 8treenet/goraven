import { useState, useCallback, useEffect, useRef, useMemo } from 'react'
import { toast } from 'sonner'
import {
  Search,
  X,
  Plus,
  Pencil,
  Trash2,
  KeyRound,
  ChevronLeft,
  ChevronRight,
  Shield,
  User,
  Loader2,
  Check,
} from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { AppDrawer } from '@/components/patterns/AppDrawer'
import { EmptyState } from '@/components/patterns/EmptyState'
import { ErrorState } from '@/components/patterns/ErrorState'
import { PageHeader } from '@/components/patterns/PageHeader'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableHead,
  TableCell,
} from '@/components/ui/table'
import { cn } from '@/lib/utils'
import { useT, t as translate } from '@/i18n'
import { adminUsersApi } from '@/api'
import type { AdminUserItem } from '@/api'
import { useUserStore } from '@/stores/user-store'

/* ============================================
   Types
   ============================================ */

interface UserFormData {
  username: string
  password: string
  nickname: string
  role: 0 | 1
}

interface EditFormData {
  nickname: string
  email: string
  role: 0 | 1
  status: 0 | 1
}

type DrawerMode = 'add' | 'edit'
type DialogMode = 'delete' | 'resetPassword' | null

type PageState = 'loading' | 'data' | 'empty' | 'error'

/* ============================================
   Helpers
   ============================================ */

function formatDate(iso: string | null): string {
  if (!iso) return '—'
  const d = new Date(iso)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

function formatRelative(iso: string | null): string {
  if (!iso) return '—'
  const diff = Date.now() - new Date(iso).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return translate('files.justNow')
  if (mins < 60) return `${mins}${translate('files.minutesAgo')}`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}${translate('files.hoursAgo')}`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}${translate('files.daysAgo')}`
  return formatDate(iso)
}

function getRoleLabel(role: number): string {
  return role === 1 ? translate('adminUsers.admin') : translate('adminUsers.regularUser')
}

function Avatar({ username, avatar }: { username: string; avatar: string }) {
  if (avatar) {
    return (
      <img
        src={avatar}
        alt={username}
        className="size-7 shrink-0 rounded-sm object-cover"
      />
    )
  }
  return (
    <div className="inline-flex size-7 shrink-0 items-center justify-center rounded-sm bg-interactive text-xs font-medium text-bg-base">
      {username.charAt(0).toUpperCase()}
    </div>
  )
}

/* ============================================
   Status Toggle
   ============================================ */

function StatusToggle({
  checked,
  disabled,
  onChange,
}: {
  checked: boolean
  disabled: boolean
  onChange: (v: boolean) => void
}) {
  return (
    <label
      className={cn(
        'relative inline-flex cursor-pointer items-center',
        disabled && 'cursor-not-allowed opacity-40',
      )}
    >
      <input
        type="checkbox"
        className="peer sr-only"
        checked={checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
      />
      <div
        className={cn(
          'h-4 w-8 rounded-full border transition-colors',
          'peer-checked:border-success peer-checked:bg-success',
          'border-border-strong bg-bg-layer-2',
          !disabled && 'peer-hover:border-text-2',
        )}
      />
      <div
        className={cn(
          'absolute left-0.5 top-0.5 size-3 rounded-full bg-bg-base shadow-sm transition-transform',
          'peer-checked:translate-x-4',
        )}
      />
    </label>
  )
}

/* ============================================
   Drawer
   ============================================ */

/* ============================================
   Add / Edit Form Drawers
   ============================================ */

function AddUserDrawer({
  open,
  onClose,
  onSave,
  existingUsernames,
}: {
  open: boolean
  onClose: () => void
  onSave: (data: UserFormData) => void
  existingUsernames: string[]
}) {
  const t = useT()
  const [form, setForm] = useState<UserFormData>({ username: '', password: '', nickname: '', role: 0 })
  const [usernameStatus, setUsernameStatus] = useState<'idle' | 'checking' | 'ok' | 'exists'>('idle')
  const debounceRef = useRef<ReturnType<typeof setTimeout>>()

  useEffect(() => {
    if (!open) {
      setForm({ username: '', password: '', nickname: '', role: 0 })
      setUsernameStatus('idle')
    }
  }, [open])

  const usernameFormatError =
    form.username.length > 0 && (
      !/^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$/.test(form.username) ||
      form.username.length < 8
    )

  const isUsernameShapeValid = (val: string) =>
    val.length >= 8 && val.length <= 16 && /^[a-zA-Z0-9][a-zA-Z0-9_-]*[a-zA-Z0-9]$/.test(val)

  const checkUsername = useCallback(
    (val: string) => {
      if (debounceRef.current) clearTimeout(debounceRef.current)
      if (!isUsernameShapeValid(val)) {
        setUsernameStatus('idle')
        return
      }
      setUsernameStatus('checking')
      debounceRef.current = setTimeout(() => {
        if (existingUsernames.includes(val.toLowerCase())) {
          setUsernameStatus('exists')
        } else {
          setUsernameStatus('ok')
        }
      }, 300)
    },
    [existingUsernames],
  )

  const passwordError =
    form.password.length > 0 && !/^(?=.*[a-zA-Z])(?=.*\d).{8,}$/.test(form.password)

  const canSave =
    isUsernameShapeValid(form.username) &&
    form.password.length >= 8 &&
    /^(?=.*[a-zA-Z])(?=.*\d)/.test(form.password) &&
    usernameStatus !== 'exists' &&
    usernameStatus !== 'checking'

  return (
    <AppDrawer
      width="sm"
      open={open}
      onClose={onClose}
      title={t('adminUsers.addUser')}
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} className="text-accent hover:text-accent-hover">{t('common.cancel')}</Button>
          <Button variant="default" size="default" disabled={!canSave} onClick={() => onSave(form)}>
            {t('common.save')}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('common.username')}</Label>
          <div className="relative">
            <Input
              value={form.username}
              onChange={(e) => {
                const v = e.target.value.replace(/[^a-zA-Z0-9_-]/g, '')
                setForm((f) => ({ ...f, username: v }))
                checkUsername(v)
              }}
              placeholder={t('adminUsers.usernamePlaceholder')}
              className="h-9 pr-8 text-[13px]"
              maxLength={16}
            />
            {usernameStatus === 'checking' && (
              <Loader2 className="absolute right-2.5 top-1/2 size-3.5 -translate-y-1/2 animate-spin text-text-3" />
            )}
            {usernameStatus === 'ok' && (
              <Check className="absolute right-2.5 top-1/2 size-3.5 -translate-y-1/2 text-accent" />
            )}
          </div>
          {usernameFormatError && (
            <p className="text-xs text-destructive">{t('adminUsers.usernamePlaceholder')}</p>
          )}
          {!usernameFormatError && usernameStatus === 'exists' && (
            <p className="text-xs text-destructive">{t('adminUsers.usernameExists')}</p>
          )}
        </div>

        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('common.password')}</Label>
          <Input
            type="password"
            value={form.password}
            onChange={(e) => setForm((f) => ({ ...f, password: e.target.value }))}
            placeholder={t('adminUsers.passwordPlaceholder')}
            className="h-9 text-[13px]"
          />
          {passwordError && (
            <p className="text-xs text-destructive">{t('adminUsers.passwordPlaceholder')}</p>
          )}
          {!passwordError && (
            <p className="text-xs text-text-muted">{t('adminUsers.passwordHint')}</p>
          )}
        </div>

        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('adminUsers.nickname')}</Label>
          <Input
            value={form.nickname}
            onChange={(e) => setForm((f) => ({ ...f, nickname: e.target.value }))}
            placeholder={t('common.optional')}
            className="h-9 text-[13px]"
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('common.role')}</Label>
          <select
            value={form.role}
            onChange={(e) => setForm((f) => ({ ...f, role: Number(e.target.value) as 0 | 1 }))}
            className="h-9 rounded-lg border border-border-strong bg-transparent px-2.5 text-[13px] text-text-1 outline-none focus:border-ring focus:ring-3 focus:ring-ring/50"
          >
            <option value={0}>{t('adminUsers.regularUser')}</option>
            <option value={1}>{t('adminUsers.admin')}</option>
          </select>
        </div>
      </div>
    </AppDrawer>
  )
}

function EditUserDrawer({
  open,
  onClose,
  onSave,
  user,
}: {
  open: boolean
  onClose: () => void
  onSave: (data: EditFormData) => void
  user: AdminUserItem | null
}) {
  const t = useT()
  const [form, setForm] = useState<EditFormData>({ nickname: '', email: '', role: 0, status: 1 })

  useEffect(() => {
    if (user && open) {
      setForm({
        nickname: user.nickname,
        email: user.email,
        role: user.role as 0 | 1,
        status: user.status as 0 | 1,
      })
    }
  }, [user, open])

  if (!user) return null

  return (
    <AppDrawer
      width="sm"
      open={open}
      onClose={onClose}
      title={t('adminUsers.editUser')}
      footer={
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} className="text-accent hover:text-accent-hover">{t('common.cancel')}</Button>
          <Button variant="default" size="default" onClick={() => onSave(form)}>
            {t('common.save')}
          </Button>
        </div>
      }
    >
      <div className="flex flex-col gap-4">
        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('common.username')}</Label>
          <Input value={user.username} disabled className="h-9 text-[13px] opacity-50" />
          <p className="text-xs text-text-muted">{t('adminUsers.usernameReadonly')}</p>
        </div>

        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('adminUsers.nickname')}</Label>
          <Input
            value={form.nickname}
            onChange={(e) => setForm((f) => ({ ...f, nickname: e.target.value }))}
            className="h-9 text-[13px]"
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('common.email')}</Label>
          <Input
            type="email"
            value={form.email}
            onChange={(e) => setForm((f) => ({ ...f, email: e.target.value }))}
            className="h-9 text-[13px]"
          />
        </div>

        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('common.role')}</Label>
          <select
            value={form.role}
            onChange={(e) => setForm((f) => ({ ...f, role: Number(e.target.value) as 0 | 1 }))}
            className="h-9 rounded-lg border border-border-strong bg-transparent px-2.5 text-[13px] text-text-1 outline-none focus:border-ring focus:ring-3 focus:ring-ring/50"
          >
            <option value={0}>{t('adminUsers.regularUser')}</option>
            <option value={1}>{t('adminUsers.admin')}</option>
          </select>
        </div>

        <div className="flex flex-col gap-1.5">
          <Label className="text-[13px] font-medium text-text-2">{t('common.status')}</Label>
          <div className="flex items-center gap-2">
            <StatusToggle
              checked={form.status === 1}
              disabled={false}
              onChange={(v) => setForm((f) => ({ ...f, status: v ? 1 : 0 }))}
            />
            <span className="text-xs text-text-3">{form.status === 1 ? t('adminUsers.statusEnabled') : t('adminUsers.statusDisabled')}</span>
          </div>
          {form.status === 0 && (
            <p className="text-xs text-text-muted">{t('adminUsers.disableHint')}</p>
          )}
        </div>
      </div>
    </AppDrawer>
  )
}

/* ============================================
   Reset Password Dialog
   ============================================ */

function ResetPasswordDialog({
  open,
  onClose,
  onConfirm,
  username,
}: {
  open: boolean
  onClose: () => void
  onConfirm: (password: string) => void
  username: string
}) {
  const t = useT()
  const [pwd, setPwd] = useState('')
  const [confirm, setConfirm] = useState('')
  const mismatch = confirm.length > 0 && pwd !== confirm
  const pwdFormatError = pwd.length > 0 && !/^(?=.*[a-zA-Z])(?=.*\d).{8,}$/.test(pwd)

  useEffect(() => {
    if (open) {
      setPwd('')
      setConfirm('')
    }
  }, [open])

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('adminUsers.resetPassword')}</DialogTitle>
          <DialogDescription>{translate('adminUsers.setPasswordFor').replace('{name}', username)}</DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-3">
          <div className="flex flex-col gap-1.5">
            <Label className="text-[13px] font-medium text-text-2">{t('adminUsers.setNewPassword')}</Label>
            <Input
              type="password"
              value={pwd}
              onChange={(e) => setPwd(e.target.value)}
              placeholder={t('adminUsers.passwordPlaceholder')}
              className="h-9 text-[13px]"
            />
            {pwdFormatError && (
              <p className="text-xs text-destructive">{t('adminUsers.passwordPlaceholder')}</p>
            )}
          </div>
          <div className="flex flex-col gap-1.5">
            <Label className="text-[13px] font-medium text-text-2">{t('common.confirmPassword')}</Label>
            <Input
              type="password"
              value={confirm}
              onChange={(e) => setConfirm(e.target.value)}
              placeholder={translate('adminUsers.passwordPlaceholder')}
              className="h-9 text-[13px]"
            />
            {mismatch && <p className="text-xs text-destructive">{translate('common.confirmPassword')}</p>}
          </div>
          <p className="text-xs text-text-muted">{t('adminUsers.passwordHint')}</p>
          <div className="mt-1 flex justify-end gap-2">
            <Button variant="ghost" size="default" onClick={onClose}>{t('common.cancel')}</Button>
            <Button
              variant="default"
              size="default"
              disabled={!pwd || pwdFormatError || mismatch}
              onClick={() => onConfirm(pwd)}
            >
              {t('adminUsers.confirmReset')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Delete Confirm Dialog
   ============================================ */

function DeleteUserDialog({
  open,
  onClose,
  onConfirm,
  username,
}: {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  username: string
}) {
  const t = useT()
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('adminUsers.deleteUser')}</DialogTitle>
          <DialogDescription>
            {translate('adminUsers.deleteConfirm').replace('{name}', username)}
          </DialogDescription>
        </DialogHeader>
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose}>{t('common.cancel')}</Button>
          <Button variant="destructive" size="default" onClick={onConfirm}>
            {t('common.delete')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Token Limit Dialog
   ============================================ */

const TOKEN_LIMIT_PRESETS = [10, 20, 50, 80, 100]

function TokenLimitDialog({
  user,
  onClose,
  onSave,
}: {
  user: AdminUserItem | null
  onClose: () => void
  onSave: (userId: string, limit: number) => Promise<void>
}) {
  const t = useT()
  const [mode, setMode] = useState<'preset' | 'custom'>('preset')
  const [preset, setPreset] = useState<number>(0)
  const [custom, setCustom] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (user) {
      const limit = user.dailyTokenLimit ?? 0
      if (TOKEN_LIMIT_PRESETS.includes(limit)) {
        setMode('preset')
        setPreset(limit)
      } else if (limit > 0) {
        setMode('custom')
        setCustom(String(limit))
      } else {
        setMode('preset')
        setPreset(0)
      }
      setSaving(false)
    }
  }, [user])

  if (!user) return null

  const value = mode === 'preset' ? preset : Math.max(0, Math.floor(Number(custom) || 0))

  const handleSave = () => {
    setSaving(true)
    onSave(user.userId, value)
      .catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
      .finally(() => setSaving(false))
  }

  return (
    <Dialog open={!!user} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-md">
        <DialogHeader>
          <DialogTitle>{t('adminUsers.editTokenLimit')}</DialogTitle>
          <DialogDescription>
            {translate('adminUsers.tokenLimitFor').replace('{name}', user.username)}
          </DialogDescription>
        </DialogHeader>
        <div className="flex flex-col gap-3">
          <div className="flex flex-wrap gap-1.5">
            {[0, ...TOKEN_LIMIT_PRESETS].map((v) => (
              <button
                key={v}
                onClick={() => { setMode('preset'); setPreset(v) }}
                className={cn(
                  'rounded-lg border px-3 py-1.5 text-xs transition-colors',
                  mode === 'preset' && preset === v
                    ? 'border-accent bg-accent/10 font-medium text-accent'
                    : 'border-border-strong text-text-2 hover:border-text-3 hover:text-text-1',
                )}
              >
                {v === 0 ? t('adminUsers.unlimited') : `${v}M`}
              </button>
            ))}
          </div>
          <div
            onClick={() => setMode('custom')}
            className={cn(
              'flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 transition-colors',
              mode === 'custom' ? 'border-accent' : 'border-border-strong',
            )}
          >
            <span className={cn('whitespace-nowrap text-xs', mode === 'custom' ? 'font-medium text-accent' : 'text-text-2')}>
              {t('common.custom')}
            </span>
            <input
              type="number"
              min={1}
              value={custom}
              onChange={(e) => { setMode('custom'); setCustom(e.target.value) }}
              onFocus={() => setMode('custom')}
              placeholder="50"
              className="h-6 w-20 bg-transparent text-sm tabular-nums text-text-1 outline-none placeholder:text-text-muted"
            />
            <span className="text-xs text-text-3">M</span>
          </div>
          <p className="text-xs text-text-muted">{t('adminUsers.tokenLimitUnit')}</p>
          <div className="mt-1 flex justify-end gap-2">
            <Button variant="ghost" size="default" onClick={onClose}>{t('common.cancel')}</Button>
            <Button
              variant="default"
              size="default"
              disabled={saving || (mode === 'custom' && value <= 0)}
              onClick={handleSave}
            >
              {saving && <Loader2 className="size-3.5 animate-spin" />}
              {t('common.save')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Table Row
   ============================================ */

function UserRow({
  user,
  isCurrentUser,
  onEdit,
  onDelete,
  onResetPassword,
  onToggleStatus,
  onEditLimit,
}: {
  user: AdminUserItem
  isCurrentUser: boolean
  onEdit: () => void
  onDelete: () => void
  onResetPassword: () => void
  onToggleStatus: (v: boolean) => void
  onEditLimit: () => void
}) {
  const t = useT()
  return (
    <TableRow className="group">
      <TableCell className="h-12 w-10 py-0 pl-4 pr-2">
        <Avatar username={user.username} avatar={user.avatar} />
      </TableCell>
      <TableCell className="h-12 whitespace-nowrap py-0 pr-4">
        <div className="flex items-center gap-2">
          <span className="text-[13px] font-medium text-text-1">{user.username}</span>
          {isCurrentUser && (
            <Shield className="size-3.5 text-amber-strong" />
          )}
        </div>
      </TableCell>
      <TableCell className="h-12 py-0 pr-4 text-text-2">{user.nickname || '—'}</TableCell>
      <TableCell className="h-12 py-0 pr-4 text-text-2">{user.email || '—'}</TableCell>
      <TableCell className="h-12 py-0 pr-4">
        <Badge variant={user.role === 1 ? 'info' : 'neutral'}>{getRoleLabel(user.role)}</Badge>
      </TableCell>
      <TableCell className="h-12 py-0 pr-4">
        <StatusToggle
          checked={user.status === 1}
          disabled={isCurrentUser}
          onChange={onToggleStatus}
        />
      </TableCell>
      <TableCell className="h-12 py-0 pr-4 text-right">
        <button
          onClick={onEditLimit}
          title={t('adminUsers.editTokenLimit')}
          className="inline-flex items-center gap-1 rounded-md px-1.5 py-0.5 text-[13px] tabular-nums text-accent transition-colors hover:bg-bg-layer-2 hover:text-accent-hover"
        >
          {user.dailyTokenLimit > 0 ? `${user.dailyTokenLimit}M` : <span className="text-text-3">—</span>}
          <Pencil className="size-3 text-text-3" />
        </button>
      </TableCell>
      <TableCell className="h-12 py-0 pr-4 text-right tabular-nums text-text-2">{user.sessionCount}</TableCell>
      <TableCell className="h-12 py-0 pr-4 text-text-3">
        {formatRelative(user.lastActiveTime)}
      </TableCell>
      <TableCell className="h-12 py-0 pr-4">
        <div className="flex items-center justify-end gap-0.5 opacity-100 transition-opacity md:opacity-0 md:group-hover:opacity-100 md:group-focus-within:opacity-100">
          {!isCurrentUser && (
            <>
              <button
                onClick={onEdit}
                className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
                title={t('common.edit')} aria-label={t('common.edit')}
              >
                <Pencil className="size-3.5" />
              </button>
              <button
                onClick={onResetPassword}
                className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1"
                title={t('adminUsers.resetPassword')} aria-label={t('adminUsers.resetPassword')}
              >
                <KeyRound className="size-3.5" />
              </button>
              <button
                onClick={onDelete}
                className="rounded p-1 text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-destructive"
                title={t('common.delete')} aria-label={t('common.delete')}
              >
                <Trash2 className="size-3.5" />
              </button>
            </>
          )}
        </div>
      </TableCell>
    </TableRow>
  )
}

/* ============================================
   Skeleton
   ============================================ */

function TableSkeleton() {
  const t = useT()
  return (
    <div className="flex-1 overflow-auto">
      <Table>
        <TableHeader>
          <TableRow className="hover:bg-transparent">
            <TableHead className="w-10 pl-4 pr-2" />
            <TableHead className="w-40 pr-4">{t('common.username')}</TableHead>
            <TableHead className="pr-4">{t('adminUsers.nickname')}</TableHead>
            <TableHead className="pr-4">{t('common.email')}</TableHead>
            <TableHead className="pr-4">{t('common.role')}</TableHead>
            <TableHead className="pr-4">{t('common.status')}</TableHead>
            <TableHead className="pr-4 text-right">{t('adminUsers.dailyTokenLimit')}</TableHead>
            <TableHead className="pr-4 text-right">{t('adminUsers.sessions')}</TableHead>
            <TableHead className="pr-4">{t('adminUsers.lastActive')}</TableHead>
            <TableHead className="pr-4" />
          </TableRow>
        </TableHeader>
        <TableBody>
          {Array.from({ length: 8 }).map((_, i) => (
            <TableRow key={i} className="hover:bg-transparent">
              <TableCell className="h-12 py-0 pl-4 pr-2">
                <div className="size-7 animate-pulse rounded-sm bg-bg-layer-3" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="h-3.5 w-20 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="h-3.5 w-16 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="h-3.5 w-28 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="h-5 w-12 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="h-4 w-8 animate-pulse rounded-full bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="ml-auto h-3.5 w-8 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="ml-auto h-3.5 w-6 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="h-3.5 w-14 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
              <TableCell className="h-12 py-0 pr-4">
                <div className="ml-auto h-3.5 w-16 animate-pulse rounded bg-bg-layer-2" />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

/* ============================================
   Pagination
   ============================================ */

function Pagination({
  page,
  totalPages,
  totalCount,
  onPageChange,
}: {
  page: number
  totalPages: number
  totalCount: number
  onPageChange: (p: number) => void
}) {
  const t = useT()
  const pages = useMemo(() => {
    const result: (number | '...')[] = []
    if (totalPages <= 7) {
      for (let i = 1; i <= totalPages; i++) result.push(i)
    } else {
      result.push(1)
      if (page > 3) result.push('...')
      for (let i = Math.max(2, page - 1); i <= Math.min(totalPages - 1, page + 1); i++) {
        result.push(i)
      }
      if (page < totalPages - 2) result.push('...')
      result.push(totalPages)
    }
    return result
  }, [page, totalPages])

  if (totalPages <= 1) return null

  return (
    <div className="flex items-center justify-between border-t border-border px-4 py-2">
      <span className="text-xs text-text-3">{t('adminUsers.totalUsers').replace('{count}', String(totalCount))}</span>
      <div className="flex items-center gap-1">
        <button
          onClick={() => onPageChange(page - 1)}
          disabled={page === 1}
          className="inline-flex size-8 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          <ChevronLeft className="size-3.5" />
        </button>
        {pages.map((p, i) =>
          p === '...' ? (
            <span key={`dots-${i}`} className="px-1 text-xs text-text-3">
              ...
            </span>
          ) : (
            <button
              key={p}
              onClick={() => onPageChange(p)}
              className={cn(
                'inline-flex size-8 items-center justify-center rounded-md text-xs tabular-nums transition-colors',
                p === page
                  ? 'bg-bg-layer-3 text-text-1 font-medium'
                  : 'text-text-3 hover:bg-bg-layer-2 hover:text-text-1',
              )}
            >
              {p}
            </button>
          ),
        )}
        <button
          onClick={() => onPageChange(page + 1)}
          disabled={page === totalPages}
          className="inline-flex size-8 items-center justify-center rounded-md text-text-3 transition-colors hover:bg-bg-layer-2 hover:text-text-1 disabled:opacity-30 disabled:hover:bg-transparent"
        >
          <ChevronRight className="size-3.5" />
        </button>
      </div>
    </div>
  )
}

/* ============================================
   Main Component
   ============================================ */

export function Component() {
  const t = useT()
  const [state, setState] = useState<PageState>('loading')
  const [users, setUsers] = useState<AdminUserItem[]>([])
  const [totalCount, setTotalCount] = useState(0)
  const [totalPages, setTotalPages] = useState(1)
  const [search, setSearch] = useState('')
  const [roleFilter, setRoleFilter] = useState<'all' | '0' | '1'>('all')
  const [page, setPage] = useState(1)

  const [drawerMode, setDrawerMode] = useState<DrawerMode | null>(null)
  const [editTarget, setEditTarget] = useState<AdminUserItem | null>(null)
  const [dialogMode, setDialogMode] = useState<DialogMode>(null)
  const [dialogTarget, setDialogTarget] = useState<AdminUserItem | null>(null)
  const [limitTarget, setLimitTarget] = useState<AdminUserItem | null>(null)

  const loadData = useCallback(() => {
    setState('loading')
    adminUsersApi.getUsers({ page, pageSize: 20, search: search || undefined, role: roleFilter === 'all' ? undefined : (Number(roleFilter) as 0 | 1) })
      .then((res) => {
        setUsers(res.list)
        setTotalCount(res.totalCount)
        setTotalPages(res.totalPage)
        setState(res.list.length > 0 ? 'data' : 'empty')
      })
      .catch(() => {
        setState('error')
      })
  }, [page, search, roleFilter])

  useEffect(() => {
    loadData()
  }, [loadData])

  // Derive table display name
  const existingUsernames = useMemo(() => users.map((u) => u.username), [users])

  const safePage = Math.min(page, totalPages)

  // Reset page on filter/search change
  useEffect(() => {
    setPage(1)
  }, [search, roleFilter])

  const handleSearch = useCallback((val: string) => {
    setSearch(val)
  }, [])

  const handleClearFilter = useCallback(() => {
    setSearch('')
    setRoleFilter('all')
  }, [])

  // Add user
  const handleAddUser = useCallback(
    (data: UserFormData) => {
      adminUsersApi.createUser({ username: data.username, password: data.password, nickname: data.nickname || undefined, role: data.role }).then(() => {
        setDrawerMode(null)
        toast.success(translate('adminUsers.userCreated'))
        loadData()
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [loadData],
  )

  // Edit user
  const handleEditUser = useCallback(
    (data: EditFormData) => {
      if (!editTarget) return
      adminUsersApi.updateUser(editTarget.userId, { nickname: data.nickname, email: data.email, role: data.role, status: data.status }).then(() => {
        setDrawerMode(null)
        setEditTarget(null)
        toast.success(translate('adminUsers.userUpdated'))
        loadData()
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [editTarget, loadData],
  )

  // Reset password
  const handleResetPassword = useCallback(
    (password: string) => {
      if (!dialogTarget) return
      adminUsersApi.resetPassword(dialogTarget.userId, { password }).then(() => {
        toast.success(translate('adminUsers.passwordReset'))
        setDialogMode(null)
        setDialogTarget(null)
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [dialogTarget],
  )

  // Delete user
  const handleDeleteUser = useCallback(() => {
    if (!dialogTarget) return
    adminUsersApi.deleteUser(dialogTarget.userId).then(() => {
      setDialogMode(null)
      setDialogTarget(null)
      toast.success(translate('adminUsers.userDeleted'))
      loadData()
    }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
  }, [dialogTarget, loadData])

  // Toggle status
  const handleToggleStatus = useCallback(
    (userId: string, enabled: boolean) => {
      const user = users.find((u) => u.userId === userId)
      if (!user) return
      adminUsersApi.updateUser(userId, { nickname: user.nickname, email: user.email, role: user.role, status: enabled ? 1 : 0 }).then(() => {
        setUsers((prev) =>
          prev.map((u) => (u.userId === userId ? { ...u, status: (enabled ? 1 : 0) as 0 | 1 } : u)),
        )
      }).catch((err: Error) => { toast.error(err.message || translate('common.failed')) })
    },
    [users],
  )

  // Update daily token limit
  const handleSaveTokenLimit = useCallback(
    (userId: string, limit: number) => {
      return adminUsersApi.updateUser(userId, { dailyTokenLimit: limit }).then(() => {
        setLimitTarget(null)
        toast.success(translate('adminUsers.userUpdated'))
        loadData()
      })
    },
    [loadData],
  )

  const currentUserId = useUserStore((s) => s.currentUser?.userId)
  const isCurrentUser = useCallback((user: AdminUserItem) => user.userId === currentUserId, [currentUserId])

  const hasFilter = search.trim().length > 0 || roleFilter !== 'all'

  return (
    <div className="flex h-full flex-col bg-bg-base">
      <PageHeader
        title={t('adminUsers.title')}
        actions={
          <>
          <div className="relative">
            {/* Hidden dummy fields to prevent Chrome autofill from targeting the search input */}
            <input type="text" name="username" style={{ position: 'absolute', left: -9999 }} tabIndex={-1} autoComplete="username" />
            <input type="password" name="password" style={{ position: 'absolute', left: -9999 }} tabIndex={-1} autoComplete="current-password" />
            <Search className="absolute left-2.5 top-1/2 size-3.5 -translate-y-1/2 text-text-3" />
            <input
              type="text"
              name="search"
              autoComplete="off"
              spellCheck={false}
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              placeholder={t('adminUsers.searchPlaceholder')}
              className="h-7 w-40 rounded-lg border border-border-strong bg-transparent pl-7 pr-2 text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-ring focus:ring-2 focus:ring-ring/30"
            />
            {search && (
              <button
                onClick={() => handleSearch('')}
                className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded-sm p-0.5 text-text-3 hover:text-text-1"
              >
                <X className="size-3" />
              </button>
            )}
          </div>
          <select
            value={roleFilter}
            onChange={(e) => setRoleFilter(e.target.value as 'all' | '0' | '1')}
            className="h-7 rounded-lg border border-border-strong bg-transparent px-2 text-xs text-text-1 outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
          >
            <option value="all">{t('adminUsers.allRoles')}</option>
            <option value="1">{t('adminUsers.admin')}</option>
            <option value="0">{t('adminUsers.regularUser')}</option>
          </select>
          <Button
            variant="default"
            size="sm"
            onClick={() => setDrawerMode('add')}
           
          >
            <Plus className="size-3.5" />
            {t('adminUsers.addUser')}
          </Button>
          </>
        }
      />

      {/* Content */}
      {state === 'loading' && <TableSkeleton />}

      {state === 'error' && <ErrorState title={t('adminUsers.fetchFailed')} description={t('adminUsers.fetchFailed')} onRetry={loadData} />}

      {state === 'empty' && (
        <EmptyState
          tone={1}
          icon={<User className="size-5" />}
          title={t('adminUsers.noUsers')}
          description={t('adminUsers.createFirst')}
        />
      )}

      {state === 'data' && users.length === 0 && (
        <EmptyState
          tone={1}
          icon={<User className="size-5" />}
          title={hasFilter ? t('adminUsers.noMatch') : t('adminUsers.noUsers')}
          description={hasFilter ? undefined : t('adminUsers.createFirst')}
          action={
            hasFilter ? (
              <button
                onClick={handleClearFilter}
                className="text-xs text-interactive transition-colors hover:text-interactive-hover"
              >
                {t('adminUsers.clearFilter')}
              </button>
            ) : undefined
          }
        />
      )}

      {state === 'data' && users.length > 0 && (
        <>
          <div className="flex-1 overflow-auto">
            <Table>
              <TableHeader>
                <TableRow className="hover:bg-transparent">
                  <TableHead className="w-10 pl-4 pr-2" />
                  <TableHead className="w-40 pr-4">{t('common.username')}</TableHead>
                  <TableHead className="pr-4">{t('adminUsers.nickname')}</TableHead>
                  <TableHead className="pr-4">{t('common.email')}</TableHead>
                  <TableHead className="pr-4">{t('common.role')}</TableHead>
                  <TableHead className="pr-4">{t('common.status')}</TableHead>
                  <TableHead className="pr-4 text-right">{t('adminUsers.dailyTokenLimit')}</TableHead>
                  <TableHead className="pr-4 text-right">{t('adminUsers.sessions')}</TableHead>
                  <TableHead className="pr-4">{t('adminUsers.lastActive')}</TableHead>
                  <TableHead className="pr-4" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <UserRow
                    key={user.userId}
                    user={user}
                    isCurrentUser={isCurrentUser(user)}
                    onEdit={() => {
                      setEditTarget(user)
                      setDrawerMode('edit')
                    }}
                    onDelete={() => {
                      setDialogTarget(user)
                      setDialogMode('delete')
                    }}
                    onResetPassword={() => {
                      setDialogTarget(user)
                      setDialogMode('resetPassword')
                    }}
                    onToggleStatus={(v) => handleToggleStatus(user.userId, v)}
                    onEditLimit={() => setLimitTarget(user)}
                  />
                ))}
              </TableBody>
            </Table>
          </div>
          <Pagination
            page={safePage}
            totalPages={totalPages}
            totalCount={totalCount}
            onPageChange={setPage}
          />
        </>
      )}

      {/* Drawers */}
      <AddUserDrawer
        open={drawerMode === 'add'}
        onClose={() => setDrawerMode(null)}
        onSave={handleAddUser}
        existingUsernames={existingUsernames}
      />

      <EditUserDrawer
        open={drawerMode === 'edit'}
        onClose={() => {
          setDrawerMode(null)
          setEditTarget(null)
        }}
        onSave={handleEditUser}
        user={editTarget}
      />

      {/* Dialogs */}
      <ResetPasswordDialog
        open={dialogMode === 'resetPassword'}
        onClose={() => {
          setDialogMode(null)
          setDialogTarget(null)
        }}
        onConfirm={handleResetPassword}
        username={dialogTarget?.username ?? ''}
      />

      <DeleteUserDialog
        open={dialogMode === 'delete'}
        onClose={() => {
          setDialogMode(null)
          setDialogTarget(null)
        }}
        onConfirm={handleDeleteUser}
        username={dialogTarget?.username ?? ''}
      />

      <TokenLimitDialog
        user={limitTarget}
        onClose={() => setLimitTarget(null)}
        onSave={handleSaveTokenLimit}
      />
    </div>
  )
}
