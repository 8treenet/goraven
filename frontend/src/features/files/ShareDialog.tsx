import { useCallback, useEffect, useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { CloudDownload, Folder, FolderGit2, Loader2, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { GitClonePayload, GitTestPayload, GitTestRsp } from '@/api/types'
import { useT } from '@/i18n'
import { validateDirName } from './file-helpers'
import { AUTH_OPTIONS, GIT_AUTH, deriveProjectNameFromUrl } from './git/git-helpers'

export type ShareDialogMode = 'create' | 'editShare' | 'delete' | null

/** 项目来源选项：图标与 GitCardChip 三态徽标共用同一套词汇 */
const SOURCE_OPTIONS = [
  { value: 'blank', labelKey: 'files.gitSourceBlank', Icon: Folder },
  { value: 'local', labelKey: 'files.gitSourceLocal', Icon: FolderGit2 },
  { value: 'clone', labelKey: 'files.gitSourceClone', Icon: CloudDownload },
] as const

interface ShareDialogProps {
  mode: ShareDialogMode
  projectName: string
  description: string
  onDescriptionChange: (value: string) => void
  onProjectNameChange?: (value: string) => void
  /** 编辑模式下允许改名（个人项目）；默认只读展示项目名（团队项目） */
  allowRenameInEdit?: boolean
  /** 自定义标题（默认 files.createProjectTitle / files.editShareDialogTitle） */
  createTitle?: string
  editTitle?: string
  /** 新建对话框描述（默认 files.createProjectDesc） */
  createDesc?: string
  /** 删除确认描述（默认 files.confirmDeleteProjectDesc） */
  deleteDesc?: string
  /** create 模式下提供则显示「测试连接」按钮（新建前 ls-remote 验证，不建项目） */
  gitTestApi?: (payload: GitTestPayload) => Promise<GitTestRsp>
  onClose: () => void
  /** 确认回调；create 且选择「从 Git 克隆」时携带 git 参数 */
  onConfirm: (git?: GitClonePayload) => void
}

export function ShareDialog({
  mode,
  projectName,
  description,
  onDescriptionChange,
  onProjectNameChange,
  allowRenameInEdit = false,
  createTitle,
  editTitle,
  createDesc,
  deleteDesc,
  gitTestApi,
  onClose,
  onConfirm,
}: ShareDialogProps) {
  const t = useT()
  const nameEditable = mode === 'create' || (mode === 'editShare' && allowRenameInEdit)
  const nameError = nameEditable && projectName.length > 0 ? validateDirName(projectName) : null
  const nameInvalid = nameError !== null

  const [source, setSource] = useState<'blank' | 'local' | 'clone'>('blank')
  const [cloneUrl, setCloneUrl] = useState('')
  const [authType, setAuthType] = useState<number>(GIT_AUTH.NONE)
  const [sshKey, setSshKey] = useState('')
  const [httpsUser, setHttpsUser] = useState('')
  const [httpsSecret, setHttpsSecret] = useState('')
  const [shallow, setShallow] = useState(true)
  const [testing, setTesting] = useState(false)
  const [nameAutoDerived, setNameAutoDerived] = useState(false)

  // 进入 create 时重置克隆表单
  useEffect(() => {
    if (mode === 'create') {
      setSource('blank')
      setCloneUrl('')
      setAuthType(GIT_AUTH.NONE)
      setSshKey('')
      setHttpsUser('')
      setHttpsSecret('')
      setShallow(true)
      setNameAutoDerived(false)
    }
  }, [mode])

  const handleCloneUrlChange = useCallback((value: string) => {
    setCloneUrl(value)
    const derived = deriveProjectNameFromUrl(value)
    if (derived && (nameAutoDerived || !projectName.trim())) {
      setNameAutoDerived(true)
      onProjectNameChange?.(derived)
    }
  }, [nameAutoDerived, projectName, onProjectNameChange])

  const handleNameChange = useCallback((value: string) => {
    if (value !== deriveProjectNameFromUrl(cloneUrl)) setNameAutoDerived(false)
    onProjectNameChange?.(value)
  }, [cloneUrl, onProjectNameChange])

  const cloneValid = source === 'blank' || source === 'local' || (
    cloneUrl.trim().length > 0 &&
    (authType === GIT_AUTH.NONE || (authType === GIT_AUTH.SSH ? sshKey.trim().length > 0 : httpsSecret.trim().length > 0))
  )
  const canSubmit = nameEditable ? projectName.trim().length > 0 && !nameInvalid && cloneValid : true

  const buildGitPayload = useCallback((): GitClonePayload | undefined => {
    if (source === 'blank') return undefined
    if (source === 'local') return { mode: 'local' }
    return {
      mode: 'clone',
      remoteUrl: cloneUrl.trim(),
      authType,
      sshPrivateKey: authType === GIT_AUTH.SSH ? sshKey : undefined,
      httpsUsername: authType === GIT_AUTH.HTTPS ? httpsUser.trim() : undefined,
      httpsSecret: authType === GIT_AUTH.HTTPS ? httpsSecret : undefined,
      shallow,
    }
  }, [source, cloneUrl, authType, sshKey, httpsUser, httpsSecret, shallow])

  const handleTest = useCallback(() => {
    if (!gitTestApi) return
    setTesting(true)
    gitTestApi({
      remoteUrl: cloneUrl.trim(),
      authType,
      sshPrivateKey: authType === GIT_AUTH.SSH ? sshKey : undefined,
      httpsUsername: authType === GIT_AUTH.HTTPS ? httpsUser.trim() : undefined,
      httpsSecret: authType === GIT_AUTH.HTTPS ? httpsSecret : undefined,
    })
      .then((rsp) => toast.success(t('files.gitTestSuccess').replace('{head}', rsp.head?.slice(0, 7) || '')))
      .catch((err: Error) => toast.error(err.message))
      .finally(() => setTesting(false))
  }, [gitTestApi, cloneUrl, authType, sshKey, httpsUser, httpsSecret, t])

  return (
    <Dialog open={mode !== null} onOpenChange={() => onClose()}>
      <DialogContent className="max-w-sm">
        {(mode === 'create' || mode === 'editShare') && (
          <>
            <DialogHeader>
              <DialogTitle>
                {mode === 'create' ? (createTitle || t('files.createProjectTitle')) : (editTitle || t('files.editShareDialogTitle'))}
              </DialogTitle>
              <DialogDescription>
                {mode === 'create' ? (createDesc || t('files.createProjectDesc')) : t('files.editDescription')}
              </DialogDescription>
            </DialogHeader>
            <div className="max-h-[85vh] space-y-3 overflow-y-auto">
              <div>
                <label className="text-xs text-text-muted">{t('files.name')}</label>
                {nameEditable ? (
                  <>
                    <input
                      value={projectName}
                      onChange={(e) => (mode === 'create' ? handleNameChange(e.target.value) : onProjectNameChange?.(e.target.value))}
                      spellCheck={false}
                      placeholder={t('files.projectNamePlaceholder')}
                      className={`mt-1 w-full rounded-md border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:ring-2 focus:ring-ring/30 ${nameInvalid ? 'border-destructive focus:border-destructive' : 'border-border focus:border-ring'}`}
                      autoFocus
                    />
                    {nameInvalid && (
                      <p className="mt-1 text-xs text-destructive">{nameError}</p>
                    )}
                  </>
                ) : (
                  <div className="mt-1 rounded-md border border-border bg-bg-layer-2 px-3 py-2 text-sm text-text-1">
                    {projectName}
                  </div>
                )}
              </div>
              <div>
                <label className="text-xs text-text-muted">{t('files.descriptionLabel')}</label>
                <textarea
                  value={description}
                  onChange={(e) => onDescriptionChange(e.target.value)}
                  spellCheck={false}
                  placeholder={t('files.descriptionPlaceholder')}
                  className="mt-1 h-20 w-full resize-none rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                  autoFocus={mode === 'editShare' && !allowRenameInEdit}
                />
              </div>
              {mode === 'create' && (
                <>
                  <div>
                    <label className="text-xs text-text-muted">{t('files.gitSourceLabel')}</label>
                    <div className="mt-1 flex gap-1 rounded-md border border-border p-0.5">
                      {SOURCE_OPTIONS.map(({ value, labelKey, Icon }) => (
                        <button
                          key={value}
                          onClick={() => setSource(value)}
                          className={cn(
                            'flex flex-1 items-center justify-center gap-1.5 rounded px-2 py-1.5 text-xs transition-colors',
                            source === value ? 'bg-bg-layer-3 font-medium text-text-1' : 'text-text-3 hover:text-text-1',
                          )}
                        >
                          <Icon className={cn('size-3.5', source === value && 'text-interactive')} />
                          {t(labelKey)}
                        </button>
                      ))}
                    </div>
                  </div>
                  {source === 'local' && (
                    <p className="rounded-md border border-border bg-bg-layer-2 px-3 py-2 text-xs text-text-3">
                      {t('files.gitLocalHint')}
                    </p>
                  )}
                  {source === 'clone' && (
                    <div className="space-y-2 rounded-md border border-border bg-bg-layer-2 p-3">
                      <div>
                        <label className="text-xs text-text-muted">{t('files.gitCloneUrl')}</label>
                        <input
                          value={cloneUrl}
                          onChange={(e) => handleCloneUrlChange(e.target.value)}
                          spellCheck={false}
                          placeholder={t('files.gitRemoteUrlPlaceholder')}
                          className="mt-1 w-full rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                        />
                      </div>
                      <div>
                        <label className="text-xs text-text-muted">{t('files.gitAuthType')}</label>
                        <div className="mt-1 flex gap-1 rounded-md border border-border p-0.5">
                          {AUTH_OPTIONS.map((opt) => (
                            <button
                              key={opt.value}
                              onClick={() => setAuthType(opt.value)}
                              className={cn(
                                'flex-1 rounded px-2 py-1 text-xs transition-colors',
                                authType === opt.value ? 'bg-bg-layer-3 font-medium text-text-1' : 'text-text-3 hover:text-text-1',
                              )}
                            >
                              {t(opt.labelKey)}
                            </button>
                          ))}
                        </div>
                      </div>
                      {authType === GIT_AUTH.SSH && (
                        <textarea
                          value={sshKey}
                          onChange={(e) => setSshKey(e.target.value)}
                          rows={3}
                          spellCheck={false}
                          placeholder={t('files.gitSshKeyPlaceholder')}
                          className="w-full resize-none rounded-md border border-border bg-transparent px-3 py-2 font-mono text-xs text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                        />
                      )}
                      {authType === GIT_AUTH.HTTPS && (
                        <>
                          <input
                            value={httpsUser}
                            onChange={(e) => setHttpsUser(e.target.value)}
                            spellCheck={false}
                            placeholder={t('files.gitHttpsUser')}
                            className="w-full rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                          />
                          <input
                            type="password"
                            value={httpsSecret}
                            onChange={(e) => setHttpsSecret(e.target.value)}
                            spellCheck={false}
                            placeholder={t('files.gitHttpsSecretPlaceholder')}
                            className="w-full rounded-md border border-border bg-transparent px-3 py-2 text-sm text-text-1 placeholder:text-text-muted outline-none focus:border-ring focus:ring-2 focus:ring-ring/30"
                          />
                        </>
                      )}
                      <label className="flex items-center gap-2 text-xs text-text-2">
                        <input
                          type="checkbox"
                          checked={shallow}
                          onChange={(e) => setShallow(e.target.checked)}
                          className="size-3.5 cursor-pointer rounded-sm border-border-strong bg-transparent"
                        />
                        {t('files.gitCloneShallow')}
                      </label>
                      {gitTestApi && (
                        <button
                          onClick={handleTest}
                          disabled={testing || !cloneUrl.trim()}
                          className="flex items-center gap-1.5 rounded-md border border-border px-2.5 py-1 text-xs text-text-2 transition-colors hover:bg-bg-hover disabled:opacity-50"
                        >
                          {testing ? <Loader2 className="size-3 animate-spin" /> : <RefreshCw className="size-3" />}
                          {testing ? t('files.gitTesting') : t('files.gitTestConnection')}
                        </button>
                      )}
                    </div>
                  )}
                </>
              )}
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button
                  variant="default"
                  size="default"
                  onClick={() => onConfirm(mode === 'create' ? buildGitPayload() : undefined)}
                  disabled={!canSubmit}
                >
                  {mode === 'create' ? t('common.create') : t('common.save')}
                </Button>
              </div>
            </div>
          </>
        )}

        {mode === 'delete' && (
          <>
            <DialogHeader>
              <DialogTitle>{t('files.confirmDeleteProject')}</DialogTitle>
              <DialogDescription>{deleteDesc || t('files.confirmDeleteProjectDesc')}</DialogDescription>
            </DialogHeader>
            <div className="space-y-3">
              <div className="rounded-md border border-border bg-bg-layer-2 px-3 py-2 text-sm text-text-1">
                {projectName}
              </div>
              <div className="flex justify-end gap-2">
                <Button variant="ghost" size="default" onClick={onClose}>
                  {t('common.cancel')}
                </Button>
                <Button variant="destructive" size="default" onClick={() => onConfirm()}>
                  {t('common.delete')}
                </Button>
              </div>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  )
}
