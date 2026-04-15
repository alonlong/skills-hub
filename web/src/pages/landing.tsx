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
    <>
      {/* Hero Section */}
      <main ref={heroView.ref} className={`relative z-10 flex flex-col items-center pt-16 pb-6 px-4 md:pt-24 md:pb-8 scroll-fade-up${heroView.inView ? ' in-view' : ''}`}>
        <h1 className="text-5xl md:text-7xl font-bold tracking-tight text-brand-gradient mb-5">
          SkillHub
        </h1>
        <h2
          className="text-2xl md:text-3xl font-semibold tracking-tight text-center text-balance mb-10 max-w-3xl px-2"
          style={{ color: 'hsl(var(--foreground))' }}
        >
          {t('landing.hero.title')}
        </h2>

        {/* Search box */}
        <div className="w-full max-w-[min(100%,54.6rem)] mb-8">
          <div
            className="flex items-center bg-white rounded-xl border shadow-sm px-5 py-3.5"
            style={{ borderColor: 'hsl(var(--border))' }}
          >
            <SearchIcon className="w-5 h-5 flex-shrink-0 mr-3" style={{ color: 'hsl(var(--text-placeholder))' }} strokeWidth={1.5} />
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
        <div className="flex flex-wrap justify-center gap-4">
          <Link
            to="/search"
            search={{ q: '', sort: 'relevance', page: 0 }}
            className="px-8 py-3.5 rounded-xl text-base font-medium text-white bg-brand-gradient shadow-sm hover:opacity-95 transition-opacity"
          >
            {t('landing.hero.exploreSkills')}
          </Link>
          <Link
            to="/dashboard/publish"
            className="px-8 py-3.5 rounded-xl text-base font-medium border transition-colors"
            style={{
              background: 'hsl(var(--secondary))',
              borderColor: 'hsl(var(--muted-foreground))',
              color: 'hsl(var(--muted-foreground))',
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
    </>
  )
}
