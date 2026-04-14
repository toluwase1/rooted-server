import { useEffect, useState } from 'react'
import { adminApi } from '../api'

export default function Logs() {
  const [logs, setLogs] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [errorsOnly, setErrorsOnly] = useState(false)

  useEffect(() => {
    loadLogs()
  }, [errorsOnly])

  const loadLogs = () => {
    setLoading(true)
    adminApi.getLogs(100, 0, errorsOnly).then((data) => {
      setLogs(data.logs || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  const statusColor = (code: number) => {
    if (code >= 500) return 'var(--red)'
    if (code >= 400) return 'var(--yellow)'
    return 'var(--green)'
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <h1 className="page-title" style={{ marginBottom: 0 }}>Request Logs</h1>
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          <label style={{ fontSize: '13px', color: 'var(--text-dim)', display: 'flex', alignItems: 'center', gap: '6px' }}>
            <input type="checkbox" checked={errorsOnly}
              onChange={(e) => setErrorsOnly(e.target.checked)} />
            Errors only
          </label>
          <button className="btn btn-secondary btn-sm" onClick={loadLogs}>Refresh</button>
        </div>
      </div>

      {loading ? (
        <div style={{ color: 'var(--text-dim)' }}>Loading...</div>
      ) : logs.length === 0 ? (
        <div className="card" style={{ textAlign: 'center', color: 'var(--text-dim)', padding: '40px' }}>
          {errorsOnly ? 'No errors found' : 'No logs yet. Logs will appear once the request logging middleware is active.'}
        </div>
      ) : (
        <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
          <table>
            <thead>
              <tr>
                <th>Time</th>
                <th>Method</th>
                <th>Path</th>
                <th>Status</th>
                <th>Latency</th>
                <th>Error</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log: any) => (
                <tr key={log.id}>
                  <td style={{ fontSize: '12px', color: 'var(--text-dim)', whiteSpace: 'nowrap' }}>
                    {new Date(log.created_at).toLocaleTimeString()}
                  </td>
                  <td><code style={{ fontSize: '12px' }}>{log.method}</code></td>
                  <td style={{ fontSize: '13px', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {log.path}
                  </td>
                  <td>
                    <span style={{ color: statusColor(log.status_code), fontWeight: '600', fontSize: '13px' }}>
                      {log.status_code}
                    </span>
                  </td>
                  <td style={{ fontSize: '12px', color: 'var(--text-dim)' }}>{log.latency_ms}ms</td>
                  <td style={{ fontSize: '12px', color: 'var(--red)', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {log.error || '—'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
