import { useState, useCallback, useRef, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import dayjs from 'dayjs'
import {
  Pencil,
  Camera,
  KeyRound,
  LogOut,
  User as UserIcon,
  Sun,
  Moon,
  Monitor,
  Braces,
} from 'lucide-react'
import { useT, t as translate } from '@/i18n'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Logo } from '@/components/brand/Logo'
import { ErrorState } from '@/components/patterns/ErrorState'
import { PageHeader } from '@/components/patterns/PageHeader'
import { PageContainer } from '@/components/patterns/PageContainer'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { useAppStore } from '@/stores/app-store'
import { useUserStore } from '@/stores/user-store'
import { getCurrentUser, updateProfile, changePassword, logout } from '@/api/auth'
import { useAvatarUpload } from '@/hooks/useAvatarUpload'
import type { User } from '@/api/types'
import { EnvVarsDialog } from './EnvVarsDialog'

type Theme = 'light' | 'dark' | 'system'

/* ============================================
   Inline Editor
   ============================================ */

function InlineEditor({
  value,
  placeholder,
  type = 'text',
  maxLength,
  onSave,
  onCancel,
  error,
}: {
  value: string
  placeholder?: string
  type?: string
  maxLength?: number
  onSave: (value: string) => void
  onCancel: () => void
  error?: string
}) {
  const [text, setText] = useState(value)
  const inputRef = useRef<HTMLInputElement>(null)

  const handleBlur = useCallback(() => {
    const trimmed = text.trim()
    if (trimmed === value || (trimmed === '' && value === '')) {
      onCancel()
      return
    }
    onSave(trimmed)
  }, [text, value, onSave, onCancel])

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Enter') {
        ;(e.target as HTMLInputElement).blur()
      } else if (e.key === 'Escape') {
        setText(value)
        onCancel()
      }
    },
    [value, onCancel],
  )

  return (
    <div className="flex flex-col gap-1">
      <Input
        ref={inputRef}
        type={type}
        value={text}
        onChange={(e) => setText(e.target.value)}
        onBlur={handleBlur}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        maxLength={maxLength}
        autoFocus
        className="h-7 text-sm"
      />
      {error && (
        <span className="text-xs text-text-3 opacity-80">{error}</span>
      )}
    </div>
  )
}

/* ============================================
   Avatar Upload Dialog
   ============================================ */

function AvatarDialog({
  open,
  onOpenChange,
  currentAvatar,
  onConfirm,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  currentAvatar: string
  onConfirm: (file: File) => void
}) {
  const t = useT()
  const [preview, setPreview] = useState<string | null>(null)
  const [file, setFile] = useState<File | null>(null)
  const [dragOver, setDragOver] = useState(false)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFile = useCallback((f: File) => {
    if (!f.type.match(/^image\/(jpeg|png|gif)$/)) {
      toast.error(translate('profile.errAvatarFormat'))
      return
    }
    if (f.size > 2 * 1024 * 1024) {
      toast.error(translate('profile.errAvatarSize'))
      return
    }
    setFile(f)
    const reader = new FileReader()
    reader.onload = () => setPreview(reader.result as string)
    reader.readAsDataURL(f)
  }, [])

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(true)
  }, [])

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault()
    setDragOver(false)
  }, [])

  const handleDrop = useCallback(
    (e: React.DragEvent) => {
      e.preventDefault()
      setDragOver(false)
      const f = e.dataTransfer.files[0]
      if (f) handleFile(f)
    },
    [handleFile],
  )

  const handleConfirm = useCallback(() => {
    if (file) {
      onConfirm(file)
      setPreview(null)
      setFile(null)
      onOpenChange(false)
    }
  }, [file, onConfirm, onOpenChange])

  const handleClose = useCallback(() => {
    setPreview(null)
    setFile(null)
    onOpenChange(false)
  }, [onOpenChange])

  const displaySrc = preview || currentAvatar

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('profile.changeAvatar')}</DialogTitle>
          <DialogDescription>{t('profile.avatarHint')}</DialogDescription>
        </DialogHeader>

        <div className="flex flex-col items-center gap-4">
          <div
            className={cn(
              'flex h-32 w-32 items-center justify-center overflow-hidden rounded-full bg-bg-layer-3',
              dragOver && 'ring-2 ring-accent ring-offset-2 ring-offset-bg-base',
            )}
          >
            {displaySrc ? (
              <img
                src={displaySrc}
                alt="avatar preview"
                className="h-full w-full object-cover"
              />
            ) : (
              <UserIcon className="size-12 text-text-3" />
            )}
          </div>

          <div
            onDragOver={handleDragOver}
            onDragLeave={handleDragLeave}
            onDrop={handleDrop}
            onClick={() => fileInputRef.current?.click()}
            className={cn(
              'w-full cursor-pointer rounded-md border border-dashed border-accent-border px-4 py-6 text-center transition-colors hover:border-accent hover:bg-accent-soft/40',
              dragOver && 'border-accent bg-accent-soft',
            )}
          >
            <p className="text-sm text-text-2">
              {file ? file.name : t('profile.avatarDropHint')}
            </p>
          </div>
          <input
            ref={fileInputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif"
            onChange={(e) => {
              const f = e.target.files?.[0]
              if (f) handleFile(f)
            }}
            className="hidden"
          />
        </div>

        <div className="mt-4 flex justify-end gap-2">
          <Button variant="outline" onClick={handleClose}>
             {t('common.cancel')}
          </Button>
          <Button onClick={handleConfirm} disabled={!file}>
            {t('common.confirm')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Password Dialog
   ============================================ */

function PasswordDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const t = useT()
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [submitting, setSubmitting] = useState(false)

  const validate = useCallback(() => {
    const e: Record<string, string> = {}
    if (!currentPassword) e.currentPassword = translate('profile.errCurrentPasswordRequired')
    if (!newPassword) e.newPassword = translate('profile.errNewPasswordRequired')
    else if (newPassword.length < 8 || !/^(?=.*[a-zA-Z])(?=.*\d)/.test(newPassword)) e.newPassword = translate('profile.errNewPasswordMin')
    else if (newPassword === currentPassword) e.newPassword = translate('profile.errPasswordSame')
    if (!confirmPassword) e.confirmPassword = translate('profile.errConfirmRequired')
    else if (confirmPassword !== newPassword) e.confirmPassword = translate('profile.errPasswordMismatch')
    setErrors(e)
    return Object.keys(e).length === 0
  }, [currentPassword, newPassword, confirmPassword])

  const handleSubmit = useCallback(
    async (e: FormEvent) => {
      e.preventDefault()
      if (!validate()) return
      setSubmitting(true)
      try {
        await changePassword({ currentPassword, newPassword })
        toast.success(translate('profile.passwordChanged'))
        setCurrentPassword('')
        setNewPassword('')
        setConfirmPassword('')
        setErrors({})
        onOpenChange(false)
      } catch (err) {
        toast.error(err instanceof Error ? err.message : translate('profile.changeFailed'))
      }
      setSubmitting(false)
    },
    [validate, onOpenChange, currentPassword, newPassword],
  )

  const handleClose = useCallback(() => {
    if (submitting) return
    setCurrentPassword('')
    setNewPassword('')
    setConfirmPassword('')
    setErrors({})
    onOpenChange(false)
  }, [submitting, onOpenChange])

  const fieldClass = 'mb-3'
  const labelClass = 'mb-1 block text-sm text-text-2'
  const errorClass = 'mt-0.5 text-xs text-text-3'

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('profile.changePassword')}</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit}>
          <div className={fieldClass}>
            <label className={labelClass}>{t('profile.currentPassword')}</label>
            <Input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              placeholder={t('profile.currentPasswordPlaceholder')}
              className="h-8"
            />
            {errors.currentPassword && (
              <p className={errorClass}>{errors.currentPassword}</p>
            )}
          </div>

          <div className={fieldClass}>
            <label className={labelClass}>{t('profile.newPassword')}</label>
            <Input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder={t('profile.newPasswordMin')}
              className="h-8"
            />
            {errors.newPassword && (
              <p className={errorClass}>{errors.newPassword}</p>
            )}
          </div>

          <div className={fieldClass}>
            <label className={labelClass}>{t('profile.confirmNewPassword')}</label>
            <Input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              placeholder={t('profile.confirmNewPasswordPlaceholder')}
              className="h-8"
            />
            {errors.confirmPassword && (
              <p className={errorClass}>{errors.confirmPassword}</p>
            )}
          </div>

          <div className="mt-5 flex justify-end gap-2">
            <Button variant="outline" type="button" onClick={handleClose}>
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={submitting}>
              {submitting ? t('common.submitting') : t('profile.confirmChange')}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Logout Dialog
   ============================================ */

function LogoutDialog({
  open,
  onOpenChange,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
}) {
  const t = useT()
  const navigate = useNavigate()

  const handleLogout = useCallback(async () => {
    onOpenChange(false)
    try {
      await logout()
    } finally {
      useUserStore.getState().clearAuth()
      navigate('/login', { replace: true })
    }
  }, [navigate, onOpenChange])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('profile.logout')}</DialogTitle>
          <DialogDescription>{t('profile.logoutConfirm')}</DialogDescription>
        </DialogHeader>

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            {t('common.cancel')}
          </Button>
          <Button variant="destructive" onClick={handleLogout}>
            {t('profile.logout')}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}

/* ============================================
   Theme Segment (placeholder)
   ============================================ */

const THEME_OPTIONS = [
  { value: 'light' as const, key: 'profile.light' as const, icon: Sun },
  { value: 'dark' as const, key: 'profile.dark' as const, icon: Moon },
  { value: 'system' as const, key: 'profile.system' as const, icon: Monitor },
] as const

function ThemeSegment() {
  const t = useT()
  const theme = useAppStore((s) => s.theme)
  const setTheme = useAppStore((s) => s.setTheme)

  const handleChange = useCallback((t: Theme) => {
    setTheme(t)
  }, [setTheme])

  return (
    <div className="grid w-full grid-cols-3 gap-0.5 rounded-md bg-bg-layer-2 p-0.5 sm:inline-flex sm:w-auto">
      {THEME_OPTIONS.map((opt) => {
        const active = theme === opt.value
        return (
          <button
            key={opt.value}
            onClick={() => handleChange(opt.value)}
            className={cn(
              'flex cursor-pointer items-center justify-center gap-1.5 rounded-sm px-2 py-1.5 text-sm transition-colors sm:px-3 sm:py-1',
              active
                ? 'bg-accent text-accent-foreground shadow-sm'
                : 'text-text-3 hover:text-text-2',
            )}
          >
            <opt.icon
              className={cn(
                'size-4 sm:size-3.5',
                active
                  ? 'text-accent-foreground'
                  : opt.value === 'light'
                    ? 'text-amber-strong'
                    : opt.value === 'dark'
                      ? 'text-cat-1'
                      : 'text-cat-2',
              )}
            />
            <span className="hidden sm:inline">{t(opt.key)}</span>
          </button>
        )
      })}
    </div>
  )
}

/* ============================================
   Skeleton
   ============================================ */

function ProfileSkeleton() {
  const t = useT()
  return (
    <div className="flex min-h-full flex-col">
      <PageHeader title={t('profile.title')} />
      <PageContainer contained className="flex flex-1 justify-center py-8">
        <div className="w-full max-w-[640px]">
        <div className="mb-8 h-7 w-24 animate-pulse rounded-sm bg-bg-layer-2" />
        <div className="mb-6 h-px bg-border" />
        <div className="flex items-start gap-6">
          <div className="h-16 w-16 animate-pulse rounded-full bg-bg-layer-2" />
          <div className="flex-1 space-y-2">
            <div className="h-5 w-32 animate-pulse rounded-sm bg-bg-layer-2" />
            <div className="h-4 w-48 animate-pulse rounded-sm bg-bg-layer-2" />
            <div className="h-4 w-36 animate-pulse rounded-sm bg-bg-layer-2" />
          </div>
        </div>
        <div className="my-8 h-px bg-border" />
        <div className="space-y-2">
          <div className="h-4 w-24 animate-pulse rounded-sm bg-bg-layer-2" />
          <div className="h-4 w-20 animate-pulse rounded-sm bg-bg-layer-2" />
        </div>
        <div className="my-8 h-px bg-border" />
        <div className="space-y-2">
          <div className="h-4 w-16 animate-pulse rounded-sm bg-bg-layer-2" />
          <div className="h-4 w-20 animate-pulse rounded-sm bg-bg-layer-2" />
        </div>
        </div>
      </PageContainer>
    </div>
  )
}

/* ============================================
   Error State
   ============================================ */

function ProfileError({ onRetry }: { onRetry: () => void }) {
  const t = useT()
  return (
    <div className="flex min-h-full flex-col">
      <PageHeader title={t('profile.title')} />
      <ErrorState description={t('profile.fetchFailed')} onRetry={onRetry} />
    </div>
  )
}

/* ============================================
   Profile Page
   ============================================ */

export function Component() {
  const t = useT()
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(false)

  const [editingField, setEditingField] = useState<'nickname' | 'email' | null>(null)
  const [avatarOpen, setAvatarOpen] = useState(false)
  const [passwordOpen, setPasswordOpen] = useState(false)
  const [logoutOpen, setLogoutOpen] = useState(false)
  const [envVarsOpen, setEnvVarsOpen] = useState(false)

  const { upload: avatarUpload } = useAvatarUpload()

  const fetchUser = useCallback(() => {
    setLoading(true)
    setError(false)
    getCurrentUser()
      .then((u) => {
        setUser(u)
        setLoading(false)
      })
      .catch(() => {
        setError(true)
        setLoading(false)
      })
  }, [])

  useState(() => {
    fetchUser()
  })

  const handleUpdateNickname = useCallback(
    async (nickname: string) => {
      if (!user) return
      if (!nickname) {
        setEditingField(null)
        return
      }
      try {
        await updateProfile({ nickname })
        const updated = await getCurrentUser()
        setUser(updated)
        useUserStore.setState({ currentUser: updated })
        toast.success(translate('profile.nicknameUpdated'))
      } catch (err) {
        toast.error(err instanceof Error ? err.message : translate('profile.changeFailed'))
      }
      setEditingField(null)
    },
    [user],
  )

  const handleUpdateEmail = useCallback(
    async (email: string) => {
      if (!user) return
      try {
        await updateProfile({ email })
        const updated = await getCurrentUser()
        setUser(updated)
        useUserStore.setState({ currentUser: updated })
        if (email) {
          toast.success(translate('profile.emailUpdated'))
        }
      } catch (err) {
        toast.error(err instanceof Error ? err.message : translate('profile.changeFailed'))
      }
      setEditingField(null)
    },
    [user],
  )

  const handleAvatarConfirm = useCallback(
    async (file: File) => {
      try {
        const path = await avatarUpload(file)
        await updateProfile({ avatar: path })
        const updated = await getCurrentUser()
        setUser(updated)
        useUserStore.setState({ currentUser: updated })
        toast.success(translate('profile.avatarUpdated'))
      } catch (err) {
        toast.error(err instanceof Error ? err.message : translate('profile.changeFailed'))
      }
    },
    [avatarUpload],
  )

  if (loading) return <ProfileSkeleton />
  if (error || !user) return <ProfileError onRetry={fetchUser} />

  const displayName = user.nickname || user.username
  const initial = displayName.charAt(0).toUpperCase()
  const roleLabel = user.role === 1 ? translate('profile.admin') : translate('profile.regularUser')
  const createdDate = dayjs(user.created).format('YYYY-MM-DD')

  return (
    <div className="flex min-h-full flex-col">
      {/* Page header */}
      <PageHeader title={t('profile.title')} />

      <PageContainer contained className="flex flex-1 justify-center py-8">
        <div className="w-full max-w-[640px] pb-16">
          {/* Section: 个人信息 */}
        <div className="mb-8">
          <h2 className="mb-5 text-sm font-semibold text-text-2">{t('profile.personalInfo')}</h2>

          <div className="flex items-start gap-5">
            {/* Avatar */}
            <div className="relative shrink-0">
              {user.avatar ? (
                <img
                  src={user.avatar}
                  alt={displayName}
                  className="h-16 w-16 rounded-full object-cover ring-2 ring-cat-1/30 ring-offset-2 ring-offset-bg-base"
                />
              ) : (
                <span className="flex h-16 w-16 items-center justify-center rounded-full bg-cat-1 text-xl font-semibold text-bg-base ring-2 ring-cat-1/30 ring-offset-2 ring-offset-bg-base">
                  {initial}
                </span>
              )}
              <button
                onClick={() => setAvatarOpen(true)}
                className="absolute -bottom-0.5 -right-0.5 flex h-6 w-6 items-center justify-center rounded-full border border-accent-border bg-accent-soft text-accent transition-colors hover:text-accent-hover"
              >
                <Camera className="size-3" />
              </button>
            </div>

            {/* Info */}
            <div className="min-w-0 flex-1 space-y-2">
              <div className="flex items-baseline gap-3">
                {editingField === 'nickname' ? (
                  <InlineEditor
                    value={user.nickname}
                    placeholder={t('profile.nicknamePlaceholder')}
                    maxLength={20}
                    onSave={handleUpdateNickname}
                    onCancel={() => setEditingField(null)}
                  />
                ) : (
                  <>
                    <span className="text-base font-semibold text-text-1">
                      {displayName}
                    </span>
                    <button
                      onClick={() => setEditingField('nickname')}
                      className="text-text-3 transition-colors hover:text-text-1"
                    >
                      <Pencil className="size-3.5" />
                    </button>
                  </>
                )}
              </div>

              <p className="text-sm text-text-3">
                @{user.username}
                <Badge variant="info" className="ml-2 align-middle px-1.5 text-xs">{roleLabel}</Badge>
              </p>

              <div className="flex items-baseline gap-2">
                <span className="shrink-0 text-sm text-text-3">
                  {t('common.email')}
                </span>
                {editingField === 'email' ? (
                  <InlineEditor
                    value={user.email}
                    placeholder={t('profile.emailPlaceholder')}
                    type="email"
                    onSave={handleUpdateEmail}
                    onCancel={() => setEditingField(null)}
                  />
                ) : user.email ? (
                  <>
                    <span className="text-sm text-text-2">
                      {user.email}
                    </span>
                    <button
                      onClick={() => setEditingField('email')}
                      className="text-text-3 transition-colors hover:text-text-1"
                    >
                      <Pencil className="size-3.5" />
                    </button>
                  </>
                ) : (
                  <button
                    onClick={() => setEditingField('email')}
                    className="inline-flex items-center gap-1.5 text-sm text-text-3 transition-colors hover:text-text-1"
                  >
                    <span>{t('profile.notSet')}</span>
                    <Pencil className="size-3" />
                  </button>
                )}
              </div>
              {!user.email && editingField !== 'email' && (
                <p className="text-xs leading-relaxed text-text-3/60">
                  {t('profile.emailHint')}
                </p>
              )}

              <p className="text-xs text-text-3">
                {t('profile.registered')}{createdDate}
              </p>
            </div>
          </div>
        </div>

        <hr className="border-border mb-8" />

        {/* Section: 外观设置 */}
        <div className="mb-8">
          <h2 className="mb-4 text-sm font-semibold text-text-2">{t('profile.appearance')}</h2>

          <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-center sm:gap-4">
            <span className="hidden text-sm text-text-2 sm:block">{t('profile.themeMode')}</span>
            <ThemeSegment />
          </div>
        </div>

        <hr className="border-border mb-8" />

        {/* Section: 环境变量 */}
        <div className="mb-8">
          <h2 className="mb-4 text-sm font-semibold text-text-2">{t('profile.envVarsTitle')}</h2>

          <div className="flex items-center justify-between">
            <p className="text-sm text-text-3">{t('profile.envVarsDesc')}</p>
            <Button variant="outline" size="default" onClick={() => setEnvVarsOpen(true)} className="border-accent text-accent hover:bg-accent/10">
              <Braces className="size-4" />
              {t('profile.envVarsManage')}
            </Button>
          </div>
        </div>

        <hr className="border-border mb-8" />

        {/* Section: 账号安全 */}
        <div className="mb-8">
          <h2 className="mb-4 text-sm font-semibold text-text-2">{t('profile.accountSecurity')}</h2>

          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <span className="text-sm text-text-2">{t('profile.loginAccount')}</span>
              <span className="text-sm text-text-1">{user.username}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm text-text-2">{t('profile.loginPassword')}</span>
              <span className="text-sm tracking-[0.25em] text-text-2">
                  {'••••••••'}
                </span>
            </div>
            <div className="flex items-center justify-end">
              <Button variant="outline" size="default" onClick={() => setPasswordOpen(true)} className="border-interactive text-interactive hover:bg-interactive/10">
                <KeyRound className="size-4" />
                {t('profile.changePassword')}
              </Button>
            </div>
          </div>
        </div>

        <hr className="border-border mb-8" />

        {/* Section: 关于 */}
        <div className="mb-8">
          <h2 className="mb-4 text-sm font-semibold text-text-2">{t('profile.about')}</h2>
          <p className="max-w-lg text-sm leading-relaxed text-text-3">
            {t('profile.aboutDesc')}
          </p>
          <p className="mt-3 text-sm text-text-3">
            <a
              href="https://goraven.dev"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 text-text-3 transition-colors hover:text-text-1"
            >
              <Logo size={16} />
              goraven.dev
            </a>
          </p>
          <p className="mt-2 text-sm text-text-3">
            <a
              href="https://github.com/8treenet/goraven"
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 text-text-3 transition-colors hover:text-text-1"
            >
              <svg className="size-4" viewBox="0 0 16 16" fill="currentColor">
                <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z" />
              </svg>
              {t('common.viewOnGithub')}
            </a>
          </p>
        </div>

        <hr className="border-border mb-8" />

        {/* Section: 退出登录 */}
        <div>
          <h2 className="mb-3 text-sm font-semibold text-text-2">{t('profile.logout')}</h2>
          <div className="flex items-center justify-between">
            <p className="text-sm text-text-3">
              {t('profile.logoutDesc')}
            </p>
            <Button variant="outline" size="default" onClick={() => setLogoutOpen(true)} className="border-destructive/30 text-destructive hover:bg-destructive-soft">
              <LogOut className="size-4" />
              {t('profile.logout')}
            </Button>
          </div>
        </div>

        {/* Dialogs */}
        <AvatarDialog
          open={avatarOpen}
          onOpenChange={setAvatarOpen}
          currentAvatar={user.avatar}
          onConfirm={handleAvatarConfirm}
        />
        <PasswordDialog open={passwordOpen} onOpenChange={setPasswordOpen} />
        <LogoutDialog open={logoutOpen} onOpenChange={setLogoutOpen} />
        <EnvVarsDialog open={envVarsOpen} onOpenChange={setEnvVarsOpen} />
        </div>
      </PageContainer>
    </div>
  )
}
