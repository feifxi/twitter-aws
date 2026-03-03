import { XLogo } from '../XLogo';

export function LoadingScreen() {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-background">
      <div className="flex flex-col items-center">
        <XLogo className="w-20 h-20 text-foreground animate-pulse" />
      </div>
    </div>
  );
}
