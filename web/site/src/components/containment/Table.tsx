import type { ReactNode, ThHTMLAttributes, TdHTMLAttributes } from 'react';

interface TableProps {
  children: ReactNode;
}

export function Table({ children }: TableProps) {
  return (
    <table className="table">
      {children}
    </table>
  );
}

interface TableHeaderProps {
  children: ReactNode;
}

export function TableHeader({ children }: TableHeaderProps) {
  return (
    <thead>
      <tr className="table__header">
        {children}
      </tr>
    </thead>
  );
}

interface ThProps extends ThHTMLAttributes<HTMLTableCellElement> {
  children: ReactNode;
  numeric?: boolean;
}

export function Th({ children, numeric = false, style, ...props }: ThProps) {
  return (
    <th
      {...props}
      className={`table__th ${numeric ? 'table__th--numeric' : ''}`}
      style={style}
    >
      {children}
    </th>
  );
}

interface TableBodyProps {
  children: ReactNode;
}

export function TableBody({ children }: TableBodyProps) {
  return <tbody>{children}</tbody>;
}

interface TrProps {
  children: ReactNode;
}

export function Tr({ children }: TrProps) {
  return (
    <tr className="table__tr">
      {children}
    </tr>
  );
}

interface TdProps extends TdHTMLAttributes<HTMLTableCellElement> {
  children: ReactNode;
  numeric?: boolean;
}

export function Td({ children, numeric = false, style, ...props }: TdProps) {
  return (
    <td
      {...props}
      className={`table__td ${numeric ? 'table__td--numeric' : ''}`}
      style={style}
    >
      {children}
    </td>
  );
}
