import { useEffect, useState } from 'react'
import { adminApi } from '../api'

export default function Dashboard() {
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    adminApi.getStats().then(setStats).catch(() => {}).finally(() => setLoading(false))
  }, [])

  if (loading) return <div style={{ color: 'var(--text-dim)' }}>Loading...</div>

  return (
    <div>
      <h1 className="page-title">Dashboard</h1>

      <div className="stat-grid">
        <div className="stat-card">
          <div className="stat-value">{stats?.total_users ?? 0}</div>
          <div className="stat-label">Total Users</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats?.active_users ?? 0}</div>
          <div className="stat-label">Active Users</div>
        </div>
        <div className="stat-card">
          <div className="stat-value">{stats?.total_matches ?? 0}</div>
          <div className="stat-label">Total Matches</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: stats?.pending_reports > 0 ? 'var(--red)' : 'var(--text)' }}>
            {stats?.pending_reports ?? 0}
          </div>
          <div className="stat-label">Pending Reports</div>
        </div>
      </div>
    </div>
  )
}
