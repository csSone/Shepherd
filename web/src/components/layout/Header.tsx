import { Bell } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { ThemeToggle } from '@/components/ui/ThemeToggle';
import { LanguageToggle } from '@/components/ui/LanguageToggle';

export function Header() {
  const { t } = useTranslation();

  return (
    <header className="flex h-16 items-center justify-end border-b bg-background px-6">
      <div className="flex items-center gap-2">
        <LanguageToggle />
        <ThemeToggle />
        <Button variant="ghost" size="icon" aria-label={t('header.notifications')}>
          <Bell size={20} />
        </Button>
      </div>
    </header>
  );
}
