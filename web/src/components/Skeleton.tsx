interface SkeletonProps {
  width?: string | number
  height?: string | number
  borderRadius?: number
  style?: React.CSSProperties
}

export function Skeleton({ width = '100%', height = 16, borderRadius = 4, style }: SkeletonProps) {
  return (
    <div
      className="skeleton"
      style={{
        width,
        height,
        borderRadius,
        display: 'inline-block',
        ...style,
      }}
    />
  )
}

export function TableSkeleton({ rows = 5, cols = 4 }: { rows?: number; cols?: number }) {
  return (
    <div style={{ padding: 16 }}>
      {Array.from({ length: rows }, (_, r) => (
        <div key={r} style={{ display: 'flex', gap: 16, marginBottom: 12 }}>
          {Array.from({ length: cols }, (_, c) => (
            <Skeleton key={c} width={`${90 / cols}%`} height={18} />
          ))}
        </div>
      ))}
    </div>
  )
}

export function StatsSkeleton() {
  return (
    <div className="stats-grid">
      {Array.from({ length: 4 }, (_, i) => (
        <div key={i} className="stat-card">
          <Skeleton width={60} height={30} style={{ marginBottom: 8 }} />
          <Skeleton width={40} height={14} />
        </div>
      ))}
    </div>
  )
}
