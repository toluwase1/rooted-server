import { Routes, Route, useLocation, useNavigate } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Config from './pages/Config'
import Features from './pages/Features'
import Reports from './pages/Reports'

const navItems = [
  { path: '/', label: 'Dashboard', icon: '◎' },
  { path: '/config', label: 'Config', icon: '⚙' },
  { path: '/features', label: 'Features', icon: '⚡' },
  { path: '/reports', label: 'Reports', icon: '⚠' },
]

export default function App() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
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
      </aside>
      <main className="main">
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/config" element={<Config />} />
          <Route path="/features" element={<Features />} />
          <Route path="/reports" element={<Reports />} />
        </Routes>
      </main>
    </div>
  )
}
