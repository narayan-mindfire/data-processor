import { Menu } from 'lucide-react';
import { Button } from '../ui/Button';

interface HeaderProps {
  onMenuClick: () => void;
}

export function Header({ onMenuClick }: HeaderProps) {
  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/80 backdrop-blur-sm px-4 sm:px-6">
      <div className="flex items-center">
        <Button variant="ghost" size="sm" className="md:hidden mr-2 p-2" onClick={onMenuClick}>
          <Menu className="h-5 w-5" />
        </Button>
        <h2 className="text-lg font-semibold md:hidden">Data Processor</h2>
      </div>

      <div className="flex items-center gap-4">
        <div className="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center border border-indigo-200">
          <span className="text-sm font-medium text-indigo-700">U</span>
        </div>
      </div>
    </header>
  );
}
