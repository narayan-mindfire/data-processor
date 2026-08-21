import React from 'react';
import { cn } from '../../utils/cn';
import { JobStatus } from '../../types/models';

interface BadgeProps extends React.HTMLAttributes<HTMLDivElement> {
  status: JobStatus;
}

export function Badge({ status, className, ...props }: BadgeProps) {
  const statusStyles: Record<JobStatus, string> = {
    pending: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-500 border-yellow-200',
    running: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-500 border-blue-200',
    completed: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-500 border-green-200',
    failed: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-500 border-red-200',
    cancelled: 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-400 border-gray-200',
  };

  return (
    <div
      className={cn(
        'inline-flex items-center rounded-full border px-2.5 py-0.5 text-xs font-semibold transition-colors focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2',
        statusStyles[status],
        className
      )}
      {...props}
    >
      {status.charAt(0).toUpperCase() + status.slice(1)}
    </div>
  );
}
