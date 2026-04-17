import { Link, useNavigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { normalizeSearchQuery } from '@/shared/lib/search-query'
import { Search as SearchIcon } from 'lucide-react'
import { LandingQuickStartSection } from '@/shared/components/landing-quick-start'
import { useInView } from '@/shared/hooks/use-in-view'

/**
 * Marketing-style landing page for unauthenticated and first-time visitors.
 */
export function LandingPage() {
  const { t } = useTranslation()
  const navigate = useNavigate()

  const heroView = useInView()
  const quickStartView = useInView()

  const handleSearch = (query: string) => {
    const normalized = normalizeSearchQuery(query)
    navigate({
      to: '/search',
      search: { q: normalized, sort: 'relevance', page: 0 },
    })
  }

  return (
    <div className="landing-viewport flex flex-col justify-center">
      {/* Hero Section */}
      <main ref={heroView.ref} className={`relative z-10 flex flex-col items-center px-4 scroll-fade-up${heroView.inView ? ' in-view' : ''}`}>
        <h1
          className="text-3xl sm:text-4xl md:text-5xl font-bold tracking-tight text-center text-balance mb-4 max-w-3xl leading-[1.1]"
          style={{ color: 'hsl(var(--foreground))' }}
        >
          {t('landing.hero.title', { defaultValue: 'The skill dock for sharp agents' })}
        </h1>
        <p
          className="text-base md:text-lg text-center text-balance mb-8 max-w-xl px-2 leading-relaxed"
          style={{ color: 'hsl(var(--text-secondary))' }}
        >
          {t('landing.hero.subtitle', { defaultValue: 'Browse, install, and publish skill packs. Versioned, searchable, open.' })}
        </p>

        {/* Search box */}
        <div className="w-full max-w-[min(100%,54.6rem)] mb-6">
          <div
            className="flex items-center rounded-full border shadow-sm px-5 py-3.5 transition-colors focus-within:border-primary/60"
            style={{ background: 'hsl(var(--card))', borderColor: 'hsl(var(--border))' }}
          >
            <SearchIcon className="w-5 h-5 flex-shrink-0 mr-3" style={{ color: 'hsl(var(--text-placeholder))' }} strokeWidth={1.75} />
            <input
              type="text"
              placeholder={t('landing.hero.searchPlaceholder')}
              className="hero-input flex-1 bg-transparent outline-none text-base"
              style={{ color: 'hsl(var(--foreground))' }}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  handleSearch((e.target as HTMLInputElement).value)
                }
              }}
            />
          </div>
        </div>

        {/* CTA buttons */}
        <div className="flex flex-wrap justify-center gap-3">
          <Link
            to="/search"
            search={{ q: '', sort: 'relevance', page: 0 }}
            className="px-6 py-2.5 rounded-full text-sm font-medium shadow-sm transition-opacity hover:opacity-90"
            style={{ background: 'hsl(var(--primary))', color: 'hsl(var(--primary-foreground))' }}
          >
            {t('landing.hero.exploreSkills')}
          </Link>
          <Link
            to="/dashboard/publish"
            className="px-6 py-2.5 rounded-full text-sm font-medium border transition-colors hover:border-primary/40"
            style={{
              background: 'hsl(var(--card))',
              borderColor: 'hsl(var(--border))',
              color: 'hsl(var(--foreground))',
            }}
          >
            {t('landing.hero.publishSkill', { defaultValue: '开始构建' })}
          </Link>
        </div>
      </main>

      {/* Quick Start */}
      <div ref={quickStartView.ref} className={`scroll-fade-up${quickStartView.inView ? ' in-view' : ''}`}>
        <LandingQuickStartSection />
      </div>
    </div>
  )
}
