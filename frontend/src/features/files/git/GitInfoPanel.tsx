import { useT } from '@/i18n'
import type { GitStatus } from '@/api/types'
import { AUTH_OPTIONS, authNeedsCredential } from './git-helpers'

interface GitInfoPanelProps {
  status: GitStatus
}

/** 项目 Git 配置（只读）：创建时确定，不支持后续修改 */
export function GitInfoPanel({ status }: GitInfoPanelProps) {
  const t = useT()
  const remoteUrl = status.remote.url.trim()
  const authLabel = AUTH_OPTIONS.find((opt) => opt.value === status.remote.authType)

  return (
    <div className="border-t border-border px-5 py-4">
      <h3 className="mb-2 text-xs font-medium text-text-2">{t('files.gitConfig')}</h3>
      <dl className="space-y-1.5 text-xs">
        <div className="flex gap-2">
          <dt className="w-20 shrink-0 text-text-3">{t('files.gitRemoteUrl')}</dt>
          <dd className="min-w-0 break-all font-mono text-text-2">
            {remoteUrl || t('files.gitLocalRepo')}
          </dd>
        </div>
        {remoteUrl && (
          <div className="flex gap-2">
            <dt className="w-20 shrink-0 text-text-3">{t('files.gitAuthType')}</dt>
            <dd className="text-text-2">{authLabel ? t(authLabel.labelKey) : '-'}</dd>
          </div>
        )}
        {remoteUrl && authNeedsCredential(status.remote.authType) && (
          <div className="flex gap-2">
            <dt className="w-20 shrink-0 text-text-3">{t('files.gitCredential')}</dt>
            <dd className={status.remote.hasCredential ? 'text-text-2' : 'text-warning'}>
              {status.remote.hasCredential ? t('files.gitCredentialSaved') : t('files.gitCredentialNone')}
            </dd>
          </div>
        )}
      </dl>
      <p className="mt-3 text-xs text-text-3">{t('files.gitAutoPushHint')}</p>
    </div>
  )
}
