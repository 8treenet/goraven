import { useState, useEffect, useCallback, useRef } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { Eye, EyeOff } from 'lucide-react'
import axios from 'axios'
import { cn } from '@/lib/utils'
import { useT, t as translate } from '@/i18n'
import { authApi } from '@/api'
import { useUserStore } from '@/stores/user-store'
import type { CaptchaRsp } from '@/api/types'

interface LoginForm {
  username: string
  password: string
}

export function Component() {
  const navigate = useNavigate()
  const t = useT()
  const [searchParams] = useSearchParams()
  const isExpired = searchParams.get('expired') === '1'

  const [form, setForm] = useState<LoginForm>({ username: '', password: '' })
  const [showPassword, setShowPassword] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [serverError, setServerError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Partial<Record<keyof LoginForm, string>>>({})

  const [captchaData, setCaptchaData] = useState<CaptchaRsp | null>(null)
  const [captchaAnswer, setCaptchaAnswer] = useState('')
  const [captchaError, setCaptchaError] = useState('')

  const usernameRef = useRef<HTMLInputElement>(null)
  const passwordRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    usernameRef.current?.focus()
  }, [])

  const setField = useCallback((field: keyof LoginForm, value: string) => {
    setForm((f) => ({ ...f, [field]: value }))
    setFieldErrors((e) => ({ ...e, [field]: undefined }))
  }, [])

  const fetchCaptcha = useCallback(async (username: string) => {
    if (!username || username.length < 8) {
      setCaptchaData(null)
      return
    }
    try {
      const rsp = await authApi.getCaptcha(username)
      if (rsp.required) {
        setCaptchaData(rsp)
        setCaptchaAnswer('')
        setCaptchaError('')
        setServerError('')
      } else {
        setCaptchaData(null)
      }
    } catch {
      setCaptchaData(null)
    }
  }, [])

  const onUsernameBlur = useCallback(() => {
    fetchCaptcha(form.username)
  }, [form.username, fetchCaptcha])

  const onSubmit = useCallback(async () => {
    const errs: Partial<Record<keyof LoginForm, string>> = {}
    if (!form.username) errs.username = translate('login.errUsernameRequired')
    if (!form.password) errs.password = translate('login.errPasswordRequired')
    if (Object.keys(errs).length > 0) {
      setFieldErrors(errs)
      return
    }

    setSubmitting(true)
    setServerError('')
    setCaptchaError('')
    try {
      const captchaVal = captchaData?.required ? Number(captchaAnswer) : undefined
      const loginResult = await authApi.login({
        username: form.username,
        password: form.password,
        captchaAnswer: captchaVal,
      })
      useUserStore.setState({ token: loginResult.accessToken })
      const user = await authApi.getCurrentUser()
      useUserStore.getState().setAuth(loginResult.accessToken, {
        userId: user.userId,
        username: user.username,
        nickname: user.nickname,
        avatar: user.avatar,
        role: user.role,
        email: user.email,
      })
      navigate('/', { replace: true })
    } catch (err) {
      if (axios.isAxiosError(err)) {
        setServerError(translate('login.errNetworkFailed'))
      } else {
        const msg = err instanceof Error ? err.message : translate('login.errNetworkFailed')
        if (msg.includes('captcha') || msg.includes('验证码')) {
          setCaptchaError(translate('login.errCaptchaIncorrect'))
        } else {
          setServerError(msg)
        }
        setForm((f) => ({ ...f, password: '' }))
        setCaptchaAnswer('')
        fetchCaptcha(form.username)
        setTimeout(() => passwordRef.current?.focus(), 0)
      }
    } finally {
      setSubmitting(false)
    }
  }, [form, navigate, captchaData, captchaAnswer, fetchCaptcha])

  const handleKeyDown = useCallback((e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !submitting) {
      e.preventDefault()
      onSubmit()
    }
  }, [onSubmit, submitting])

  return (
    <div className="relative flex min-h-screen items-center justify-center bg-bg-base px-4">
      <div className="absolute left-6 top-6 flex items-center gap-2">
        <img src="/favicon.svg" alt="GoRaven" className="h-7 w-7" />
        <span className="text-sm font-semibold text-text-2">GoRaven</span>
      </div>
      <div className="w-full max-w-md">

        <div className="mb-10 flex justify-center">
          <h1 className="text-2xl font-semibold text-text-1 tracking-tight">GoRaven</h1>
        </div>

        {isExpired && (
          <div className="mb-6 text-center">
            <p className="text-xs text-text-3">{t('login.sessionExpired')}</p>
          </div>
        )}

        <div className="space-y-5">
          <div className="space-y-1.5">
            <label className="text-xs text-text-2">{t('common.username')}</label>
            <input
              ref={usernameRef}
              value={form.username}
              onChange={(e) => setField('username', e.target.value)}
              onBlur={onUsernameBlur}
              onKeyDown={handleKeyDown}
              placeholder={t('login.usernamePlaceholder')}
              autoComplete="username"
              disabled={submitting}
              className={cn(
                'h-8 w-full min-w-0 rounded-lg border border-border-custom bg-transparent px-2.5 py-1 text-sm text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-2 focus:ring-ring/30 md:text-sm',
                fieldErrors.username && 'border-text-3 ring-1 ring-text-3',
              )}
            />
            {fieldErrors.username && (
              <p className="text-xs text-text-3">{fieldErrors.username}</p>
            )}
          </div>

          <div className="space-y-1.5">
            <label className="text-xs text-text-2">{t('common.password')}</label>
            <div className="relative">
              <input
                ref={passwordRef}
                value={form.password}
                onChange={(e) => setField('password', e.target.value)}
                onKeyDown={handleKeyDown}
                type={showPassword ? 'text' : 'password'}
                placeholder={t('login.passwordPlaceholder')}
                autoComplete="current-password"
                disabled={submitting}
                className={cn(
                  'h-8 w-full min-w-0 rounded-lg border border-border-custom bg-transparent pr-10 pl-2.5 py-1 text-sm text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-2 focus:ring-ring/30 md:text-sm',
                  fieldErrors.password && 'border-text-3 ring-1 ring-text-3',
                )}
              />
              <button
                type="button"
                onClick={() => setShowPassword((v) => !v)}
                tabIndex={-1}
                className="absolute right-2 top-1/2 -translate-y-1/2 text-text-muted transition-colors hover:text-text-3"
              >
                {showPassword ? (
                  <EyeOff className="size-4" />
                ) : (
                  <Eye className="size-4" />
                )}
              </button>
            </div>
            {fieldErrors.password && (
              <p className="text-xs text-text-3">{fieldErrors.password}</p>
            )}
          </div>

          {captchaData?.required && (
            <div className="space-y-1.5">
              <label className="text-xs text-text-2">{t('login.captchaLabel')}</label>
              <div className="flex items-center gap-2">
                <img src={captchaData.image1} alt="" className="h-8 w-auto rounded" />
                <span className="text-sm text-text-2">+</span>
                <img src={captchaData.image2} alt="" className="h-8 w-auto rounded" />
                <span className="text-sm text-text-2">=</span>
                <input
                  value={captchaAnswer}
                  onChange={(e) => {
                    setCaptchaAnswer(e.target.value.replace(/\D/g, ''))
                    setCaptchaError('')
                  }}
                  onKeyDown={handleKeyDown}
                  placeholder={t('login.captchaPlaceholder')}
                  disabled={submitting}
                  autoComplete="off"
                  className={cn(
                    'h-8 w-16 min-w-0 rounded-lg border border-border-custom bg-transparent px-2 py-1 text-sm text-text-1 placeholder:text-text-muted outline-none transition-colors focus:border-ring focus:ring-2 focus:ring-ring/30',
                    captchaError && 'border-text-3 ring-1 ring-text-3',
                  )}
                />
              </div>
              {captchaError ? (
                <p className="text-xs text-text-3">{captchaError}</p>
              ) : (
                <p className="text-xs text-text-muted">{t('login.captchaHint')}</p>
              )}
            </div>
          )}

          {serverError && (
            <p className="text-center text-xs text-text-3">{serverError}</p>
          )}

          <button
            type="button"
            onClick={onSubmit}
            disabled={submitting}
            className={cn(
              'w-full rounded-md bg-highlight py-2 text-sm text-highlight-fg transition-colors hover:opacity-90',
              submitting && 'pointer-events-none opacity-50',
            )}
          >
            {submitting ? t('login.loggingIn') : t('login.login')}
          </button>
        </div>

      </div>
    </div>
  )
}
