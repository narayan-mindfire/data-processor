import { NavLink } from 'react-router-dom';
import { LayoutDashboard, PlusCircle, Settings } from 'lucide-react';
import { cn } from '../../utils/cn';

export function Sidebar() {
  const links = [
    { name: 'Dashboard', to: '/', icon: LayoutDashboard },
    { name: 'New Job', to: '/jobs/new', icon: PlusCircle },
    { name: 'Settings', to: '/settings', icon: Settings },
  ];

  return (
    <aside className="w-64 border-r border-border bg-card min-h-screen p-4 hidden md:flex flex-col">
      <div className="flex items-center mb-8 px-2">
        <div className="w-8 h-8 rounded bg-primary flex items-center justify-center mr-3">
          <span className="text-primary-foreground font-bold text-lg">D</span>
        </div>
        <h1 className="text-xl font-bold tracking-tight">Data Processor</h1>
      </div>

      <nav className="flex-1 space-y-1">
        {links.map((link) => {
          const Icon = link.icon;
          return (
            <NavLink
              key={link.to}
              to={link.to}
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
