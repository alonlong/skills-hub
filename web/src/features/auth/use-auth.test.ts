import { describe, expect, it } from 'vitest'
import { getAuthQueryOptions } from './use-auth'

describe('getAuthQueryOptions', () => {
  it('refreshes auth on focus and reconnect without background polling', () => {
    const options = getAuthQueryOptions(true)

    expect(options.queryKey).toEqual(['auth', 'me'])
    expect(options.staleTime).toBe(0)
    expect(options.refetchOnWindowFocus).toBe(true)
    expect(options.refetchOnReconnect).toBe(true)
    expect(options.refetchInterval).toBe(false)
    expect(options.enabled).toBe(true)
  })
})
