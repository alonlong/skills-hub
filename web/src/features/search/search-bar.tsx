import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Loader2, Search, X } from 'lucide-react'
import { MAX_SEARCH_QUERY_LENGTH } from '@/shared/lib/search-query'
import { Input } from '@/shared/ui/input'
import { Button } from '@/shared/ui/button'

interface SearchBarProps {
  defaultValue?: string
  value?: string
  placeholder?: string
  isSearching?: boolean
  onChange?: (query: string) => void
  onSearch?: (query: string) => void
}

/**
 * Shared search bar used by search-driven pages and landing surfaces.
 *
 * It supports both controlled and uncontrolled usage so page-level containers can decide whether
 * query text should be driven from URL state or local form state.
 */
export function SearchBar({ defaultValue = '', value, placeholder, isSearching = false, onChange, onSearch }: SearchBarProps) {
  const { t } = useTranslation()
  const [query, setQuery] = useState(defaultValue)
  const isControlled = value !== undefined
  const currentQuery = isControlled ? value : query

  useEffect(() => {
    if (!isControlled) {
      setQuery(defaultValue)
    }
  }, [defaultValue, isControlled])

  const handleChange = (nextQuery: string) => {
    if (!isControlled) {
      setQuery(nextQuery)
    }
    onChange?.(nextQuery)
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (onSearch) {
      onSearch(currentQuery)
    }
  }

  const handleClear = () => {
    handleChange('')
    onSearch?.('')
  }

  return (
    <form onSubmit={handleSubmit} className="flex gap-2 items-center">
      <div className="relative flex-1">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground pointer-events-none" />
        <Input
          type="text"
          value={currentQuery}
          onChange={(e) => handleChange(e.target.value)}
          maxLength={MAX_SEARCH_QUERY_LENGTH}
          placeholder={placeholder || t('searchBar.placeholder')}
          className="pl-9 pr-9 h-11 rounded-lg"
        />
        {currentQuery ? (
          <button
            type="button"
            onClick={handleClear}
            className="absolute right-2 top-1/2 inline-flex h-6 w-6 -translate-y-1/2 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-secondary/70 hover:text-foreground"
            aria-label={t('searchBar.clear')}
            title={t('searchBar.clear')}
          >
            <X className="h-3.5 w-3.5" />
          </button>
        ) : null}
      </div>
      <Button type="submit" className="h-11 px-6 min-w-24" disabled={isSearching}>
        {isSearching ? <Loader2 className="h-4 w-4 animate-spin" /> : t('searchBar.button')}
      </Button>
    </form>
  )
}
