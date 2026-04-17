import type { NotificationItem } from '@/api/types'

export type NotificationDisplay = {
  title: string
  description: string
}

type NotificationBody = {
  skillName?: string
  version?: string
}

function parseBody(bodyJson?: string): NotificationBody {
  if (!bodyJson) {
    return {}
  }
  try {
    const parsed = JSON.parse(bodyJson)
    return typeof parsed === 'object' && parsed !== null ? parsed as NotificationBody : {}
  } catch {
    return {}
  }
}

export function resolveNotificationDisplay(item: NotificationItem, _language?: string): NotificationDisplay {
  void _language
  const body = parseBody(item.bodyJson)
  const skillName = body.skillName ?? ''
  const version = body.version ?? ''
  const versionSuffix = version ? ` (${version})` : ''

  switch (item.eventType) {
    case 'REVIEW_SUBMITTED':
      return {
        title: 'Review submitted',
        description: skillName ? `${skillName}${versionSuffix} was submitted for review.` : '',
      }
    case 'REVIEW_APPROVED':
      return {
        title: 'Review approved',
        description: skillName ? `${skillName}${versionSuffix} was approved.` : '',
      }
    case 'REVIEW_REJECTED':
      return {
        title: 'Review rejected',
        description: skillName ? `${skillName}${versionSuffix} was rejected.` : '',
      }
    case 'PROMOTION_SUBMITTED':
      return {
        title: 'Promotion submitted',
        description: skillName ? `${skillName}${versionSuffix} was submitted for promotion.` : '',
      }
    case 'PROMOTION_APPROVED':
      return {
        title: 'Promotion approved',
        description: skillName ? `${skillName}${versionSuffix} promotion was approved.` : '',
      }
    case 'PROMOTION_REJECTED':
      return {
        title: 'Promotion rejected',
        description: skillName ? `${skillName}${versionSuffix} promotion was rejected.` : '',
      }
    case 'REPORT_SUBMITTED':
      return {
        title: 'Report submitted',
        description: skillName ? `${skillName} received a new report.` : '',
      }
    case 'REPORT_RESOLVED':
      return {
        title: 'Report resolved',
        description: skillName ? `${skillName} report has been resolved.` : '',
      }
    case 'SKILL_PUBLISHED':
      return {
        title: 'Skill published',
        description: skillName ? `${skillName}${versionSuffix} was published.` : '',
      }
    default:
      return {
        title: item.title,
        description: '',
      }
  }
}
