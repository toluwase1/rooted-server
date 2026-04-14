import { useEffect, useState } from 'react'
import { api } from '../api/client'

interface Props {
  user: any
}

export default function Premium({ user }: Props) {
  const [plans, setPlans] = useState<any[]>([])
  const [items, setItems] = useState<any[]>([])
  const [selectedPlan, setSelectedPlan] = useState<string>('plus')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getPlans().then((data) => {
      setPlans(data.plans || [])
      setItems(data.items || [])
      setLoading(false)
    }).catch(() => setLoading(false))
  }, [])

  if (loading) return <div className="loading">Loading plans...</div>

  const currentPlan = plans.find((p: any) => p.plan === selectedPlan)

  return (
    <div className="container page">
      <h1 className="page-header">Upgrade</h1>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '20px' }}>
        Get more matches, see who liked you, and unlock premium features.
      </p>

      {/* Plan selector */}
      <div style={{ display: 'flex', gap: '8px', marginBottom: '20px' }}>
        {plans.map((plan: any) => (
          <button key={plan.plan}
            className={`btn ${selectedPlan === plan.plan ? 'btn-primary' : 'btn-secondary'}`}
            style={{ flex: 1 }}
            onClick={() => setSelectedPlan(plan.plan)}>
            <div>
              <div style={{ fontWeight: '600' }}>{plan.label}</div>
              <div style={{ fontSize: '12px', opacity: 0.8 }}>
                {plan.star_price} Stars/mo
              </div>
            </div>
          </button>
        ))}
      </div>

      {/* Selected plan features */}
      {currentPlan && (
        <div className="card">
          <h3 style={{ marginBottom: '12px' }}>{currentPlan.label}</h3>
          {currentPlan.features.map((feature: string, i: number) => (
            <div key={i} style={{
              display: 'flex', alignItems: 'center', gap: '8px',
              padding: '6px 0', fontSize: '14px',
            }}>
              <span style={{ color: 'var(--accent)' }}>✓</span>
              <span>{feature}</span>
            </div>
          ))}

          <button className="btn btn-primary" style={{ marginTop: '16px' }}>
            Subscribe — {currentPlan.star_price} Stars/month
          </button>

          <p style={{ fontSize: '11px', color: 'var(--text-secondary)', textAlign: 'center', marginTop: '8px' }}>
            Auto-renews monthly. Cancel anytime.
          </p>
        </div>
      )}

      {/* A la carte items */}
      <h3 style={{ marginTop: '24px', marginBottom: '12px' }}>One-time purchases</h3>
      {items.map((item: any) => (
        <div key={item.type} className="card" style={{
          display: 'flex', justifyContent: 'space-between', alignItems: 'center',
        }}>
          <div>
            <div style={{ fontWeight: '600' }}>{item.label}</div>
            <div style={{ fontSize: '13px', color: 'var(--text-secondary)' }}>
              {item.description}
            </div>
          </div>
          <button className="btn btn-outline" style={{ width: 'auto', padding: '8px 16px', fontSize: '13px' }}>
            {item.star_price} ★
          </button>
        </div>
      ))}

      {/* Comparison */}
      <div style={{ marginTop: '24px', padding: '16px', background: 'var(--card-bg)', borderRadius: 'var(--radius)' }}>
        <h4 style={{ marginBottom: '12px' }}>Free vs Plus vs Premium</h4>
        <table style={{ width: '100%', fontSize: '13px', borderCollapse: 'collapse' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid var(--border)' }}>
              <th style={{ textAlign: 'left', padding: '8px 0' }}>Feature</th>
              <th style={{ textAlign: 'center', padding: '8px 4px' }}>Free</th>
              <th style={{ textAlign: 'center', padding: '8px 4px' }}>Plus</th>
              <th style={{ textAlign: 'center', padding: '8px 4px' }}>Premium</th>
            </tr>
          </thead>
          <tbody>
            {[
              ['Daily matches', '5-8', '15', '15+'],
              ['Explore swipes', '15/day', '∞', '∞'],
              ['See who liked you', '—', '✓', '✓'],
              ['Advanced filters', '—', '✓', '✓'],
              ['Read receipts', '—', '✓', '✓'],
              ['Crossed Continents', '—', '—', '✓'],
              ['Incognito mode', '—', '—', '✓'],
              ['Priority matching', '—', '—', '✓'],
              ['Message translation', '—', '—', '✓'],
            ].map(([feature, free, plus, premium]) => (
              <tr key={feature} style={{ borderBottom: '1px solid var(--border)' }}>
                <td style={{ padding: '8px 0' }}>{feature}</td>
                <td style={{ textAlign: 'center', padding: '8px 4px' }}>{free}</td>
                <td style={{ textAlign: 'center', padding: '8px 4px' }}>{plus}</td>
                <td style={{ textAlign: 'center', padding: '8px 4px' }}>{premium}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
