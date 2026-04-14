import { useEffect, useState } from 'react'
import { adminApi } from '../api'

export default function Features() {
  const [flags, setFlags] = useState<any[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    adminApi.getFeatures().then((data) => {
      setFlags(data || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [])

  const toggle = async (key: string, enabled: boolean) => {
    await adminApi.updateFeature(key, { enabled })
    setFlags((prev) => prev.map((f) => f.key === key ? { ...f, enabled } : f))
  }

  const updateRollout = async (key: string, percent: number) => {
    await adminApi.updateFeature(key, { rollout_percent: percent })
    setFlags((prev) => prev.map((f) => f.key === key ? { ...f, rollout_percent: percent } : f))
  }

  if (loading) return <div style={{ color: 'var(--text-dim)' }}>Loading...</div>

  return (
    <div>
      <h1 className="page-title">Feature Flags</h1>

      {flags.map((flag: any) => (
        <div key={flag.key} className="card">
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <div style={{ fontWeight: '600', fontFamily: 'monospace' }}>{flag.key}</div>
              <div style={{ fontSize: '13px', color: 'var(--text-dim)', marginTop: '4px' }}>
                {flag.description}
              </div>
            </div>
            <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
              <span className={`badge ${flag.enabled ? 'badge-green' : 'badge-red'}`}>
                {flag.enabled ? 'ON' : 'OFF'}
              </span>
              <button
                className={`btn btn-sm ${flag.enabled ? 'btn-danger' : 'btn-primary'}`}
                onClick={() => toggle(flag.key, !flag.enabled)}>
                {flag.enabled ? 'Disable' : 'Enable'}
              </button>
            </div>
          </div>

          {flag.enabled && (
            <div style={{ marginTop: '12px', display: 'flex', alignItems: 'center', gap: '12px' }}>
              <label style={{ fontSize: '13px', color: 'var(--text-dim)' }}>Rollout:</label>
              <input type="range" min="0" max="100" value={flag.rollout_percent}
                onChange={(e) => updateRollout(flag.key, parseInt(e.target.value))}
                style={{ width: '200px' }} />
              <span style={{ fontSize: '14px', fontWeight: '600' }}>{flag.rollout_percent}%</span>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
