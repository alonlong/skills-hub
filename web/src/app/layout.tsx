import { Suspense, useEffect, useState } from 'react'
import { Outlet, Link, useRouterState } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useAuth } from '@/features/auth/use-auth'
import { LanguageSwitcher } from '@/shared/components/language-switcher'
import { UserMenu } from '@/shared/components/user-menu'
import { NotificationBell } from '@/features/notification/notification-bell'
import { getAppHeaderClassName } from './layout-header-style'
import { getAppMainContentLayout, resolveAppMainContentPathname } from './layout-main-content'

/**
 * Application shell shared by all routed pages.
 *
 * It owns the global header, language switcher, auth-aware navigation, and suspense
 * fallback used while lazy route modules are loading.
 */
export function Layout() {
  const { t } = useTranslation()
  const { pathname, resolvedPathname } = useRouterState({
    select: (s) => ({
      pathname: s.location.pathname,
      resolvedPathname: s.resolvedLocation?.pathname,
    }),
  })
  const { user, isLoading } = useAuth()
  const [isHeaderElevated, setIsHeaderElevated] = useState(false)
  const contentLayoutPathname = resolveAppMainContentPathname(pathname, resolvedPathname)
  const mainContentLayout = getAppMainContentLayout(contentLayoutPathname)

  useEffect(() => {
    const updateHeaderElevation = () => {
      setIsHeaderElevated(window.scrollY > 0)
    }

    updateHeaderElevation()
    window.addEventListener('scroll', updateHeaderElevation, { passive: true })

    return () => {
      window.removeEventListener('scroll', updateHeaderElevation)
    }
  }, [])

  const navItems: Array<{
    label: string
    to: string
    exact?: boolean
    auth?: boolean
  }> = [
    { label: t('nav.search'), to: '/search' },
    { label: t('nav.publish'), to: '/dashboard/publish', auth: true },
    { label: t('nav.mySkills'), to: '/dashboard/skills', auth: true },
  ]

  const isActive = (to: string, exact?: boolean) => {
    if (exact) return pathname === to
    return pathname === to
  }

  return (
    <div className="min-h-screen flex flex-col relative overflow-x-clip" style={{ background: 'hsl(var(--background))' }}>
      {/* Decorative warm gradient */}
      <div
        className="absolute top-0 right-0 w-[720px] h-[540px] pointer-events-none z-0"
        style={{
          background: 'radial-gradient(ellipse at 80% 10%, rgba(205,94,74,0.18) 0%, rgba(216,134,94,0.10) 35%, transparent 70%)',
          filter: 'blur(40px)',
        }}
      />

      {/* Header */}
      <header className={getAppHeaderClassName(isHeaderElevated)} style={{ borderColor: 'hsl(var(--border))' }}>
        <Link to="/" className="flex items-center gap-2.5 text-xl font-bold tracking-tight" style={{ color: 'hsl(var(--foreground))' }}>
          <span
            className="inline-flex h-9 w-9 items-center justify-center rounded-full text-lg"
            style={{ background: 'hsl(var(--primary) / 0.12)', color: 'hsl(var(--primary))' }}
            aria-hidden
          >
            🦞
          </span>
          SkillsHub
        </Link>

        <nav className="hidden md:flex items-center gap-8 text-[15px] font-medium" style={{ color: 'hsl(var(--foreground))' }}>
          <Link
            to="/"
            className={
              isActive('/', true)
                ? 'text-foreground'
                : 'transition-opacity hover:opacity-70'
            }
            style={{ color: isActive('/', true) ? 'hsl(var(--foreground))' : 'hsl(var(--text-secondary))' }}
          >
            {t('nav.landing')}
          </Link>
          {navItems.map((item) => {
            if (item.auth && !user) return null
            const active = isActive(item.to, item.exact)
            return (
              <Link
                key={item.to}
                to={item.to}
                className="transition-opacity hover:opacity-70"
                style={{ color: active ? 'hsl(var(--foreground))' : 'hsl(var(--text-secondary))' }}
              >
                {item.label}
              </Link>
            )
          })}
        </nav>

        <div className="flex items-center gap-3 text-[15px] font-normal" style={{ color: 'hsl(var(--text-secondary))' }}>
          {user && (
            <Link
              to="/dashboard/publish"
              className="inline-flex items-center gap-1.5 rounded-full px-4 py-2 text-sm font-medium shadow-sm transition-opacity hover:opacity-90"
              style={{ background: 'hsl(var(--primary))', color: 'hsl(var(--primary-foreground))' }}
            >
              <span className="text-base leading-none">+</span>
              {t('nav.publish')}
            </Link>
          )}
          <LanguageSwitcher />
          {user && <NotificationBell />}
          {isLoading ? null : user ? (
            <UserMenu user={user} />
          ) : (
            <Link
              to="/login"
              search={{ returnTo: '' }}
              className="rounded-full px-4 py-2 text-sm font-medium transition-opacity hover:opacity-80"
              style={{ border: '1px solid hsl(var(--border))', color: 'hsl(var(--foreground))', background: 'hsl(var(--card))' }}
            >
              {t('nav.login')}
            </Link>
          )}
        </div>
      </header>

      {/* Main content */}
      <main className={mainContentLayout.mainClassName}>
        <Suspense
          fallback={
            <div className="space-y-4 animate-fade-up">
              <div className="h-10 w-48 animate-shimmer rounded-lg" />
              <div className="h-5 w-72 animate-shimmer rounded-md" />
              <div className="h-64 animate-shimmer rounded-xl" />
            </div>
          }
        >
          <div className={mainContentLayout.contentClassName}>
            <Outlet />
          </div>
        </Suspense>
      </main>
    </div>
  )
}
