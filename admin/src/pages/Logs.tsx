import { useEffect, useState } from 'react'
import { adminApi } from '../api'

export default function Logs() {
  const [entries, setEntries] = useState<any[]>([])
  const [stats, setStats] = useState<any>(null)
  const [loading, setLoading] = useState(true)
  const [timeRange, setTimeRange] = useState('1h')
  const [severity, setSeverity] = useState('')
  const [search, setSearch] = useState('')
  const [view, setView] = useState<'logs' | 'stats'>('stats')
  const [selectedEntry, setSelectedEntry] = useState<any>(null)

  useEffect(() => {
    loadData()
  }, [timeRange, severity, view])

  const loadData = async () => {
    setLoading(true)
    try {
      if (view === 'stats') {
        const data = await adminApi.getLogStats(timeRange)
        setStats(data)
      } else {
        const data = await adminApi.getLogs(100, timeRange, severity, search)
        setEntries(data.entries || [])
      }
    } catch {
      // Cloud Logging may not be available locally
    }
    setLoading(false)
  }

  const statusColor = (code: number) => {
    if (code >= 500) return 'var(--red)'
    if (code >= 400) return 'var(--yellow)'
    return 'var(--green)'
  }

  const severityBadge = (sev: string) => {
    const colors: Record<string, string> = {
      ERROR: 'badge-red', WARNING: 'badge-yellow', INFO: 'badge-green', DEFAULT: 'badge-green',
    }
    return colors[sev] || 'badge-green'
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
        <h1 className="page-title" style={{ marginBottom: 0 }}>Logs</h1>
        <div style={{ display: 'flex', gap: '8px' }}>
          <button className={`btn btn-sm ${view === 'stats' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setView('stats')}>Overview</button>
          <button className={`btn btn-sm ${view === 'logs' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setView('logs')}>Entries</button>
        </div>
      </div>

      {/* Filters */}
      <div style={{ display: 'flex', gap: '8px', marginBottom: '20px', flexWrap: 'wrap' }}>
        {['1h', '6h', '24h', '7d'].map((r) => (
          <button key={r} className={`btn btn-sm ${timeRange === r ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setTimeRange(r)}>
            {r}
          </button>
        ))}
        {view === 'logs' && (
          <>
            <select value={severity} onChange={(e) => setSeverity(e.target.value)}
              style={{ width: '120px', padding: '4px 8px', fontSize: '13px' }}>
              <option value="">All levels</option>
              <option value="ERROR">Error</option>
              <option value="WARNING">Warning</option>
              <option value="INFO">Info</option>
            </select>
            <input placeholder="Search..." value={search}
              onChange={(e) => setSearch(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && loadData()}
              style={{ width: '180px', padding: '4px 8px', fontSize: '13px' }} />
            <button className="btn btn-secondary btn-sm" onClick={loadData}>Search</button>
          </>
        )}
      </div>

      {loading ? (
        <div style={{ color: 'var(--text-dim)' }}>Loading...</div>
      ) : view === 'stats' && stats ? (
        <>
          {/* Stats overview */}
          <div className="stat-grid">
            <div className="stat-card">
              <div className="stat-value">{stats.total_entries || 0}</div>
              <div className="stat-label">Total Requests</div>
            </div>
            <div className="stat-card">
              <div className="stat-value" style={{ color: stats.total_errors > 0 ? 'var(--red)' : 'var(--text)' }}>
                {stats.total_errors || 0}
              </div>
              <div className="stat-label">Errors ({(stats.error_rate || 0).toFixed(1)}%)</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{(stats.avg_response_ms || 0).toFixed(0)}ms</div>
              <div className="stat-label">Avg Response</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {Object.entries(stats.by_severity || {}).map(([k, v]) => (
                  <span key={k} style={{ fontSize: '14px', marginRight: '8px' }}>
                    <span className={`badge ${severityBadge(k)}`}>{k}: {v as number}</span>
                  </span>
                ))}
              </div>
              <div className="stat-label">By Severity</div>
            </div>
          </div>

          {/* Top errors */}
          {stats.top_errors?.length > 0 && (
            <div className="card">
              <div className="card-title">Top Errors</div>
              {stats.top_errors.map((e: any, i: number) => (
                <div key={i} style={{
                  display: 'flex', justifyContent: 'space-between', padding: '6px 0',
                  borderBottom: '1px solid var(--border)', fontSize: '13px',
                }}>
                  <span style={{ color: 'var(--red)', maxWidth: '80%', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                    {e.message}
                  </span>
                  <span style={{ fontWeight: '600' }}>{e.count}x</span>
                </div>
              ))}
            </div>
          )}

          {/* Top paths */}
          {stats.top_paths?.length > 0 && (
            <div className="card">
              <div className="card-title">Top Endpoints</div>
              <table style={{ fontSize: '13px' }}>
                <thead>
                  <tr><th>Endpoint</th><th>Requests</th><th>Avg ms</th><th>Errors</th></tr>
                </thead>
                <tbody>
                  {stats.top_paths.map((p: any, i: number) => (
                    <tr key={i}>
                      <td><code>{p.method} {p.path}</code></td>
                      <td>{p.count}</td>
                      <td>{p.avg_ms.toFixed(0)}ms</td>
                      <td style={{ color: p.errors > 0 ? 'var(--red)' : 'var(--text-dim)' }}>{p.errors}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </>
      ) : view === 'logs' ? (
        <div style={{ display: 'flex', gap: '16px' }}>
          {/* Log table */}
          <div className="card" style={{ padding: 0, overflow: 'hidden', flex: 1 }}>
            {entries.length === 0 ? (
              <div style={{ padding: '40px', textAlign: 'center', color: 'var(--text-dim)' }}>
                No log entries found
              </div>
            ) : (
              <table>
                <thead>
                  <tr>
                    <th>Time</th>
                    <th>Level</th>
                    <th>Method</th>
                    <th>Path</th>
                    <th>Status</th>
                    <th>Latency</th>
                    <th>Message</th>
                  </tr>
                </thead>
                <tbody>
                  {entries.map((e: any) => (
                    <tr key={e.id} onClick={() => setSelectedEntry(e)}
                      style={{
                        cursor: 'pointer',
                        background: selectedEntry?.id === e.id ? 'var(--border)' : undefined,
                      }}>
                      <td style={{ fontSize: '11px', color: 'var(--text-dim)', whiteSpace: 'nowrap' }}>
                        {new Date(e.timestamp).toLocaleTimeString()}
                      </td>
                      <td><span className={`badge ${severityBadge(e.severity)}`}>{e.severity}</span></td>
                      <td><code style={{ fontSize: '11px' }}>{e.method}</code></td>
                      <td style={{ fontSize: '12px', maxWidth: '180px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {e.path}
                      </td>
                      <td>
                        {e.status_code > 0 && (
                          <span style={{ color: statusColor(e.status_code), fontWeight: '600', fontSize: '12px' }}>
                            {e.status_code}
                          </span>
                        )}
                      </td>
                      <td style={{ fontSize: '11px', color: 'var(--text-dim)' }}>
                        {e.duration_ms > 0 ? `${e.duration_ms}ms` : ''}
                      </td>
                      <td style={{ fontSize: '12px', maxWidth: '200px', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                        {e.error || e.message || ''}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>

          {/* Detail panel */}
          {selectedEntry && (
            <div style={{ width: '360px', flexShrink: 0 }}>
              <div className="card">
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                  <h3 style={{ fontSize: '16px' }}>Request Detail</h3>
                  <button className="btn btn-secondary btn-sm" onClick={() => setSelectedEntry(null)}>✕</button>
                </div>

                {/* Summary */}
                <div style={{ display: 'flex', gap: '8px', marginBottom: '16px', flexWrap: 'wrap' }}>
                  <span className={`badge ${severityBadge(selectedEntry.severity)}`}>{selectedEntry.severity}</span>
                  {selectedEntry.status_code > 0 && (
                    <span style={{
                      color: statusColor(selectedEntry.status_code),
                      fontWeight: '600', fontSize: '13px',
                    }}>
                      HTTP {selectedEntry.status_code}
                    </span>
                  )}
                  {selectedEntry.duration_ms > 0 && (
                    <span style={{ fontSize: '13px', color: 'var(--text-dim)' }}>
                      {selectedEntry.duration_ms}ms
                    </span>
                  )}
                </div>

                {/* Request info */}
                <div className="card-title">Request</div>
                <div style={{ fontSize: '13px', marginBottom: '16px' }}>
                  <div style={{ marginBottom: '4px' }}>
                    <code style={{ fontSize: '12px', color: 'var(--primary)' }}>
                      {selectedEntry.method} {selectedEntry.path}
                    </code>
                  </div>
                  <div style={{ color: 'var(--text-dim)' }}>
                    {new Date(selectedEntry.timestamp).toLocaleString()}
                  </div>
                  {selectedEntry.service && (
                    <div style={{ color: 'var(--text-dim)', marginTop: '2px' }}>
                      Service: {selectedEntry.service}
                    </div>
                  )}
                  {selectedEntry.telegram_id > 0 && (
                    <div style={{ color: 'var(--text-dim)', marginTop: '2px' }}>
                      Telegram ID: {selectedEntry.telegram_id}
                    </div>
                  )}
                </div>

                {/* Error message */}
                {selectedEntry.error && (
                  <>
                    <div className="card-title">Error</div>
                    <div style={{
                      fontSize: '13px', color: 'var(--red)',
                      background: '#EF444410', padding: '10px',
                      borderRadius: '8px', marginBottom: '16px',
                      whiteSpace: 'pre-wrap', wordBreak: 'break-word',
                    }}>
                      {selectedEntry.error}
                    </div>
                  </>
                )}

                {/* Message */}
                {selectedEntry.message && selectedEntry.message !== selectedEntry.error && (
                  <>
                    <div className="card-title">Message</div>
                    <div style={{
                      fontSize: '13px', marginBottom: '16px',
                      whiteSpace: 'pre-wrap', wordBreak: 'break-word',
                    }}>
                      {selectedEntry.message}
                    </div>
                  </>
                )}

                {/* Raw payload */}
                {selectedEntry.raw_payload && Object.keys(selectedEntry.raw_payload).length > 0 && (
                  <>
                    <div className="card-title">Raw Payload</div>
                    <pre style={{
                      fontSize: '11px', color: 'var(--text-dim)',
                      background: 'var(--bg)', padding: '10px',
                      borderRadius: '8px', overflow: 'auto',
                      maxHeight: '300px', whiteSpace: 'pre-wrap',
                      wordBreak: 'break-word',
                    }}>
                      {JSON.stringify(selectedEntry.raw_payload, null, 2)}
                    </pre>
                  </>
                )}
              </div>
            </div>
          )}
        </div>
      ) : null}
    </div>
  )
}
