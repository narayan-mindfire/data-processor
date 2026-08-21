import React from 'react';
import { cn } from '../../utils/cn';

interface ProgressBarProps extends React.HTMLAttributes<HTMLDivElement> {
  progress: number;
}

export function ProgressBar({ progress, className, ...props }: ProgressBarProps) {
  const safeProgress = Math.min(Math.max(progress, 0), 100);
  
  return (
    <div
      className={cn('h-2 w-full overflow-hidden rounded-full bg-secondary', className)}
      {...props}
    >
      <div
        className="h-full bg-primary transition-all duration-500 ease-in-out"
        style={{ width: `${safeProgress}%` }}
      />
    </div>
  );
}
