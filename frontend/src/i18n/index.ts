import { useCallback } from 'react'
import { useAppStore } from '@/stores/app-store'
import zh from './locales/zh'
import en from './locales/en'

export type TranslationKey = keyof typeof zh & keyof typeof en

const translations = { zh, en } as const

export function t(key: TranslationKey): string {
  const language = useAppStore.getState().language
  return translations[language][key]
}

export function useT() {
  const language = useAppStore((s) => s.language)
  const dict = translations[language]

  return useCallback(
    (key: TranslationKey) => dict[key],
    [dict],
  )
}
