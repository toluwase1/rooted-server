import { useEffect, useState } from 'react'
import { adminApi } from '../api'

export default function Config() {
  const [configs, setConfigs] = useState<any[]>([])
  const [loading, setLoading] = useState(true)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [editValue, setEditValue] = useState('')
  const [editReason, setEditReason] = useState('')
  const [filterCategory, setFilterCategory] = useState('')
  const [search, setSearch] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    loadConfigs()
  }, [])

  const loadConfigs = () => {
    adminApi.getConfig().then((data) => {
      setConfigs(data || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  const categories = [...new Set(configs.map((c: any) => c.category))].sort()

  const filtered = configs.filter((c: any) => {
    if (filterCategory && c.category !== filterCategory) return false
    if (search) {
      const q = search.toLowerCase()
      return c.key.toLowerCase().includes(q) || (c.description || '').toLowerCase().includes(q)
    }
    return true
  })

  const handleSave = async () => {
    if (!editingKey || !editReason) return
    setSaving(true)

    let parsedValue: any = editValue
    try { parsedValue = JSON.parse(editValue) } catch { /* keep as string */ }

    try {
      await adminApi.updateConfig(editingKey, parsedValue, editReason)
      setEditingKey(null)
      setEditValue('')
      setEditReason('')
      loadConfigs()
    } catch (err: any) {
      alert(err.message)
    }
    setSaving(false)
  }

  if (loading) return <div style={{ color: 'var(--text-dim)' }}>Loading...</div>

  return (
    <div>
      <h1 className="page-title">Configuration</h1>

      {/* Search */}
      <input
        type="text"
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search config keys..."
        style={{
          width: '100%', padding: '10px 14px', marginBottom: '16px',
          borderRadius: '8px', border: '1px solid var(--border)',
          background: 'var(--bg)', color: 'var(--text)', fontSize: '14px',
        }}
      />

      {/* Category filter */}
      <div style={{ display: 'flex', gap: '8px', marginBottom: '20px', flexWrap: 'wrap' }}>
        <button className={`btn ${!filterCategory ? 'btn-primary' : 'btn-secondary'} btn-sm`}
          onClick={() => setFilterCategory('')}>
          All ({configs.length})
        </button>
        {categories.map((cat) => (
          <button key={cat}
            className={`btn ${filterCategory === cat ? 'btn-primary' : 'btn-secondary'} btn-sm`}
            onClick={() => setFilterCategory(cat)}>
            {cat} ({configs.filter((c: any) => c.category === cat).length})
          </button>
        ))}
      </div>

      {/* Config table */}
      <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
        <table>
          <thead>
            <tr>
              <th>Key</th>
              <th>Value</th>
              <th>Category</th>
              <th>Description</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {filtered.map((cfg: any) => (
              <tr key={cfg.key}>
                <td style={{ fontFamily: 'monospace', fontSize: '13px' }}>{cfg.key}</td>
                <td>
                  {editingKey === cfg.key ? (
                    <div style={{ display: 'flex', flexDirection: 'column', gap: '6px' }}>
                      <input value={editValue} onChange={(e) => setEditValue(e.target.value)}
                        style={{ width: '200px' }} />
                      <input value={editReason} onChange={(e) => setEditReason(e.target.value)}
                        placeholder="Reason for change..." style={{ width: '200px' }} />
                      <div style={{ display: 'flex', gap: '4px' }}>
                        <button className="btn btn-primary btn-sm" onClick={handleSave} disabled={saving || !editReason}>
                          {saving ? '...' : 'Save'}
                        </button>
                        <button className="btn btn-secondary btn-sm" onClick={() => setEditingKey(null)}>
                          Cancel
                        </button>
                      </div>
                    </div>
                  ) : (
                    <code style={{ fontSize: '13px', color: 'var(--primary)' }}>{cfg.value}</code>
                  )}
                </td>
                <td><span className="badge badge-green">{cfg.category}</span></td>
                <td style={{ fontSize: '12px', color: 'var(--text-dim)', maxWidth: '200px' }}>
                  {cfg.description}
                </td>
                <td>
                  {editingKey !== cfg.key && (
                    <button className="btn btn-secondary btn-sm"
                      onClick={() => { setEditingKey(cfg.key); setEditValue(cfg.value); setEditReason('') }}>
                      Edit
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
