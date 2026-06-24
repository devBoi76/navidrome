import polyglotI18nProvider from 'ra-i18n-polyglot'
import deepmerge from 'deepmerge'
import dataProvider from '../dataProvider'
import pl from './pl.json'
import { i18nProvider } from './index'
import config from '../config'

// Only returns current selected locale if its translations are found in localStorage
const defaultLocale = function () {
  const locale = localStorage.getItem('locale') || config.defaultLanguage
  const current = JSON.parse(localStorage.getItem('translation'))
  if (current && current.id === locale) {
    // Asynchronously reload the translation from the server
    retrieveTranslation(locale).then(() => {
      i18nProvider.changeLocale(locale)
    })
    return locale
  }
  return config.defaultLanguage || 'pl'
}

export function retrieveTranslation(locale) {
  return dataProvider.getOne('translation', { id: locale }).then((res) => {
    localStorage.setItem('translation', JSON.stringify(res.data))
    return prepareLanguage(JSON.parse(res.data.data))
  })
}

const removeEmpty = (obj) => {
  for (let k in obj) {
    if (
      Object.prototype.hasOwnProperty.call(obj, k) &&
      typeof obj[k] === 'object'
    ) {
      removeEmpty(obj[k])
    } else {
      if (!obj[k]) {
        delete obj[k]
      }
    }
  }
}

const prepareLanguage = (lang) => {
  removeEmpty(lang)
  // Make "albumSong" and "playlistTrack" resource use the same translations as "song"
  lang.resources.albumSong = lang.resources.song
  lang.resources.playlistTrack = lang.resources.song
  // ra.boolean.null should always be empty
  lang.ra.boolean.null = ''
  // Fallback to english translations
  return deepmerge(pl, lang)
}

export default polyglotI18nProvider((locale) => {
  // English is bundled
  if (locale === 'pl') {
    return prepareLanguage(pl)
  }
  // If the requested locale is in already loaded, return it
  const current = JSON.parse(localStorage.getItem('translation'))
  if (current && current.id === locale) {
    return prepareLanguage(JSON.parse(current.data))
  }
  // If not, get it from the server, and store it in localStorage
  return retrieveTranslation(locale)
}, defaultLocale())
