import { queryOptions, useQuery } from '@tanstack/react-query'
import { authApi } from '@/api/client'

/**
 * Auth hook used throughout the app.
 *
 * The Go backend is token-based, so we rely on explicit login/logout updates plus
 * focus/reconnect refreshes instead of background session polling.
 */
export function getAuthQueryOptions(enabled = true) {
  return queryOptions({
    queryKey: ['auth', 'me'] as const,
    queryFn: authApi.getMe,
    retry: false,
    enabled,
    staleTime: 0,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    refetchInterval: false as const,
  })
}

export function useAuth(enabled = true) {
  const { data: user, isLoading, error } = useQuery(getAuthQueryOptions(enabled))

  return {
    user: user ?? null,
    isLoading,
    isAuthenticated: !!user,
    hasRole: (role: string) => user?.platformRoles?.includes(role) ?? false,
    error,
  }
}
