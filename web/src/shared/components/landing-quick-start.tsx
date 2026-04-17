import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { Bot, Check, Copy, UserRound } from 'lucide-react'
import { useCopyToClipboard } from '@/shared/lib/clipboard'

type LandingQuickStartTabId = 'agent' | 'human'

interface LandingQuickStartTab {
  id: LandingQuickStartTabId
  label: string
  description: string
  command: string
}

/**
 * Get the base URL for the application.
 * Prefers the runtime config if set and not localhost.
 * Falls back to the current page origin.
 */
function getAppBaseUrl(): string {
  if (typeof window === 'undefined') {
    return ''
  }
  const runtimeConfig = window.__SKILLHUB_RUNTIME_CONFIG__
  const configuredUrl = runtimeConfig?.appBaseUrl
  // Use configured URL only if it's set and not localhost
  if (configuredUrl && !configuredUrl.includes('localhost')) {
    return configuredUrl
  }
  // Fallback to current page origin
  return `${window.location.protocol}//${window.location.host}`
}

function CompactCopyButton({ text }: { text: string }) {
  const { t } = useTranslation()
  const [copied, copy] = useCopyToClipboard()

  const handleCopy = async () => {
    try {
      await copy(text)
    } catch (err) {
      console.error('Failed to copy:', err)
    }
  }

  const label = copied ? (t('copyButton.copied') || 'Copied') : (t('copyButton.copy') || 'Copy')

  return (
    <button
      type="button"
      onClick={handleCopy}
      aria-label={label}
      title={label}
      className="absolute right-2.5 top-1/2 flex h-9 w-9 -translate-y-1/2 items-center justify-center rounded-full border transition-colors hover:opacity-80 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 cursor-pointer"
      style={{ background: 'hsl(var(--card))', borderColor: 'hsl(var(--border))', color: 'hsl(var(--foreground))' }}
    >
      {copied ? <Check className="h-4 w-4" /> : <Copy className="h-4 w-4" />}
    </button>
  )
}

export function LandingQuickStartSection() {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<LandingQuickStartTabId>('agent')
  const baseUrl = useMemo(() => getAppBaseUrl(), [])

  // Build dynamic agent command with actual registry URL
  const agentCommand = t('landing.quickStart.agent.commandTemplate', {
    defaultValue: t('landing.quickStart.agent.command'),
    url: `${baseUrl}/registry/skill.md`,
  })

  const tabs: LandingQuickStartTab[] = [
    {
      id: 'agent',
      label: t('landing.quickStart.tabs.agent'),
      description: t('landing.quickStart.agent.description'),
      command: agentCommand,
    },
    {
      id: 'human',
      label: t('landing.quickStart.tabs.human'),
      description: t('landing.quickStart.human.description'),
      command: t('landing.quickStart.human.command'),
    },
  ]

  const currentTab = tabs.find((tab) => tab.id === activeTab) ?? tabs[0]

  return (
    <section className="relative z-10 w-full px-6 pt-4 pb-4 md:pt-6 md:pb-6" style={{ background: 'var(--bg-page, hsl(var(--background)))' }}>
      <div className="mx-auto max-w-[min(100%,72.8rem)]">
        <div className="text-center mb-3 md:mb-4">
          <h2 className="text-2xl md:text-3xl font-bold tracking-tight mb-1.5" style={{ color: 'hsl(var(--foreground))' }}>
            {t('landing.quickStart.title')}
          </h2>
          <p className="text-sm md:text-base max-w-2xl mx-auto leading-snug" style={{ color: 'hsl(var(--text-secondary))' }}>
            {t('landing.quickStart.description', { defaultValue: t('landing.quickStart.subtitle') })}
          </p>
        </div>

        <div
          className="mx-auto max-w-[min(100%,54.6rem)] rounded-[24px] border p-2.5 shadow-[0_18px_40px_-24px_rgba(120,60,40,0.20)]"
          style={{ background: 'hsl(var(--card))', borderColor: 'hsl(var(--border-card))' }}
        >
          <div
            className="grid grid-cols-2 gap-2 rounded-2xl p-1.5"
            style={{ background: 'hsl(var(--muted) / 0.6)' }}
          >
            {tabs.map((tab) => {
              const isActive = tab.id === currentTab.id
              const Icon = tab.id === 'agent' ? Bot : UserRound

              return (
                <button
                  key={tab.id}
                  type="button"
                  onClick={() => setActiveTab(tab.id)}
                  aria-pressed={isActive}
                  className="flex min-h-9 items-center justify-center gap-2 rounded-[14px] px-3 py-2 text-sm font-medium transition-all duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 cursor-pointer"
                  style={{
                    background: isActive ? 'hsl(var(--card))' : 'transparent',
                    color: isActive ? 'hsl(var(--foreground))' : 'hsl(var(--muted-foreground))',
                    boxShadow: isActive ? '0 4px 14px rgba(120,60,40,0.10)' : 'none',
                  }}
                >
                  <Icon className="h-4 w-4" strokeWidth={1.75} />
                  <span>{tab.label}</span>
                </button>
              )
            })}
          </div>

          <div className="px-3 pb-3 pt-4 md:px-6 md:pb-4 md:pt-5">
            <p
              className="mx-auto mb-3 max-w-xl text-center text-sm font-medium leading-snug md:text-base"
              style={{ color: 'hsl(var(--foreground))' }}
            >
              {currentTab.description}
            </p>

            <div
              className="relative rounded-2xl border px-4 py-2.5 pr-14"
              style={{ background: 'hsl(var(--muted) / 0.5)', borderColor: 'hsl(var(--border))' }}
            >
              <div className="min-w-0">
                <code
                  className="block font-mono text-xs md:text-sm break-all whitespace-pre-wrap leading-relaxed"
                  style={{ color: currentTab.id === 'agent' ? 'hsl(var(--primary))' : 'hsl(var(--foreground))' }}
                >
                  {currentTab.command}
                </code>
              </div>
              <CompactCopyButton text={currentTab.command} />
            </div>
          </div>
        </div>
      </div>
    </section>
  )
}
