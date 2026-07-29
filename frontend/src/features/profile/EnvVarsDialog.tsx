import { useCallback, useEffect, useState } from 'react'
import { toast } from 'sonner'
import { Check, Pencil, Plus, Trash2, X } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useT, t as translate } from '@/i18n'
import { getProfile, createProfile, updateProfile, deleteProfile } from '@/api/files'
import type { ProfileEntry } from '@/api/types'

const KEY_REGEX = /^[A-Za-z0-9_]+$/

interface EnvVarsDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function EnvVarsDialog({ open, onOpenChange }: EnvVarsDialogProps) {
  const t = useT()
  const [items, setItems] = useState<ProfileEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [adding, setAdding] = useState(false)
  const [newKey, setNewKey] = useState('')
  const [newValue, setNewValue] = useState('')
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(() => {
    setLoading(true)
    setError(null)
    getProfile()
      .then((data) => setItems(data.items))
      .catch(() => setError('Failed to load'))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (open) {
      setAdding(false)
      setEditingKey(null)
      setError(null)
      load()
    }
  }, [open, load])

  const resetForm = useCallback(() => {
    setAdding(false)
    setNewKey('')
    setNewValue('')
    setEditingKey(null)
    setEditValue('')
    setError(null)
  }, [])

  const handleAdd = useCallback(() => {
    if (!newKey.trim() || !KEY_REGEX.test(newKey.trim())) {
      setError(t('profile.envVarKeyInvalid'))
      return
    }
    setError(null)
    createProfile(newKey.trim(), newValue)
      .then(() => {
        resetForm()
        toast.success(t('profile.envVarAdded'))
        load()
      })
      .catch((err: Error) => {
        toast.error(err.message || t('profile.envVarExists').replace('{key}', newKey.trim()))
      })
  }, [newKey, newValue, resetForm, load, t])

  const startEdit = useCallback((entry: ProfileEntry) => {
    setAdding(false)
    setEditingKey(entry.key)
    setEditValue(entry.value)
    setError(null)
  }, [])

  const handleUpdate = useCallback((key: string) => {
    setError(null)
    updateProfile(key, editValue)
      .then(() => {
        resetForm()
        toast.success(t('profile.envVarUpdated'))
        load()
      })
      .catch((err: Error) => {
        toast.error(err.message || t('profile.envVarNotFound').replace('{key}', key))
      })
  }, [editValue, resetForm, load, t])

  const handleDelete = useCallback((key: string) => {
    deleteProfile(key)
      .then(() => {
        toast.success(t('profile.envVarDeleted'))
        load()
      })
      .catch((err: Error) => {
        toast.error(err.message || translate('common.deleteFailed'))
      })
  }, [load, t])

  return (
    <Dialog open={open} onOpenChange={() => onOpenChange(false)}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>{t('profile.envVarsTitle')}</DialogTitle>
          <DialogDescription>{t('profile.envVarsDesc')}</DialogDescription>
        </DialogHeader>
        <div className="space-y-3">
          <p className="text-xs text-text-3 leading-relaxed">
            {t('profile.envVarsHint')}
          </p>

          {loading ? (
            <div className="rounded-md border border-border p-1.5 space-y-1">
              {[1, 2, 3].map((i) => (
                <div key={i} className="flex h-8 items-center gap-2 px-2 rounded-md bg-bg-layer-2">
                  <div className="h-4 flex-1 animate-pulse rounded bg-bg-layer-3" />
                  <div className="h-4 w-20 animate-pulse rounded bg-bg-layer-3" />
                </div>
              ))}
            </div>
          ) : items.length === 0 && !adding ? (
            <div className="flex flex-col items-center gap-2 py-6">
              <p className="text-sm text-text-2">{t('profile.envVarEmpty')}</p>
              <p className="text-sm text-text-3">{t('profile.envVarEmptyHint')}</p>
            </div>
          ) : (
            <div className="max-h-[400px] overflow-auto rounded-md border border-border p-1.5 space-y-1">
              {adding && (
                <div className="flex items-center gap-2 px-2 py-1.5 rounded-md bg-bg-layer-2">
                  <Input
                    value={newKey}
                    onChange={(e) => { setNewKey(e.target.value); setError(null) }}
                    onKeyDown={(e) => { if (e.key === 'Enter') handleAdd() }}
                    placeholder={t('profile.envVarKeyPlaceholder')}
                    className="h-7 flex-[2] text-sm"
                    autoFocus
                  />
                  <Input
                    value={newValue}
                    onChange={(e) => setNewValue(e.target.value)}
                    onKeyDown={(e) => { if (e.key === 'Enter') handleAdd() }}
                    placeholder={t('profile.envVarValuePlaceholder')}
                    className="h-7 flex-[3] text-sm"
                  />
                  <Button variant="ghost" size="icon" className="size-7 shrink-0" onClick={handleAdd}>
                    <Check className="size-3.5" />
                  </Button>
                  <Button variant="ghost" size="icon" className="size-7 shrink-0" onClick={resetForm}>
                    <X className="size-3.5" />
                  </Button>
                </div>
              )}
              {error && <p className="px-2 pb-1 text-xs text-text-3">{error}</p>}

              {items.map((entry) => (
                <div
                  key={entry.key}
                  className="flex items-center gap-2 px-2 py-1.5 rounded-md bg-bg-layer-2"
                >
                  {editingKey === entry.key ? (
                    <>
                      <span className="flex-[2] text-sm text-text-1 font-mono truncate">
                        {entry.key}
                      </span>
                      <Input
                        value={editValue}
                        onChange={(e) => setEditValue(e.target.value)}
                        onKeyDown={(e) => {
                          if (e.key === 'Enter') handleUpdate(entry.key)
                          if (e.key === 'Escape') resetForm()
                        }}
                        placeholder={t('profile.envVarValuePlaceholder')}
                        className="h-7 flex-[3] text-sm"
                        autoFocus
                      />
                      <Button variant="ghost" size="icon" className="size-7 shrink-0" onClick={() => handleUpdate(entry.key)}>
                        <Check className="size-3.5" />
                      </Button>
                      <Button variant="ghost" size="icon" className="size-7 shrink-0" onClick={resetForm}>
                        <X className="size-3.5" />
                      </Button>
                    </>
                  ) : (
                    <>
                      <span className="flex-[2] text-sm text-text-1 font-mono truncate">
                        {entry.key}
                      </span>
                      <span className="flex-[3] text-sm text-text-2 truncate">
                        {entry.value || <span className="text-text-muted">--</span>}
                      </span>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7 shrink-0 text-text-muted hover:text-text-2"
                        onClick={() => startEdit(entry)}
                      >
                        <Pencil className="size-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="size-7 shrink-0 text-text-muted hover:text-text-2"
                        onClick={() => handleDelete(entry.key)}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    </>
                  )}
                </div>
              ))}
            </div>
          )}

          {!adding && !loading && (
            <Button
              variant="ghost"
              size="default"
              className="w-full"
              onClick={() => { setAdding(true); setNewKey(''); setNewValue(''); setError(null) }}
            >
              <Plus className="size-4" />
              {t('profile.envVarAdd')}
            </Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
