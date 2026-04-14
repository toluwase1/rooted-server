import { useLocation, useNavigate } from 'react-router-dom'

const tabs = [
  { path: '/', label: 'Circle', icon: '◎' },
  { path: '/explore', label: 'Explore', icon: '◈' },
  { path: '/matches', label: 'Matches', icon: '♡' },
  { path: '/settings', label: 'Settings', icon: '⚙' },
]

export default function Nav() {
  const location = useLocation()
  const navigate = useNavigate()

  return (
    <nav className="nav">
      {tabs.map((tab) => (
        <div
          key={tab.path}
          className={`nav-item ${location.pathname === tab.path ? 'active' : ''}`}
          onClick={() => navigate(tab.path)}
        >
          <span className="nav-icon">{tab.icon}</span>
          <span>{tab.label}</span>
        </div>
      ))}
    </nav>
  )
}
