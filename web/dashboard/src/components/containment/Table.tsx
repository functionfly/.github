import type { ReactNode } from 'react';

export interface TableColumn<T> {
  key: string;
  header: string;
  render?: (row: T) => ReactNode;
  align?: 'left' | 'right' | 'center';
  isNumeric?: boolean;
}

export interface TableProps<T> {
  columns: TableColumn<T>[];
  data: T[];
  emptyMessage?: string;
  className?: string;
  keyExtractor?: (row: T) => string;
}

export function Table<T>({
  columns,
  data,
  emptyMessage = 'No data available',
  className = '',
  keyExtractor,
}: TableProps<T>) {
  if (data.length === 0) {
    return (
      <div className="table-empty">
        <span className="table-empty__text">{emptyMessage}</span>
      </div>
    );
  }

  return (
    <div className={`table-wrapper ${className}`}>
      <table className="table">
        <thead className="table__header">
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                className={`table__header-cell ${col.align ? `table__header-cell--${col.align}` : ''}`}
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="table__body">
          {data.map((row, index) => {
            const key = keyExtractor ? keyExtractor(row) : String(index);
            return (
              <tr key={key} className="table__row">
                {columns.map((col) => {
                  const content = col.render
                    ? col.render(row)
                    : (row as Record<string, unknown>)[col.key];
                  return (
                    <td
                      key={col.key}
                      className={`table__cell ${col.isNumeric ? 'table__cell--numeric' : ''} ${
                        col.align ? `table__cell--${col.align}` : ''
                      }`}
                    >
                      {content as ReactNode}
                    </td>
                  );
                })}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}
