import { useState, useCallback, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { useAppStore } from '@/stores/app-store'
import { useT, t as translate } from '@/i18n'
import type { TranslationKey } from '@/i18n'
import { authApi } from '@/api'

/* ============================================
   Types
   ============================================ */

type DbType = 'sqlite' | 'mysql' | 'pg'
type CacheType = 'local' | 'redis'

interface FormData {
  language: 'zh' | 'en'
  domain: string
  username: string
  password: string
  confirmPassword: string
  email: string
  dbType: DbType
  dbAddr: string
  dbPort: number
  dbUser: string
  dbPass: string
  dbName: string
  cacheType: CacheType
  redisAddr: string
  redisPort: number
  redisPass: string
  redisDB: number
}

const INITIAL: FormData = {
  language: 'zh',
  domain: '',
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  dbType: 'sqlite',
  dbAddr: '',
  dbPort: 3306,
  dbUser: '',
  dbPass: '',
  dbName: '',
  cacheType: 'local',
  redisAddr: '127.0.0.1',
  redisPort: 6379,
  redisPass: '',
  redisDB: 0,
}

export function Component() {
  const navigate = useNavigate()
  const t = useT()
  const stepLabels = useMemo<TranslationKey[]>(() => [
    'install.stepLanguage',
    'install.stepDomain',
    'install.stepAdmin',
    'install.stepDb',
    'install.stepCache',
  ], [])
  const [step, setStep] = useState(0)
  const [data, setData] = useState<FormData>(INITIAL)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [dbTesting, setDbTesting] = useState(false)
  const [dbTestResult, setDbTestResult] = useState<'idle' | 'ok' | 'fail'>('idle')
  const [redisTesting, setRedisTesting] = useState(false)
  const [redisTestResult, setRedisTestResult] = useState<'idle' | 'ok' | 'fail'>('idle')
  const [submitting, setSubmitting] = useState(false)
  const [succeeded, setSucceeded] = useState(false)
  const [submitError, setSubmitError] = useState('')

  const update = useCallback((patch: Partial<FormData>) => {
    setData((d) => ({ ...d, ...patch }))
    setErrors({})
    setSubmitError('')
  }, [])

  const setField = useCallback((key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
    update({ [key]: e.target.value })
  }, [update])

  const setFieldNumber = useCallback((key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = parseInt(e.target.value, 10)
    update({ [key]: isNaN(val) ? 0 : val })
  }, [update])

  /* ---- validation per step ---- */

  const validateStep = useCallback((s: number): boolean => {
    const errs: Record<string, string> = {}
    switch (s) {
      case 0:
        if (!data.language) errs.language = translate('install.errSelectLanguage')
        break
      case 2:
        if (!data.username || data.username.length < 8) errs.username = translate('install.errUsernameLen')
        if (!/^[a-zA-Z0-9]+$/.test(data.username)) errs.username = translate('install.errUsernameAlnum')
        if (data.username.length > 64) errs.username = translate('install.errUsernameMax')
        if (!data.password || data.password.length < 8) errs.password = translate('install.errPasswordLen')
        if (!/^(?=.*[a-zA-Z])(?=.*\d)/.test(data.password)) errs.password = translate('install.errPasswordComplex')
        if (data.password.length > 256) errs.password = translate('install.errPasswordMax')
        if (data.password !== data.confirmPassword) errs.confirmPassword = translate('install.errPasswordMismatch')
        break
      case 3:
        if (data.dbType !== 'sqlite') {
          if (!data.dbAddr) errs.dbAddr = translate('install.errHostRequired')
          if (!data.dbPort) errs.dbPort = translate('install.errPortRequired')
          if (!data.dbUser) errs.dbUser = translate('install.errDbUserRequired')
          if (!data.dbPass) errs.dbPass = translate('install.errDbPassRequired')
          if (!data.dbName) errs.dbName = translate('install.errDbNameRequired')
        }
        break
      case 4:
        if (data.cacheType === 'redis') {
          if (!data.redisAddr) errs.redisAddr = translate('install.errRedisAddrRequired')
          if (!data.redisPort) errs.redisPort = translate('install.errPortRequired')
          if (!data.redisPass) errs.redisPass = translate('install.errRedisPassRequired')
        }
        break
    }
    setErrors(errs)
    return Object.keys(errs).length === 0
  }, [data])

  /* ---- navigation ---- */

  const handleNext = useCallback(() => {
    if (!validateStep(step)) return
    setSubmitError('')
    if (step < 4) {
      setStep((s) => s + 1)
      setDbTestResult('idle')
      setRedisTestResult('idle')
    }
  }, [step, validateStep])

  const handleBack = useCallback(() => {
    if (step > 0) {
      setStep((s) => s - 1)
      setDbTestResult('idle')
      setRedisTestResult('idle')
      setSubmitError('')
    }
  }, [step])

  /* ---- connection tests ---- */

  const testDb = useCallback(async () => {
    if (!validateStep(3)) return
    setDbTesting(true)
    setDbTestResult('idle')
    try {
      await authApi.checkDb({ dbType: data.dbType, dbAddr: data.dbAddr, dbPort: data.dbPort, dbUser: data.dbUser, dbPass: data.dbPass, dbName: data.dbName })
      setDbTestResult('ok')
    } catch {
      setDbTestResult('fail')
      setErrors({ dbTest: translate('common.connectionFailed') })
    }
    setDbTesting(false)
  }, [data, validateStep])

  const testRedis = useCallback(async () => {
    if (!validateStep(4)) return
    setRedisTesting(true)
    setRedisTestResult('idle')
    try {
      await authApi.checkRedis({ redisAddr: data.redisAddr, redisPort: data.redisPort, redisPass: data.redisPass, redisDB: data.redisDB })
      setRedisTestResult('ok')
    } catch {
      setRedisTestResult('fail')
      setErrors({ redisTest: translate('common.connectionFailed') })
    }
    setRedisTesting(false)
  }, [data, validateStep])

  /* ---- submit ---- */

  const handleSubmit = useCallback(async () => {
    if (!validateStep(step)) return
    setSubmitting(true)
    setSubmitError('')
    try {
      await authApi.initSystem({
        language: data.language,
        domain: data.domain,
        username: data.username,
        password: data.password,
        email: data.email,
        dbType: data.dbType,
        dbAddr: data.dbAddr,
        dbPort: data.dbPort,
        dbUser: data.dbUser,
        dbPass: data.dbPass,
        dbName: data.dbName,
        cacheType: data.cacheType,
        redisAddr: data.redisAddr,
        redisPort: data.redisPort,
        redisPass: data.redisPass,
        redisDB: data.redisDB,
      })
      useAppStore.getState().setLanguage(data.language)
      setSucceeded(true)
      await new Promise((r) => setTimeout(r, 2500))
      for (let i = 0; i < 30; i++) {
        await new Promise((r) => setTimeout(r, 1000))
        try {
          const res = await fetch('/api/ping')
          const text = await res.text()

          if (text === 'pong') break
        } catch { /* keep polling */ }
      }
      navigate('/login', { replace: true })
    } catch {
      setSubmitError(translate('install.errInitFailed'))
      setSubmitting(false)
    }
  }, [data, step, validateStep, navigate])

  /* ---- render ---- */

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-bg-base px-4">
      <div className="absolute left-6 top-6 flex items-center gap-2">
        <img src="/favicon.svg" alt="Raven" className="h-7 w-7" />
        <span className="text-sm font-semibold text-text-2">Raven</span>
      </div>
      <div className="w-full max-w-md">

        {/* Logo */}
        <div className="mb-10 flex justify-center">
          <h1 className="text-xl font-semibold text-text-1 tracking-tight">Raven</h1>
        </div>

        {/* Step indicator */}
        <div className="mb-12 text-center">
          <p className="mb-2 text-sm text-text-3">
            {String(step + 1).padStart(2, '0')} / {String(stepLabels.length).padStart(2, '0')}
          </p>
          <h1 className="text-2xl font-semibold text-text-1">
            {t(stepLabels[step])}
          </h1>
        </div>

        {/* Form area */}
        {succeeded ? (
          <div className="space-y-6 text-center">
            <div className="mx-auto h-3 w-3 rounded-full bg-highlight animate-pulse" />
            <p className="text-lg text-text-1">{t('install.systemInitializing')}</p>
            <p className="text-sm text-text-3">{t('install.autoRedirect')}</p>
          </div>
        ) : (
          <div className="space-y-6">
            {step === 0 && <LanguageStep value={data.language} onChange={(v) => { update({ language: v }); useAppStore.getState().setLanguage(v) }} />}
            {step === 1 && <DomainStep value={data.domain} onChange={setField('domain')} error={errors.domain} />}
            {step === 2 && <AdminStep data={data} onField={setField} errors={errors} />}
            {step === 3 && (
              <DbStep
                data={data}
                onField={setField}
                onFieldNumber={setFieldNumber}
                onType={(v) => update({ dbType: v as DbType })}
                errors={errors}
                testing={dbTesting}
                testResult={dbTestResult}
                onTest={testDb}
              />
            )}
            {step === 4 && (
              <CacheStep
                data={data}
                onField={setField}
                onFieldNumber={setFieldNumber}
                onType={(v) => update({ cacheType: v as CacheType })}
                errors={errors}
                testing={redisTesting}
                testResult={redisTestResult}
                onTest={testRedis}
              />
            )}

            {/* Submit error */}
            {submitError && (
              <p className="text-center text-sm text-text-3">{submitError}</p>
            )}

            {/* Navigation */}
            <div className="flex items-center justify-between pt-4">
              <button
                type="button"
                onClick={handleBack}
                disabled={step === 0 || submitting}
                className={cn(
                  'text-sm text-text-3 transition-colors hover:text-text-2',
                  (step === 0 || submitting) && 'pointer-events-none opacity-30',
                )}
              >
                {t('common.previous')}
              </button>

              {step < 4 ? (
                <button
                  type="button"
                  onClick={handleNext}
                  className="rounded-md bg-highlight px-6 py-2 text-sm text-highlight-fg transition-colors hover:opacity-90"
                >
                  {t('common.next')}
                </button>
              ) : (
                <button
                  type="button"
                  onClick={handleSubmit}
                  disabled={submitting}
                  className={cn(
                    'rounded-md bg-highlight px-6 py-2 text-sm text-highlight-fg transition-colors hover:opacity-90',
                    submitting && 'pointer-events-none opacity-50',
                  )}
                >
                  {submitting ? t('common.submitting') : t('common.submit')}
                </button>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

/* ============================================
   Step components
   ============================================ */

function LanguageStep({ value, onChange }: { value: string; onChange: (v: 'zh' | 'en') => void }) {
  const options = [
    { value: 'zh' as const, label: '中文', sub: '简体中文' },
    { value: 'en' as const, label: 'English', sub: 'English' },
  ]
  return (
    <div className="space-y-3">
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={cn(
            'w-full rounded-md px-5 py-4 text-left transition-colors',
            value === opt.value
              ? 'bg-bg-layer-3 ring-1 ring-bg-hover'
              : 'bg-bg-layer-2 hover:bg-bg-layer-3',
          )}
        >
          <span className="text-sm text-text-1">{opt.label}</span>
          <span className="ml-3 text-xs text-text-3">{opt.sub}</span>
        </button>
      ))}
    </div>
  )
}

function DomainStep({ value, onChange, error }: { value: string; onChange: (e: React.ChangeEvent<HTMLInputElement>) => void; error?: string }) {
  const t = useT()
  return (
    <div className="space-y-2">
      <Label className="text-xs text-text-2">{t('install.domainLabel')}</Label>
      <Input
        value={value}
        onChange={onChange}
        placeholder="https://goraven.dev"
        className="border-0 bg-bg-layer-2 text-sm text-text-1 placeholder:text-text-muted"
      />
      {error && <p className="text-xs text-text-3">{error}</p>}
      <p className="text-xs text-text-muted">{t('install.domainHint')}</p>
    </div>
  )
}

function AdminStep({ data, onField, errors }: { data: FormData; onField: (key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => void; errors: Record<string, string> }) {
  const t = useT()
  return (
    <div className="space-y-4">
      <Field label={t('common.username')} value={data.username} onChange={onField('username')} placeholder={t('install.usernamePlaceholder')} error={errors.username} filter="alphanumeric" />
      <Field label={t('common.password')} value={data.password} onChange={onField('password')} placeholder={t('install.passwordPlaceholder')} error={errors.password} type="password" />
      <Field label={t('common.confirmPassword')} value={data.confirmPassword} onChange={onField('confirmPassword')} placeholder={t('install.confirmPasswordPlaceholder')} error={errors.confirmPassword} type="password" />
      <Field label={t('install.emailOptional')} value={data.email} onChange={onField('email')} placeholder="admin@example.com" />
    </div>
  )
}

function DbStep({ data, onField, onFieldNumber, onType, errors, testing, testResult, onTest }: {
  data: FormData
  onField: (key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => void
  onFieldNumber: (key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => void
  onType: (v: string) => void
  errors: Record<string, string>
  testing: boolean
  testResult: 'idle' | 'ok' | 'fail'
  onTest: () => void
}) {
  const t = useT()
  return (
    <div className="space-y-4">
      <TypeSelect label={t('install.dbType')} value={data.dbType} onChange={onType} options={[
        { value: 'sqlite', label: 'SQLite' },
        { value: 'mysql', label: 'MySQL' },
        { value: 'pg', label: 'PostgreSQL' },
      ]} />
      {data.dbType !== 'sqlite' && (
        <>
          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <Field label={t('common.host')} value={data.dbAddr} onChange={onField('dbAddr')} placeholder="127.0.0.1" error={errors.dbAddr} />
            </div>
            <Field label={t('common.port')} value={String(data.dbPort)} onChange={onFieldNumber('dbPort')} placeholder="3306" error={errors.dbPort} />
          </div>
          <Field label={t('common.username')} value={data.dbUser} onChange={onField('dbUser')} placeholder="root" error={errors.dbUser} />
          <Field label={t('common.password')} value={data.dbPass} onChange={onField('dbPass')} placeholder={t('install.dbPassPlaceholder')} type="password" error={errors.dbPass} />
          <Field label={t('install.dbName')} value={data.dbName} onChange={onField('dbName')} placeholder="raven" error={errors.dbName} />
          <TestButton label={t('common.testConnection')} testing={testing} result={testResult} onTest={onTest} />
          {errors.dbTest && <p className="text-xs text-text-3">{errors.dbTest}</p>}
        </>
      )}
    </div>
  )
}

function CacheStep({ data, onField, onFieldNumber, onType, errors, testing, testResult, onTest }: {
  data: FormData
  onField: (key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => void
  onFieldNumber: (key: keyof FormData) => (e: React.ChangeEvent<HTMLInputElement>) => void
  onType: (v: string) => void
  errors: Record<string, string>
  testing: boolean
  testResult: 'idle' | 'ok' | 'fail'
  onTest: () => void
}) {
  const t = useT()
  return (
    <div className="space-y-4">
      <TypeSelect label={t('install.cacheType')} value={data.cacheType} onChange={onType} options={[
        { value: 'local', label: t('install.localMemory') },
        { value: 'redis', label: 'Redis' },
      ]} />
      {data.cacheType === 'redis' && (
        <>
          <div className="grid grid-cols-3 gap-3">
            <div className="col-span-2">
              <Field label={t('common.host')} value={data.redisAddr} onChange={onField('redisAddr')} placeholder="127.0.0.1" error={errors.redisAddr} />
            </div>
            <Field label={t('common.port')} value={String(data.redisPort)} onChange={onFieldNumber('redisPort')} placeholder="6379" error={errors.redisPort} />
          </div>
          <Field label={t('common.password')} value={data.redisPass} onChange={onField('redisPass')} placeholder={t('install.redisPassPlaceholder')} type="password" error={errors.redisPass} />
          <Field label={t('install.dbIndex')} value={String(data.redisDB)} onChange={onFieldNumber('redisDB')} placeholder={t('install.dbIndexPlaceholder')} />
          <TestButton label={t('common.testConnection')} testing={testing} result={testResult} onTest={onTest} />
          {errors.redisTest && <p className="text-xs text-text-3">{errors.redisTest}</p>}
        </>
      )}
    </div>
  )
}

/* ============================================
   Shared primitives
   ============================================ */

function Field({ label, value, onChange, placeholder, error, type, filter }: {
  label: string
  value: string
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void
  placeholder?: string
  error?: string
  type?: string
  filter?: 'alphanumeric'
}) {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (filter === 'alphanumeric') {
      e.target.value = e.target.value.replace(/[^a-zA-Z0-9]/g, '')
    }
    onChange(e)
  }

  return (
    <div className="space-y-1.5">
      <Label className="text-xs text-text-2">{label}</Label>
      <Input
        value={value}
        onChange={handleChange}
        placeholder={placeholder}
        type={type || 'text'}
        className={cn(
          'border-0 bg-bg-layer-2 text-sm text-text-1 placeholder:text-text-muted',
          error && 'ring-1 ring-text-3',
        )}
      />
      {error && <p className="text-xs text-text-3">{error}</p>}
    </div>
  )
}

function TypeSelect({ label, value, onChange, options }: {
  label: string
  value: string
  onChange: (v: string) => void
  options: { value: string; label: string }[]
}) {
  return (
    <div className="space-y-2">
      <Label className="text-xs text-text-2">{label}</Label>
      <div className="grid rounded-md bg-bg-layer-2 p-0.5" style={{ gridTemplateColumns: `repeat(${options.length}, 1fr)` }}>
        {options.map((opt) => (
          <button
            key={opt.value}
            type="button"
            onClick={() => onChange(opt.value)}
            className={cn(
              'rounded-sm py-2 text-sm text-center transition-colors',
              value === opt.value
                ? 'bg-bg-layer-3 text-text-1'
                : 'text-text-3 hover:text-text-2',
            )}
          >
            {opt.label}
          </button>
        ))}
      </div>
    </div>
  )
}

function TestButton({ label, testing, result, onTest }: {
  label: string
  testing: boolean
  result: 'idle' | 'ok' | 'fail'
  onTest: () => void
}) {
  const t = useT()
  return (
    <div className="flex items-center gap-3">
      <button
        type="button"
        onClick={onTest}
        disabled={testing}
        className={cn(
          'rounded-md bg-highlight px-4 py-1.5 text-xs text-highlight-fg transition-colors hover:opacity-90',
          testing && 'pointer-events-none opacity-50',
        )}
      >
        {testing ? t('common.testing') : label}
      </button>
      {result === 'ok' && <span className="text-xs text-text-2">{t('common.connectionSuccess')}</span>}
      {result === 'fail' && <span className="text-xs text-text-3">{t('common.connectionFailed')}</span>}
    </div>
  )
}
