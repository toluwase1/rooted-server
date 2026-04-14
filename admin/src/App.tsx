import { Routes, Route, useLocation, useNavigate, Navigate } from 'react-router-dom'
import { isLoggedIn, clearToken } from './api'
import Login from './pages/Login'
import Dashboard from './pages/Dashboard'
import Config from './pages/Config'
import Features from './pages/Features'
import Reports from './pages/Reports'
import Users from './pages/Users'
import Logs from './pages/Logs'

const navItems = [
  { path: '/', label: 'Dashboard', icon: '◎' },
  { path: '/users', label: 'Users', icon: '◉' },
  { path: '/config', label: 'Config', icon: '⚙' },
  { path: '/features', label: 'Features', icon: '⚡' },
  { path: '/reports', label: 'Reports', icon: '⚠' },
  { path: '/logs', label: 'Logs', icon: '▤' },
]

function AuthGuard({ children }: { children: React.ReactNode }) {
  if (!isLoggedIn()) {
    return <Navigate to="/login" replace />
  }
  return <>{children}</>
}

export default function App() {
  const location = useLocation()
  const navigate = useNavigate()

  if (location.pathname === '/login') {
    return (
      <Routes>
        <Route path="/login" element={<Login />} />
      </Routes>
    )
  }

  return (
    <AuthGuard>
      <div className="layout">
        <aside className="sidebar">
          <div className="sidebar-logo">Rooted Admin</div>
          <nav className="sidebar-nav">
            {navItems.map((item) => (
              <div key={item.path}
                className={`sidebar-link ${location.pathname === item.path ? 'active' : ''}`}
                onClick={() => navigate(item.path)}>
                <span>{item.icon}</span>
                <span>{item.label}</span>
              </div>
            ))}
          </nav>
          <div style={{ marginTop: 'auto', padding: '16px 20px', borderTop: '1px solid var(--border)' }}>
            <div className="sidebar-link" onClick={() => { clearToken(); navigate('/login') }}>
              <span>↩</span>
              <span>Logout</span>
            </div>
          </div>
        </aside>
        <main className="main">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/users" element={<Users />} />
            <Route path="/config" element={<Config />} />
            <Route path="/features" element={<Features />} />
            <Route path="/reports" element={<Reports />} />
            <Route path="/logs" element={<Logs />} />
          </Routes>
        </main>
      </div>
    </AuthGuard>
  )
}
