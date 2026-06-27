import React from "react";

interface Column<T> {
  key: string;
  header: string;
  render?: (row: T) => React.ReactNode;
  align?: "left" | "right" | "center";
  numeric?: boolean;
}

interface TableProps<T extends Record<string, unknown>> {
  columns: Column<T>[];
  data: T[];
  emptyMessage?: string;
}

export function Table<T extends Record<string, unknown>>({
  columns,
  data,
  emptyMessage = "No data available",
}: TableProps<T>) {
  if (data.length === 0) {
    return (
      <div
        style={{
          padding: "var(--space-7)",
          textAlign: "center",
          fontFamily: "var(--font-mono)",
          fontSize: "11px",
          fontWeight: 500,
          textTransform: "uppercase",
          letterSpacing: "0.06em",
          color: "var(--text-faint)",
          background: "var(--panel)",
          border: "1px solid var(--panel-edge)",
          borderRadius: "var(--radius-lg)",
        }}
      >
        {emptyMessage}
      </div>
    );
  }

  return (
    <div
      className="table-wrapper"
      style={{
        background: "var(--panel)",
        border: "1px solid var(--panel-edge)",
        borderRadius: "var(--radius-lg)",
        overflow: "hidden",
      }}
    >
      <table
        style={{
          width: "100%",
          borderCollapse: "collapse",
        }}
      >
        <thead>
          <tr>
            {columns.map((col) => (
              <th
                key={col.key}
                style={{
                  fontFamily: "var(--font-mono)",
                  fontSize: "11px",
                  fontWeight: 500,
                  textTransform: "uppercase",
                  letterSpacing: "0.06–0.08em",
                  color: "var(--text-faint)",
                  textAlign: col.align || "left",
                  padding: "var(--space-3) var(--space-4)",
                  borderBottom: "1px solid var(--panel-edge)",
                  whiteSpace: "nowrap",
                }}
              >
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {data.map((row, i) => (
            <tr
              key={i}
              className="table-row"
              style={{
                borderBottom: "1px solid var(--panel-edge)",
                transition: "background var(--duration-fast) var(--ease-out)",
              }}
            >
              {columns.map((col) => (
                <td
                  key={col.key}
                  style={{
                    fontFamily: col.numeric
                      ? "var(--font-mono)"
                      : "var(--font-body)",
                    fontSize: "13px",
                    lineHeight: 1.5,
                    color: "var(--text)",
                    textAlign: col.align || "left",
                    padding: "var(--space-3) var(--space-4)",
                  }}
                >
                  {col.render ? col.render(row) : String(row[col.key] ?? "")}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
