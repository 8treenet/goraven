import { useState, useEffect } from 'react'
import { useT, t as translate } from '@/i18n'
import { Loader2 } from 'lucide-react'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { adminModelsApi } from '@/api'
import type { AdminModelItem, ProviderItem } from '@/api'
import {
  Drawer, StatusToggle,
} from './shared'
import type { ModelFormData, ModelEditFormData, TestState } from './shared'
import { IconPicker, SearchableCombobox, ApiKeyField, TestIndicator } from './fields'

/* ============================================
   Add Model Drawer
   ============================================ */

export function AddModelDrawer({
  open,
  onClose,
  onSave,
  providers,
  duplicateModel,
}: {
  open: boolean
  onClose: () => void
  onSave: (data: ModelFormData) => Promise<void>
  providers: ProviderItem[]
  duplicateModel: AdminModelItem | null
}) {
  const t = useT()
  const [form, setForm] = useState<ModelFormData>({
    providerDisplayName: '',
    displayName: '',
    providerId: '',
    modelName: '',
    icon: '',
    apiKey: '',
    baseUrl: '',
    proxyUrl: '',
    contextLen: 200,
    extraFields: '',
    isDefault: 0,
    isFlash: 0,
    isVisual: 0,
    remark: '',
  })
  const [testState, setTestState] = useState<TestState>('idle')
  const [testError, setTestError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)

  useEffect(() => {
    if (!open) {
      setForm({
        providerDisplayName: '',
        displayName: '',
        providerId: '',
        modelName: '',
        icon: '',
        apiKey: '',
        baseUrl: '',
        proxyUrl: '',
        contextLen: 200,
        extraFields: '',
        isDefault: 0,
        isFlash: 0,
        isVisual: 0,
        remark: '',
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setDetailLoading(false)
      return
    }

    if (duplicateModel) {
      setForm({
        providerDisplayName: duplicateModel.providerDisplayName,
        displayName: duplicateModel.displayName + '-duplicate',
        providerId: duplicateModel.providerId,
        modelName: duplicateModel.modelName,
        icon: duplicateModel.icon,
        apiKey: '',
        baseUrl: duplicateModel.baseUrl,
        proxyUrl: duplicateModel.proxyUrl,
        contextLen: duplicateModel.contextLen,
        extraFields: '',
        isDefault: 0,
        isFlash: duplicateModel.isFlash,
        isVisual: duplicateModel.isVisual,
        remark: duplicateModel.remark,
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setDetailLoading(true)
      adminModelsApi
        .getModelDetail(duplicateModel.aiModelId)
        .then((detail) => {
          setForm((f) => ({
            ...f,
            apiKey: detail.apiKey || '',
            extraFields: detail.extraFields || '',
          }))
        })
        .catch(() => {
          setTestError(t('adminModels.fetchFailed'))
          setTestState('fail')
        })
        .finally(() => {
          setDetailLoading(false)
        })
    } else {
      setForm({
        providerDisplayName: '',
        displayName: '',
        providerId: '',
        modelName: '',
        icon: '',
        apiKey: '',
        baseUrl: '',
        proxyUrl: '',
        contextLen: 200,
        extraFields: '',
        isDefault: 0,
        isFlash: 0,
        isVisual: 0,
        remark: '',
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setDetailLoading(false)
    }
  }, [open, duplicateModel, t])

  const selectedProvider = providers.find((p) => p.providerId === form.providerId)
  const showApiKey = !selectedProvider || selectedProvider.requireApiKey
  const showBaseUrl = selectedProvider?.requireBaseUrl || false

  const canFetchModels =
    form.providerId.length > 0 &&
    (!showApiKey || form.apiKey.trim().length > 0) &&
    (!showBaseUrl || form.baseUrl.trim().length > 0)

  const canSave =
    form.providerDisplayName.trim().length > 0 &&
    form.providerId.length > 0 &&
    form.modelName.trim().length > 0 &&
    (!showBaseUrl || form.baseUrl.trim().length > 0) &&
    (!showApiKey || form.apiKey.trim().length > 0) &&
    !submitting &&
    !detailLoading

  const handleSubmit = () => {
    setSubmitting(true)
    setTestState('testing')
    onSave(form)
      .then(() => {
        setTestState('success')
        setTimeout(() => {
          setSubmitting(false)
        }, 600)
      })
      .catch((err) => {
        setTestState('fail')
        setTestError(err?.message || translate('adminModels.testFailed'))
        setSubmitting(false)
      })
  }

  return (
    <Drawer open={open} onClose={onClose} title={t('adminModels.addModel')}>
      <div className="flex flex-col gap-4">
        {/* provider select */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.provider')}</Label>
          <select
            value={form.providerId}
            onChange={(e) => {
              const pid = e.target.value
              const p = providers.find((pr) => pr.providerId === pid)
              setForm((f) => ({
                ...f,
                providerId: pid,
                modelName: '',
                providerDisplayName: p ? p.providerDisplayNameZh : f.providerDisplayName,
                icon: p && p.icon ? p.icon : f.icon,
                baseUrl: p ? p.defaultBaseUrl : '',
              }))
            }}
            className="h-8 rounded-lg border border-border-strong bg-transparent px-2.5 text-sm text-text-1 outline-none focus:border-ring focus:ring-3 focus:ring-ring/50"
          >
            <option value="">{t('adminModels.selectProvider')}</option>
            {providers.map((p) => (
              <option key={p.providerId} value={p.providerId}>
                {p.providerDisplayNameZh}
              </option>
            ))}
          </select>
        </div>

        {/* display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.provider')}
            <span className="ml-0.5 text-destructive">*</span>
          </Label>
          <Input
            value={form.providerDisplayName}
            onChange={(e) => setForm((f) => ({ ...f, providerDisplayName: e.target.value }))}
            placeholder="DeepSeek"
            className="h-8 text-sm"
          />
          <p className="text-xs text-text-muted">{t('adminModels.nameHint')}</p>
        </div>

        {/* model display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.modelDisplayName')}
          </Label>
          <Input
            value={form.displayName}
            onChange={(e) => setForm((f) => ({ ...f, displayName: e.target.value }))}
            placeholder="deepseek-v4-pro"
            className="h-8 text-sm"
          />
          <p className="text-xs text-text-muted">{t('adminModels.displayNameHint')}</p>
        </div>

        {/* icon */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('common.icon')}</Label>
          <IconPicker
            value={form.icon}
            onChange={(path) => setForm((f) => ({ ...f, icon: path }))}
          />
        </div>

        {/* api keys */}
        {showApiKey && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">
              API Key
              {selectedProvider && <span className="ml-0.5 text-destructive">*</span>}
            </Label>
            <ApiKeyField
              value={form.apiKey}
              onChange={(key) => setForm((f) => ({ ...f, apiKey: key }))}
              disabled={false}
            />
          </div>
        )}

        {/* base url */}
        {showBaseUrl && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">
              Base URL
              <span className="ml-0.5 text-destructive">*</span>
            </Label>
            <Input
              value={form.baseUrl}
              onChange={(e) => setForm((f) => ({ ...f, baseUrl: e.target.value }))}
              placeholder="https://api.deepseek.com"
              className="h-8 text-sm font-mono"
            />
          </div>
        )}

        {/* model name searchable combobox */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('common.model')}
            <span className="ml-0.5 text-destructive">*</span>
          </Label>
          <SearchableCombobox
            value={form.modelName}
            onChange={(v) => setForm((f) => ({ ...f, modelName: v }))}
            providerId={form.providerId}
            canFetch={canFetchModels}
            apiKey={form.apiKey}
            baseUrl={form.baseUrl}
          />
        </div>

        {/* proxy url */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.networkProxy')}</Label>
          <Input
            value={form.proxyUrl}
            onChange={(e) => setForm((f) => ({ ...f, proxyUrl: e.target.value }))}
            placeholder={`http://127.0.0.1:7890（${t('common.optional')}）`}
            className="h-8 text-sm font-mono"
          />
        </div>

        {/* context len */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.contextKB')}</Label>
          <Input
            type="number"
            value={String(form.contextLen)}
            onChange={(e) => setForm((f) => ({ ...f, contextLen: Number(e.target.value) || 0 }))}
            className="h-8 text-sm"
          />
        </div>

        {/* flags */}
        <div className="flex flex-col gap-2">
          <Label className="text-xs text-text-2">{t('adminModels.modelLabels')}</Label>
          <div className="flex flex-col gap-2">
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.defaultModel')}</span>
              <StatusToggle
                checked={form.isDefault === 1}
                onChange={(v) => setForm((f) => ({ ...f, isDefault: v ? 1 : 0 }))}
              />
            </label>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.flashModel')}</span>
              <StatusToggle
                checked={form.isFlash === 1}
                onChange={(v) => setForm((f) => ({ ...f, isFlash: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.flashHint')}</p>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.multimodalModel')}</span>
              <StatusToggle
                checked={form.isVisual === 1}
                onChange={(v) => setForm((f) => ({ ...f, isVisual: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.multimodalHint')}</p>
          </div>
        </div>

        {/* extra fields - only for openai_compatible / claude_compatible */}
        {(form.providerId === 'openai_compatible' || form.providerId === 'claude_compatible') && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">extra_body</Label>
            <textarea
              value={form.extraFields}
              onChange={(e) => setForm((f) => ({ ...f, extraFields: e.target.value }))}
              placeholder='{"thinking":{"type":"enabled"}}'
              rows={3}
              className="min-h-[60px] w-full rounded-lg border border-border-strong bg-transparent px-2.5 py-1.5 font-mono text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-ring focus:ring-3 focus:ring-ring/50"
            />
          </div>
        )}

        {/* remark */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.remark')}</Label>
          <Input
            value={form.remark}
            onChange={(e) => setForm((f) => ({ ...f, remark: e.target.value }))}
            placeholder={t('common.optional')}
            className="h-8 text-sm"
          />
        </div>

        {/* test indicator */}
        <TestIndicator state={testState} errorMsg={testError} />

        {/* buttons */}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} disabled={submitting} className="text-highlight hover:text-highlight/80">
            {t('common.cancel')}
          </Button>
          <Button variant="default" size="default" disabled={!canSave} onClick={handleSubmit} className="text-highlight">
            {submitting ? (
              <>
                <Loader2 className="size-3.5 animate-spin" />
                {t('common.testing')}
              </>
            ) : (
              t('adminModels.saveAndTest')
            )}
          </Button>
        </div>
      </div>
    </Drawer>
  )
}

/* ============================================
   Edit Model Drawer
   ============================================ */

export function EditModelDrawer({
  open,
  onClose,
  onSave,
  model,
}: {
  open: boolean
  onClose: () => void
  onSave: (data: ModelEditFormData) => Promise<void>
  model: AdminModelItem | null
}) {
  const t = useT()
  const [form, setForm] = useState<ModelEditFormData>({
    providerDisplayName: '',
    displayName: '',
    modelName: '',
    icon: '',
    apiKey: '',
    baseUrl: '',
    proxyUrl: '',
    contextLen: 200,
    extraFields: '',
    isDefault: 0,
    isFlash: 0,
    isVisual: 0,
    status: 1,
    remark: '',
  })
  const [testState, setTestState] = useState<TestState>('idle')
  const [testError, setTestError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [connectionChanged, setConnectionChanged] = useState(false)
  const [detailLoading, setDetailLoading] = useState(false)

  useEffect(() => {
    if (model && open) {
      setForm({
        providerDisplayName: model.providerDisplayName,
        displayName: model.displayName,
        modelName: model.modelName,
        icon: model.icon,
        apiKey: '',
        baseUrl: model.baseUrl,
        proxyUrl: model.proxyUrl,
        contextLen: model.contextLen,
        extraFields: '',
        isDefault: model.isDefault,
        isFlash: model.isFlash,
        isVisual: model.isVisual,
        status: model.status,
        remark: model.remark,
      })
      setTestState('idle')
      setTestError('')
      setSubmitting(false)
      setConnectionChanged(false)
      setDetailLoading(true)
      adminModelsApi
        .getModelDetail(model.aiModelId)
        .then((detail) => {
          setForm((f) => ({
            ...f,
            apiKey: detail.apiKey || '',
            extraFields: detail.extraFields || '',
          }))
        })
        .catch(() => {
          setTestError(t('adminModels.fetchFailed'))
          setTestState('fail')
        })
        .finally(() => {
          setDetailLoading(false)
        })
    }
  }, [model, open, t])

  if (!model) return null

  const canSave =
    form.providerDisplayName.trim().length > 0 &&
    form.modelName.trim().length > 0 &&
    !submitting &&
    !detailLoading

  const handleSubmit = () => {
    setSubmitting(true)
    if (connectionChanged) {
      setTestState('testing')
    }
    const submitData: any = { ...form }
    if (!connectionChanged) {
      delete submitData.apiKey
    }
    onSave(submitData)
      .then(() => {
        if (connectionChanged) {
          setTestState('success')
        }
        setTimeout(() => {
          setSubmitting(false)
        }, 600)
      })
      .catch((err) => {
        if (connectionChanged) {
          setTestState('fail')
          setTestError(err?.message || translate('adminModels.testFailed'))
        }
        setSubmitting(false)
      })
  }

  const markConnectionChanged = () => setConnectionChanged(true)

  return (
    <Drawer open={open} onClose={onClose} title={t('adminModels.editModel')}>
      <div className="flex flex-col gap-4">
        {/* provider id (disabled) */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.provider')}</Label>
          <Input value={model.providerId} disabled className="h-8 text-sm opacity-50" />
          <p className="text-xs text-text-muted">{t('adminModels.providerReadonly')}</p>
        </div>

        {/* display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.provider')}
            <span className="ml-0.5 text-destructive">*</span>
          </Label>
          <Input
            value={form.providerDisplayName}
            onChange={(e) => setForm((f) => ({ ...f, providerDisplayName: e.target.value }))}
            className="h-8 text-sm"
          />
        </div>

        {/* model display name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('adminModels.modelDisplayName')}
          </Label>
          <Input
            value={form.displayName}
            onChange={(e) => setForm((f) => ({ ...f, displayName: e.target.value }))}
            className="h-8 text-sm"
          />
          <p className="text-xs text-text-muted">{t('adminModels.displayNameHint')}</p>
        </div>

        {/* icon */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('common.icon')}</Label>
          <IconPicker
            value={form.icon}
            onChange={(path) => setForm((f) => ({ ...f, icon: path }))}
          />
        </div>

        {/* api keys (with copy) */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">API Key</Label>
          <ApiKeyField
            value={form.apiKey}
            onChange={(key) => {
              setForm((f) => ({ ...f, apiKey: key }))
              markConnectionChanged()
            }}
            disabled={detailLoading}
          />
          {detailLoading && (
            <div className="flex items-center gap-2 text-xs text-text-3">
              <Loader2 className="size-3 animate-spin" />
              {t('adminModels.testingConnection')}
            </div>
          )}
        </div>

        {/* base url */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">Base URL</Label>
          <Input
            value={form.baseUrl}
            onChange={(e) => {
              setForm((f) => ({ ...f, baseUrl: e.target.value }))
              markConnectionChanged()
            }}
            className="h-8 text-sm font-mono"
            placeholder="https://api.deepseek.com"
          />
        </div>

        {/* model name */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">
            {t('common.model')}
            <span className="ml-0.5 text-destructive">*</span>
          </Label>
          <Input
            value={form.modelName}
            onChange={(e) => setForm((f) => ({ ...f, modelName: e.target.value }))}
            className="h-8 text-sm"
          />
        </div>

        {/* proxy url */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.networkProxy')}</Label>
          <Input
            value={form.proxyUrl}
            onChange={(e) => {
              setForm((f) => ({ ...f, proxyUrl: e.target.value }))
              markConnectionChanged()
            }}
            className="h-8 text-sm font-mono"
          />
        </div>

        {/* context len */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.contextKB')}</Label>
          <Input
            type="number"
            value={String(form.contextLen)}
            onChange={(e) => setForm((f) => ({ ...f, contextLen: Number(e.target.value) || 0 }))}
            className="h-8 text-sm"
          />
        </div>

        {/* flags */}
        <div className="flex flex-col gap-2">
          <Label className="text-xs text-text-2">{t('adminModels.modelLabels')}</Label>
          <div className="flex flex-col gap-2">
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.defaultModel')}</span>
              <StatusToggle
                checked={form.isDefault === 1}
                onChange={(v) => setForm((f) => ({ ...f, isDefault: v ? 1 : 0 }))}
              />
            </label>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.flashModel')}</span>
              <StatusToggle
                checked={form.isFlash === 1}
                onChange={(v) => setForm((f) => ({ ...f, isFlash: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.flashHint')}</p>
            <label className="flex items-center justify-between">
              <span className="text-sm text-text-2">
                {t('adminModels.multimodalModel')}</span>
              <StatusToggle
                checked={form.isVisual === 1}
                onChange={(v) => setForm((f) => ({ ...f, isVisual: v ? 1 : 0 }))}
              />
            </label>
            <p className="text-xs text-text-muted">{t('adminModels.multimodalHint')}</p>
          </div>
        </div>

        {/* extra fields - only for openai_compatible / claude_compatible */}
        {(model.providerId === 'openai_compatible' || model.providerId === 'claude_compatible') && (
          <div className="flex flex-col gap-1.5">
            <Label className="text-xs text-text-2">extra_body</Label>
            <textarea
              value={form.extraFields}
              onChange={(e) => setForm((f) => ({ ...f, extraFields: e.target.value }))}
              rows={3}
              className="min-h-[60px] w-full rounded-lg border border-border-strong bg-transparent px-2.5 py-1.5 font-mono text-xs text-text-1 outline-none placeholder:text-text-muted focus:border-ring focus:ring-3 focus:ring-ring/50"
            />
          </div>
        )}

        {/* remark */}
        <div className="flex flex-col gap-1.5">
          <Label className="text-xs text-text-2">{t('adminModels.remark')}</Label>
          <Input
            value={form.remark}
            onChange={(e) => setForm((f) => ({ ...f, remark: e.target.value }))}
            className="h-8 text-sm"
          />
        </div>

        {/* test indicator */}
        <TestIndicator state={testState} errorMsg={testError} />

        {/* buttons */}
        <div className="flex justify-end gap-2">
          <Button variant="ghost" size="default" onClick={onClose} disabled={submitting} className="text-highlight hover:text-highlight/80">
            {t('common.cancel')}
          </Button>
          <Button variant="default" size="default" disabled={!canSave} onClick={handleSubmit} className="text-highlight">
            {submitting ? (
              <>
                <Loader2 className="size-3.5 animate-spin" />
                {t('common.saving')}
              </>
            ) : (
              t('common.save')
            )}
          </Button>
        </div>
      </div>
    </Drawer>
  )
}

/* ============================================
   Delete Confirm Dialog
   ============================================ */

export function DeleteModelDialog({
  open,
  onClose,
  onConfirm,
  modelName,
}: {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  modelName: string
}) {
  const t = useT()
  return (
    <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>{t('adminModels.deleteModel')}</DialogTitle>
          <DialogDescription>
            {t('adminModels.deleteConfirm').replace('{name}', modelName)}
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
