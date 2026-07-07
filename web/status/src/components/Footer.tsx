import React from 'react';
import { Link } from 'react-router-dom';

export const Footer: React.FC = () => {
  return (
    <footer
      style={{
        borderTop: '1px solid var(--panel-edge)',
        marginTop: 'var(--space-8)',
      }}
    >
      <div
        className="mx-auto flex flex-col md:flex-row items-center justify-between gap-4"
        style={{
          maxWidth: '1180px',
          padding: 'var(--space-6) var(--space-7)',
        }}
      >
        <div
          className="flex items-center gap-2"
          style={{ fontSize: '13px', color: 'var(--text-faint)' }}
        >
          <div
            style={{
              width: '6px',
              height: '6px',
              clipPath: 'polygon(50% 0%, 100% 50%, 50% 100%, 0% 50%)',
              backgroundColor: 'var(--accent)',
            }}
          />
          <span>&copy; {new Date().getFullYear()} FunctionFly</span>
        </div>
        <div
          className="flex items-center gap-6"
          style={{ fontSize: '13px', color: 'var(--text-faint)' }}
        >
          <Link to="/" className="hover:opacity-80 transition-opacity" style={{ color: 'inherit', textDecoration: 'none' }}>
            Status
          </Link>
          <Link to="/history" className="hover:opacity-80 transition-opacity" style={{ color: 'inherit', textDecoration: 'none' }}>
            History
          </Link>
          <a
            href="https://docs.functionfly.com"
            className="hover:opacity-80 transition-opacity"
            style={{ color: 'inherit', textDecoration: 'none' }}
            target="_blank"
            rel="noopener noreferrer"
          >
            API
          </a>
        </div>
      </div>
    </footer>
  );
};
