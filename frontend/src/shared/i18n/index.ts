import { computed, ref } from 'vue'

import type { AuthProvider } from '@/entities/user/model/types'
import { readStorage, writeStorage } from '@/shared/lib/storage'

import { messages, type AppLocale, type MessageKey } from './messages'

const LOCALE_STORAGE_KEY = 'plos-locale'
const currentLocale = ref<AppLocale>('ru')

function detectLocale(): AppLocale {
  if (typeof navigator === 'undefined') {
    return 'ru'
  }

  return navigator.language.toLowerCase().startsWith('en') ? 'en' : 'ru'
}

function interpolate(template: string, params?: Record<string, string | number>) {
  if (!params) {
    return template
  }

  return template.replace(/\{(\w+)\}/g, (_, key: string) => String(params[key] ?? `{${key}}`))
}

function syncDocumentLanguage(locale: AppLocale) {
  if (typeof document === 'undefined') {
    return
  }

  document.documentElement.lang = locale
}

export function bootstrapI18n() {
  currentLocale.value = readStorage<AppLocale>(LOCALE_STORAGE_KEY, detectLocale())
  syncDocumentLanguage(currentLocale.value)
}

export function getLocale() {
  return currentLocale.value
}

export function getLocaleTag(locale: AppLocale = currentLocale.value) {
  return locale === 'en' ? 'en-US' : 'ru-RU'
}

export function setLocale(locale: AppLocale) {
  currentLocale.value = locale
  writeStorage(LOCALE_STORAGE_KEY, locale)
  syncDocumentLanguage(locale)
}

export function translate(key: MessageKey, params?: Record<string, string | number>) {
  const template = messages[currentLocale.value][key] ?? messages.en[key] ?? key
  return interpolate(template, params)
}

export function translateAccountType(type: string) {
  switch (type) {
    case 'cash':
      return translate('accountType.cash')
    case 'savings':
      return translate('accountType.savings')
    default:
      return translate('accountType.bank')
  }
}

export function translateCategoryKind(kind: string) {
  return kind === 'income' ? translate('categoryKind.income') : translate('categoryKind.expense')
}

export function translateCategoryScope(scope: string) {
  return scope === 'system' ? translate('categoryScope.system') : translate('categoryScope.custom')
}

export function translateProvider(provider: AuthProvider | null) {
  if (provider === 'telegram') {
    return translate('provider.telegram')
  }

  if (provider === 'yandex') {
    return translate('provider.yandex')
  }

  return translate('common.noProvider')
}

export function useI18n() {
  return {
    locale: computed(() => currentLocale.value),
    setLocale,
    t: translate,
  }
}

export { currentLocale }
