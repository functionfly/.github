import { Outlet, Link } from 'react-router-dom'
import './Layout.css'

export default function Layout() {
  return (
    <div className="layout">
      <header className="header">
        <div className="container header-content">
          <Link to="/" className="logo">
            <svg width="32" height="32" viewBox="0 0 32 32" fill="none">
              <circle cx="16" cy="16" r="14" stroke="currentColor" strokeWidth="2" />
              <path d="M10 16L14 20L22 12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
            <span>FunctionFly</span>
          </Link>
          <nav className="nav">
            <Link to="/" className="nav-link">Functions</Link>
            <a href="https://functionfly.com/docs" className="nav-link">Docs</a>
            <a href="https://functionfly.com" className="nav-link">Home</a>
          </nav>
        </div>
      </header>

      <main className="main">
        <Outlet />
      </main>

      <footer className="footer">
        <div className="container">
          <p>&copy; {new Date().getFullYear()} FunctionFly. Open source serverless functions.</p>
        </div>
      </footer>
    </div>
  )
}
