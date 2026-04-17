import { describe, expect, it, vi } from 'vitest'

const initMock = vi.fn().mockReturnThis()
const useMock = vi.fn().mockReturnThis()

vi.mock('i18next', () => ({
  default: {
    use: useMock,
    init: initMock,
  },
}))

vi.mock('react-i18next', () => ({
  initReactI18next: { type: '3rdParty', init: vi.fn() },
}))

vi.mock('./locales/en.json', () => ({
  default: { greeting: 'Hello' },
}))

await import('./config')

describe('i18n config', () => {
  it('chains the react-i18next plugin', () => {
    expect(useMock).toHaveBeenCalled()
  })

  it('initializes with English as the only language', () => {
    expect(initMock).toHaveBeenCalledTimes(1)
    const initOptions = initMock.mock.calls[0][0]
    expect(initOptions.lng).toBe('en')
    expect(initOptions.fallbackLng).toBe('en')
  })

  it('disables HTML escaping for React interpolation', () => {
    const initOptions = initMock.mock.calls[0][0]
    expect(initOptions.interpolation.escapeValue).toBe(false)
  })

  it('registers only the english resource bundle', () => {
    const initOptions = initMock.mock.calls[0][0]
    expect(initOptions.resources).toHaveProperty('en')
    expect(initOptions.resources).not.toHaveProperty('zh')
    expect(initOptions.resources.en).toHaveProperty('translation')
  })
})
