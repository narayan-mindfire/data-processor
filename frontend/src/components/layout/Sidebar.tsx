import { NavLink } from 'react-router-dom';
import { LayoutDashboard, PlusCircle, X } from 'lucide-react';
import { cn } from '../../utils/cn';
import { Button } from '../ui/Button';

interface SidebarProps {
  isOpen: boolean;
  onClose: () => void;
}

export function Sidebar({ isOpen, onClose }: SidebarProps) {
  const links = [
    { name: 'Dashboard', to: '/', icon: LayoutDashboard },
    { name: 'New Job', to: '/jobs/new', icon: PlusCircle },
  ];

  return (
    <aside 
      className={cn(
        "fixed inset-y-0 left-0 z-40 w-64 border-r border-border bg-card p-4 transition-transform duration-300 ease-in-out md:relative md:translate-x-0 flex flex-col shadow-xl md:shadow-none",
        isOpen ? "translate-x-0" : "-translate-x-full"
      )}
    >
      <div className="flex items-center justify-between mb-8 px-2">
        <div className="flex items-center">
          <div className="w-8 h-8 rounded bg-primary flex items-center justify-center mr-3">
            <span className="text-primary-foreground font-bold text-lg">D</span>
          </div>
          <h1 className="text-xl font-bold tracking-tight">Data Processor</h1>
        </div>
        <Button variant="ghost" size="sm" className="md:hidden p-1" onClick={onClose}>
          <X className="h-5 w-5" />
        </Button>
      </div>

      <nav className="flex-1 space-y-1">
        {links.map((link) => {
          const Icon = link.icon;
          return (
            <NavLink
              key={link.to}
              to={link.to}
              onClick={() => {
                if (window.innerWidth < 768) {
                  onClose();
                }
              }}
              className={({ isActive }) =>
                cn(
                  'flex items-center px-3 py-2 text-sm font-medium rounded-md transition-colors',
                  isActive
                    ? 'bg-primary/10 text-primary'
                    : 'text-muted-foreground hover:bg-muted hover:text-foreground'
                )
              }
            >
              <Icon className="mr-3 h-5 w-5 flex-shrink-0" />
              {link.name}
            </NavLink>
          );
        })}
      </nav>
    </aside>
  );
}
