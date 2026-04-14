import { useEffect, useState } from 'react'
import { adminApi } from '../api'

export default function Reports() {
  const [reports, setReports] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadReports()
  }, [])

  const loadReports = () => {
    adminApi.getReports().then((data) => {
      setReports(data.reports || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  const handleAction = async (id: string, action: string) => {
    await adminApi.reviewReport(id, action)
    loadReports()
  }

  if (loading) return <div style={{ color: 'var(--text-dim)' }}>Loading...</div>

  return (
    <div>
      <h1 className="page-title">Reports ({reports.length} pending)</h1>

      {reports.length === 0 && (
        <div className="card" style={{ textAlign: 'center', color: 'var(--text-dim)' }}>
          No pending reports
        </div>
      )}

      {reports.map((report: any) => (
        <div key={report.id} className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div>
              <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                <span className={`badge ${
                  report.category === 'scam' ? 'badge-red' :
                  report.category === 'fake_profile' ? 'badge-yellow' : 'badge-green'
                }`}>
                  {report.category}
                </span>
                {report.reported_name && (
                  <span style={{ fontWeight: '600' }}>{report.reported_name}</span>
                )}
              </div>
              {report.description && (
                <p style={{ fontSize: '14px', color: 'var(--text-dim)', marginBottom: '8px' }}>
                  {report.description}
                </p>
              )}
              <div style={{ fontSize: '12px', color: 'var(--text-dim)' }}>
                Reported: {new Date(report.created_at).toLocaleString()}
              </div>
              <div style={{ fontSize: '11px', color: 'var(--text-dim)', fontFamily: 'monospace' }}>
                Reporter: {report.reporter_id.slice(0, 8)}... → Reported: {report.reported_id.slice(0, 8)}...
              </div>
            </div>

            <div style={{ display: 'flex', gap: '6px', flexShrink: 0 }}>
              <button className="btn btn-secondary btn-sm"
                onClick={() => handleAction(report.id, 'dismiss')}>
                Dismiss
              </button>
              <button className="btn btn-sm" style={{ background: 'var(--yellow)', color: '#000' }}
                onClick={() => handleAction(report.id, 'warning')}>
                Warn
              </button>
              <button className="btn btn-danger btn-sm"
                onClick={() => handleAction(report.id, 'ban')}>
                Ban
              </button>
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
