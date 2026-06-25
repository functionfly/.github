import React from 'react';
import { StatusPill } from './StatusPill';
import { TrustSeal } from './TrustSeal';

interface Column<T> {
  key: keyof T | string;
  header: string;
  width?: string;
  align?: 'left' | 'center' | 'right';
  render?: (row: T) => React.ReactNode;
}

interface TableProps<T extends { id: string | number }> {
  columns: Column<T>[];
  data: T[];
  emptyMessage?: string;
  className?: string;
}

export function Table<T extends { id: string | number }>({
  columns,
  data,
  emptyMessage = 'No data available',
  className = '',
}: TableProps<T>) {
  if (data.length === 0) {
    return (
      <div
        className="flex items-center justify-center py-[var(--space-8)]"
        role="status"
      >
        <p
          className="font-[var(--font-mono)] text-[11px] uppercase tracking-widest text-[var(--text-faint)]"
        >
          {emptyMessage}
        </p>
      </div>
    );
  }

  return (
    <div className={`overflow-x-auto ${className}`}>
      <table
        className="w-full border-collapse"
        style={{ fontSize: '13px' }}
      >
        <thead>
          <tr style={{ borderBottom: '1px solid var(--panel-edge)' }}>
            {columns.map((column) => (
              <th
                key={String(column.key)}
                className="font-[var(--font-mono)] text-[11px] uppercase tracking-widest text-[var(--text-faint)] font-medium py-[var(--space-3)] px-[var(--space-4)] text-left"
                style={{
                  width: column.width,
                  textAlign: column.align ?? 'left',
                }}
              >
                {column.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row) => (
            <tr
              key={row.id}
              className="transition-colors duration-[var(--duration-fast)]"
              style={{ borderBottom: '1px solid var(--panel-edge)' }}
            >
              {columns.map((column) => (
                <td
                  key={String(column.key)}
                  className="py-[var(--space-3)] px-[var(--space-4)] text-[var(--text-dim)]"
                  style={{
                    textAlign: column.align ?? 'left',
                  }}
                >
                  {column.render
                    ? column.render(row)
                    : String(row[column.key as keyof T] ?? '')}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

interface StatusCellProps {
  status: 'live' | 'pending' | 'revoked';
  label: string;
}

export const StatusCell: React.FC<StatusCellProps> = ({ status, label }) => {
  return <StatusPill status={status} label={label} />;
};

interface TrustSealCellProps {
  variant?: 'live' | 'verified' | 'trust';
  label: string;
  size?: 'sm' | 'md' | 'lg';
}

export const TrustSealCell: React.FC<TrustSealCellProps> = ({
  variant = 'verified',
  label,
  size = 'sm',
}) => {
  return <TrustSeal label={label} variant={variant} size={size} />;
};

interface NumericCellProps {
  value: string | number;
  align?: 'right' | 'left' | 'center';
}

export const NumericCell: React.FC<NumericCellProps> = ({
  value,
  align = 'right',
}) => {
  return (
    <span
      className="font-[var(--font-mono)] text-[15px] font-medium"
      style={{ textAlign: align }}
    >
      {value}
    </span>
  );
};
