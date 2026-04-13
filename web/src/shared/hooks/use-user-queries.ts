import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import type { SkillSummary, PagedResponse } from '@/api/types'
import { meApi, promotionApi, namespaceApi } from '@/api/client'

async function getMySkills(params: { page?: number; size?: number; filter?: string } = {}): Promise<PagedResponse<SkillSummary>> {
  return meApi.getSkills(params)
}

async function submitPromotion(params: { sourceSkillId: number; sourceVersionId: number }): Promise<void> {
  const globalNamespace = await namespaceApi.getDetail('global')
  await promotionApi.submit({
    sourceSkillId: params.sourceSkillId,
    sourceVersionId: params.sourceVersionId,
    targetNamespaceId: globalNamespace.id,
  })
}

export function useMySkills(params: { page?: number; size?: number; filter?: string } = {}) {
  return useQuery({
    queryKey: ['skills', 'my', params],
    queryFn: () => getMySkills(params),
  })
}

export function useSubmitPromotion() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: submitPromotion,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['promotions'] })
      queryClient.invalidateQueries({ queryKey: ['governance'] })
      queryClient.invalidateQueries({ queryKey: ['skills'] })
    },
  })
}
