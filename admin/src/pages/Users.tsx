import { useEffect, useState } from 'react'
import { adminApi } from '../api'

export default function Users() {
  const [users, setUsers] = useState<any[]>([])
  const [total, setTotal] = useState(0)
  const [search, setSearch] = useState('')
  const [statusFilter, setStatusFilter] = useState('')
  const [loading, setLoading] = useState(true)
  const [selectedUser, setSelectedUser] = useState<any>(null)

  useEffect(() => { loadUsers() }, [search, statusFilter])

  const loadUsers = () => {
    setLoading(true)
    adminApi.listUsers(search, statusFilter).then((data) => {
      setUsers(data.users || [])
      setTotal(data.total || 0)
      setLoading(false)
    }).catch(() => setLoading(false))
  }

  const loadUserDetail = async (id: string) => {
    const data = await adminApi.getUser(id)
    setSelectedUser(data)
  }

  const handleStatusChange = async (userId: string, status: string) => {
    await adminApi.updateUserStatus(userId, status)
    loadUsers()
    if (selectedUser?.user?.id === userId) {
      loadUserDetail(userId)
    }
  }

  const statusBadge = (status: string) => {
    const colors: Record<string, string> = {
      active: 'badge-green', banned: 'badge-red', paused: 'badge-yellow',
      deleted: 'badge-red', suspended: 'badge-yellow',
    }
    return colors[status] || 'badge-green'
  }

  return (
    <div>
      <h1 className="page-title">Users ({total})</h1>

      <div style={{ display: 'flex', gap: '12px', marginBottom: '20px' }}>
        <input placeholder="Search by name or Telegram ID..."
          value={search} onChange={(e) => setSearch(e.target.value)}
          style={{ flex: 1 }} />
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)}
          style={{ width: '140px' }}>
          <option value="">All statuses</option>
          <option value="active">Active</option>
          <option value="paused">Paused</option>
          <option value="banned">Banned</option>
          <option value="deleted">Deleted</option>
        </select>
      </div>

      <div style={{ display: 'flex', gap: '20px' }}>
        {/* User list */}
        <div style={{ flex: 1 }}>
          {loading ? (
            <div style={{ color: 'var(--text-dim)' }}>Loading...</div>
          ) : (
            <div className="card" style={{ padding: 0, overflow: 'hidden' }}>
              <table>
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Status</th>
                    <th>Verified</th>
                    <th>Plan</th>
                    <th>Joined</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((u: any) => (
                    <tr key={u.id} style={{ cursor: 'pointer' }}
                      onClick={() => loadUserDetail(u.id)}>
                      <td>
                        <div style={{ fontWeight: '600' }}>{u.first_name || 'No profile'}</div>
                        <div style={{ fontSize: '11px', color: 'var(--text-dim)' }}>
                          {u.city && `${u.city}, `}{u.country || ''} · TG: {u.telegram_id}
                        </div>
                      </td>
                      <td><span className={`badge ${statusBadge(u.status)}`}>{u.status}</span></td>
                      <td>{u.verification === 'photo_verified' ? '✓' : '—'}</td>
                      <td style={{ textTransform: 'capitalize' }}>{u.subscription}</td>
                      <td style={{ fontSize: '12px', color: 'var(--text-dim)' }}>
                        {new Date(u.created_at).toLocaleDateString()}
                      </td>
                    </tr>
                  ))}
                  {users.length === 0 && (
                    <tr><td colSpan={5} style={{ textAlign: 'center', color: 'var(--text-dim)' }}>No users found</td></tr>
                  )}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* User detail panel */}
        {selectedUser && (
          <div style={{ width: '320px', flexShrink: 0 }}>
            <div className="card">
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
                <h3>{selectedUser.profile?.first_name || 'No profile'}</h3>
                <button className="btn btn-secondary btn-sm" onClick={() => setSelectedUser(null)}>✕</button>
              </div>

              <div className="card-title">User Info</div>
              <div style={{ fontSize: '13px', display: 'flex', flexDirection: 'column', gap: '6px', marginBottom: '16px' }}>
                <div>ID: <code style={{ fontSize: '11px' }}>{selectedUser.user.id}</code></div>
                <div>Telegram: {selectedUser.user.telegram_id}</div>
                <div>Status: <span className={`badge ${statusBadge(selectedUser.user.status)}`}>{selectedUser.user.status}</span></div>
                <div>Verification: {selectedUser.user.verification}</div>
                <div>Trust Score: {selectedUser.user.trust_score}/100</div>
                <div>Plan: {selectedUser.user.subscription}</div>
                <div>Joined: {new Date(selectedUser.user.created_at).toLocaleString()}</div>
              </div>

              {/* Photos */}
              {selectedUser.photos?.length > 0 && (
                <>
                  <div className="card-title">Photos ({selectedUser.photos.length})</div>
                  <div style={{
                    display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)',
                    gap: '6px', marginBottom: '16px',
                  }}>
                    {selectedUser.photos.map((photo: any) => (
                      <div key={photo.id} style={{
                        aspectRatio: '1', borderRadius: '8px', overflow: 'hidden',
                        cursor: 'pointer', position: 'relative',
                        border: photo.moderation_status === 'rejected' ? '2px solid var(--red)' : '1px solid var(--border)',
                      }}
                        onClick={() => window.open(photo.url_large, '_blank')}
                      >
                        <img src={photo.url_medium || photo.url_thumbnail} alt=""
                          style={{ width: '100%', height: '100%', objectFit: 'cover' }} />
                        {photo.is_primary && (
                          <span style={{
                            position: 'absolute', bottom: '2px', left: '2px',
                            background: 'var(--primary)', color: '#fff',
                            padding: '1px 5px', borderRadius: '4px', fontSize: '9px',
                          }}>Primary</span>
                        )}
                        {photo.moderation_status !== 'approved' && (
                          <span style={{
                            position: 'absolute', top: '2px', right: '2px',
                            background: photo.moderation_status === 'rejected' ? 'var(--red)' : 'var(--yellow)',
                            color: '#fff', padding: '1px 5px', borderRadius: '4px', fontSize: '9px',
                          }}>{photo.moderation_status}</span>
                        )}
                      </div>
                    ))}
                  </div>
                </>
              )}

              {selectedUser.profile?.city && (
                <>
                  <div className="card-title">Profile</div>
                  <div style={{ fontSize: '13px', display: 'flex', flexDirection: 'column', gap: '4px', marginBottom: '16px' }}>
                    <div>Location: {selectedUser.profile.city}, {selectedUser.profile.country}</div>
                    <div>Heritage: {selectedUser.profile.heritage?.join(', ')}</div>
                    <div>Completeness: {selectedUser.profile.completeness}%</div>
                    {selectedUser.profile.bio && <div>Bio: {selectedUser.profile.bio}</div>}
                  </div>
                </>
              )}

              <div className="card-title">Activity</div>
              <div style={{ fontSize: '13px', display: 'flex', flexDirection: 'column', gap: '4px', marginBottom: '16px' }}>
                <div>Matches: {selectedUser.stats?.matches || 0}</div>
                <div>Messages sent: {selectedUser.stats?.messages || 0}</div>
                <div>Reports against: {selectedUser.stats?.reports || 0}</div>
              </div>

              <div className="card-title">Actions</div>
              <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
                {selectedUser.user.status !== 'active' && (
                  <button className="btn btn-sm" style={{ background: 'var(--green)', color: '#fff' }}
                    onClick={() => handleStatusChange(selectedUser.user.id, 'active')}>
                    Activate
                  </button>
                )}
                {selectedUser.user.status === 'active' && (
                  <button className="btn btn-sm" style={{ background: 'var(--yellow)', color: '#000' }}
                    onClick={() => handleStatusChange(selectedUser.user.id, 'paused')}>
                    Suspend
                  </button>
                )}
                {selectedUser.user.status !== 'banned' && (
                  <button className="btn btn-danger btn-sm"
                    onClick={() => handleStatusChange(selectedUser.user.id, 'banned')}>
                    Ban
                  </button>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
