import { useMutation, useQueryClient } from '@tanstack/react-query'
import { authApi } from '@/api/client'
import { ApiError } from '@/shared/lib/api-error'
import type { LocalLoginRequest, User } from '@/api/types'
import { clearSessionScopedQueries } from '@/features/notification/notification-session'

/**
 * Password-login mutation that can switch between local auth and direct upstream auth based on
 * runtime configuration.
 */
export function usePasswordLogin() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: LocalLoginRequest) => authApi.localLogin(request),
    onSuccess: ({ accessToken, user }) => {
      window.localStorage.setItem('skillhub.accessToken', accessToken)
      clearSessionScopedQueries(queryClient)
      queryClient.setQueryData<User | null>(['auth', 'me'], user)
    },
    onError: (error) => {
      // Keep invalid credentials on the login page instead of falling back to the
      // global 401 redirect handler used for background API requests.
      if (error instanceof ApiError) {
        return
      }
    },
  })
}
