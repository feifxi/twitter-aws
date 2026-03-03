'use client';

import { Moon, Sun } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useTheme } from 'next-themes';
import { useEffect, useState } from 'react';

interface ThemeToggleProps {
  className?: string;
}

export function ThemeToggle({ className }: ThemeToggleProps) {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <Button variant="ghost" size="icon" className={`rounded-full hover:bg-card text-foreground transition-colors ${className || ''}`} disabled>
        <span className="w-6 h-6 border-2 border-muted border-t-foreground rounded-full animate-spin"></span>
      </Button>
    )
  }

  return (
    <Button 
        variant="ghost" 
        size="icon" 
        onClick={() => setTheme(theme === 'light' ? 'dark' : 'light')}
        className={`rounded-full hover:bg-card text-foreground transition-colors ${className || ''}`}
        aria-label="Toggle theme"
    >
      {theme === 'light' ? (
        <Moon className="w-6 h-6" />
      ) : (
        <Sun className="w-6 h-6" />
      )}
    </Button>
  );
}
