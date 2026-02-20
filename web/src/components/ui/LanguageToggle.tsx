import { Languages, ChevronDown, Check } from 'lucide-react';
import { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { cn } from '@/lib/utils';
import { SUPPORTED_LANGUAGES, type SupportedLanguage } from '@/lib/i18n';

export function LanguageToggle() {
  const { i18n } = useTranslation();
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const currentLanguage = SUPPORTED_LANGUAGES.find(
    (lang) => lang.code === i18n.language
  );

  const handleLanguageChange = (languageCode: SupportedLanguage) => {
    i18n.changeLanguage(languageCode);
    setIsOpen(false);
  };

  return (
    <div ref={containerRef} className="relative">
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={cn(
          'flex items-center gap-1.5 rounded-lg px-2.5 py-1.5',
          'transition-all duration-200',
          'border border-border/40 hover:border-border/60',
          'bg-background/50 hover:bg-background/80',
          'hover:bg-accent',
          'focus:outline-none focus:ring-2 focus:ring-ring focus:border-primary/50'
        )}
        aria-label={`Select language (Current: ${currentLanguage?.nativeName})`}
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        title={currentLanguage?.nativeName}
      >
        <Languages size={16} />
        <ChevronDown
          size={12}
          className={cn(
            'transition-transform duration-200 text-muted-foreground',
            isOpen && 'rotate-180'
          )}
        />
      </button>

      {isOpen && (
        <div
          className={cn(
            'absolute right-0 top-full z-50 mt-1 w-36',
            'rounded-lg border bg-popover shadow-md overflow-hidden',
            'animate-in fade-in-0 zoom-in-95'
          )}
          role="listbox"
          aria-label="Language options"
        >
          <div className="py-0.5">
            {SUPPORTED_LANGUAGES.map((option) => {
              const isSelected = option.code === i18n.language;

              return (
                <button
                  key={option.code}
                  type="button"
                  onClick={() => handleLanguageChange(option.code)}
                  className={cn(
                    'flex w-full items-center justify-between gap-2 px-3 py-1.5 text-xs',
                    'transition-colors duration-150',
                    'hover:bg-accent hover:text-accent-foreground',
                    isSelected && 'bg-accent',
                    'focus:outline-none focus:bg-accent'
                  )}
                  role="option"
                  aria-selected={isSelected}
                >
                  <span
                    className={cn(
                      'truncate',
                      isSelected && 'font-medium'
                    )}
                  >
                    {option.nativeName}
                  </span>
                  {isSelected && (
                    <Check size={12} className="text-primary shrink-0" />
                  )}
                </button>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
}
