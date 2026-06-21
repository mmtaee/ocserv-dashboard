import en from "./locales/en.json"

export type Locale = "en" | "fa" | "es" | "fr"

export type TranslationSchema = typeof en

export type LocaleConfig = {
  name: string
  direction: "ltr" | "rtl"
}
