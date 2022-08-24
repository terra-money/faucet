import i18next from 'i18next'
import LanguageDetector from 'i18next-browser-languagedetector'
import HttpApi from 'i18next-http-backend'
import { initReactI18next } from 'react-i18next'

i18next
    .use(HttpApi)
    .use(LanguageDetector)
    .use(initReactI18next)
    .init({
        backend: {
            crossDomain: true,
            loadPath() {
                return 'https://raw.githubusercontent.com/mars-protocol/translations/develop/{{lng}}.json'
            },
        },
        react: {
            useSuspense: true,
        },
        fallbackLng: ['en'],
        preload: ['en'],
        keySeparator: '.',
        interpolation: { escapeValue: false },
        lowerCaseLng: true,
        load: 'languageOnly',
    })

export default i18next
