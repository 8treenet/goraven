import { useState, useEffect, useCallback, useRef } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { AlertCircle, Eye, EyeOff } from 'lucide-react'
import axios from 'axios'
import { useT, t as translate } from '@/i18n'
import { authApi } from '@/api'
import { useUserStore } from '@/stores/user-store'
import { Logo } from '@/components/brand/Logo'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
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
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-bg-base px-4 py-12">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -left-24 -top-24 size-80 rounded-full bg-accent-soft/60 blur-3xl"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute -bottom-32 -right-24 size-96 rounded-full bg-amber-soft/70 blur-3xl"
      />
      <div className="absolute left-6 top-6 flex items-center gap-2">
        <Logo size={28} />
        <span className="text-sm font-semibold text-text-2">GoRaven</span>
      </div>
      <div className="w-full max-w-[440px]">
        <div className="mb-8 flex justify-center">
          <h1 className="flex items-center gap-2 text-title font-semibold tracking-tight text-text-1">
            GoRaven
            <span aria-hidden="true" className="size-1.5 rounded-full bg-amber" />
          </h1>
        </div>

        <Card className="border-t-2 border-t-amber">
          <CardContent className="space-y-5 p-6">
            {isExpired && (
              <p className="flex items-center justify-center gap-1.5 text-center text-xs text-warning">
                <AlertCircle className="size-3.5 shrink-0" />
                {t('login.sessionExpired')}
              </p>
            )}

            <div className="space-y-1.5">
              <Label className="text-text-2">{t('common.username')}</Label>
              <Input
                ref={usernameRef}
                value={form.username}
                onChange={(e) => setField('username', e.target.value)}
                onBlur={onUsernameBlur}
                onKeyDown={handleKeyDown}
                placeholder={t('login.usernamePlaceholder')}
                autoComplete="username"
                disabled={submitting}
                aria-invalid={!!fieldErrors.username}
              />
              {fieldErrors.username && (
                <p className="text-xs text-destructive">{fieldErrors.username}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label className="text-text-2">{t('common.password')}</Label>
              <div className="relative">
                <Input
                  ref={passwordRef}
                  value={form.password}
                  onChange={(e) => setField('password', e.target.value)}
                  onKeyDown={handleKeyDown}
                  type={showPassword ? 'text' : 'password'}
                  placeholder={t('login.passwordPlaceholder')}
                  autoComplete="current-password"
                  disabled={submitting}
                  aria-invalid={!!fieldErrors.password}
                  className="pr-10"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="icon-sm"
                  onClick={() => setShowPassword((v) => !v)}
                  tabIndex={-1}
                  aria-label={t('common.password')}
                  className="absolute right-1 top-1/2 -translate-y-1/2 text-text-muted hover:text-accent"
                >
                  {showPassword ? (
                    <EyeOff className="size-4" />
                  ) : (
                    <Eye className="size-4" />
                  )}
                </Button>
              </div>
              {fieldErrors.password && (
                <p className="text-xs text-destructive">{fieldErrors.password}</p>
              )}
            </div>

            {captchaData?.required && (
              <div className="space-y-1.5">
                <Label className="text-text-2">{t('login.captchaLabel')}</Label>
                <div className="flex items-center gap-2">
                  <img src={captchaData.image1} alt="" className="h-9 w-auto rounded-lg border border-accent-border bg-accent-soft" />
                  <span className="text-sm text-text-2">+</span>
                  <img src={captchaData.image2} alt="" className="h-9 w-auto rounded-lg border border-accent-border bg-accent-soft" />
                  <span className="text-sm text-text-2">=</span>
                  <Input
                    value={captchaAnswer}
                    onChange={(e) => {
                      setCaptchaAnswer(e.target.value.replace(/\D/g, ''))
                      setCaptchaError('')
                    }}
                    onKeyDown={handleKeyDown}
                    placeholder={t('login.captchaPlaceholder')}
                    disabled={submitting}
                    autoComplete="off"
                    aria-invalid={!!captchaError}
                    className="w-16"
                  />
                </div>
                {captchaError ? (
                  <p className="text-xs text-destructive">{captchaError}</p>
                ) : (
                  <p className="text-xs text-text-3">{t('login.captchaHint')}</p>
                )}
              </div>
            )}

            {serverError && (
              <p aria-live="polite" className="text-center text-xs text-destructive">
                {serverError}
              </p>
            )}

            <Button type="button" onClick={onSubmit} disabled={submitting} className="w-full">
              {submitting ? t('login.loggingIn') : t('login.login')}
            </Button>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
