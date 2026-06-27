import { createI18n } from 'vue-i18n';

type MessageSchema = Record<string, any>;
type Language = { code: string; label: string };

const localeFiles = import.meta.glob<MessageSchema>('/src/locales/*.json', {
    eager: true
});

export const languages: Language[] = (import.meta.env.VITE_I18N_LANGUAGES || 'en:English')
    .split(',')
    .map((lang: string) => {
        const [code, label] = lang.split(':');
        return { code: code.trim(), label: label?.trim() || code.trim() };
    });

const messages: Record<string, MessageSchema> = {};

languages.forEach((lang) => {
    const filePath = `/src/locales/${lang.code}.json`; // use lang.code
    if (localeFiles[filePath]) {
        messages[lang.code] = localeFiles[filePath] as MessageSchema;
    } else {
        console.warn(`⚠️ Missing locale file for: ${lang.code} (expected ${filePath})`);
    }
});

const userLang = localStorage.getItem('language') || 'en';

const RTL_LOCALES = new Set(['fa', 'ar']);

export const isRTL = (code: string) => RTL_LOCALES.has(code);

export const applyLocaleDirection = (code: string) => {
    if (typeof document === 'undefined') return;
    const dir = isRTL(code) ? 'rtl' : 'ltr';
    document.documentElement.setAttribute('dir', dir);
    document.documentElement.setAttribute('lang', code);
};

applyLocaleDirection(userLang);

export default createI18n<[MessageSchema], string>({
    legacy: false,
    locale: userLang,
    fallbackLocale: userLang,
    messages
});
