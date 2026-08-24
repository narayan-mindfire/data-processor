import { Menu, Bell } from 'lucide-react';
import { Button } from '../ui/Button';

export function Header() {
  return (
    <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-border bg-background/80 backdrop-blur-sm px-4 sm:px-6">
      <div className="flex items-center">
        <Button variant="ghost" size="sm" className="md:hidden mr-2 p-2">
          <Menu className="h-5 w-5" />
        </Button>
        <h2 className="text-lg font-semibold md:hidden">Data Processor</h2>
      </div>

      <div className="flex items-center gap-4">
        <Button variant="ghost" size="sm" className="relative p-2 rounded-full">
          <Bell className="h-5 w-5 text-muted-foreground" />
          <span className="absolute top-1 right-1 flex h-2 w-2">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-primary opacity-75"></span>
            <span className="relative inline-flex rounded-full h-2 w-2 bg-primary"></span>
          </span>
        </Button>
        <div className="h-8 w-8 rounded-full bg-indigo-100 flex items-center justify-center border border-indigo-200">
          <span className="text-sm font-medium text-indigo-700">U</span>
        </div>
      </div>
    </header>
  );
}
