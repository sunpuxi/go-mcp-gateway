interface LogoProps {
  size?: number
}

function Logo({ size = 32 }: LogoProps) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 48 48"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      {/* 外层六边形 — 网关主体 */}
      <polygon
        points="24,2 44,14 44,34 24,46 4,34 4,14"
        fill="#1a1a2e"
        stroke="#1677ff"
        strokeWidth="2.5"
        strokeLinejoin="round"
      />
      {/* 内部菱形 — 核心路由 */}
      <polygon
        points="24,10 38,18 24,26 10,18"
        fill="none"
        stroke="#1677ff"
        strokeWidth="2"
        strokeLinejoin="round"
      />
      {/* 左侧箭头 — 协议输入 */}
      <line x1="2" y1="24" x2="10" y2="20" stroke="#4fc3f7" strokeWidth="2" strokeLinecap="round" />
      <polyline points="2,24 8,22 8,26" fill="none" stroke="#4fc3f7" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
      {/* 右侧箭头 — 协议输出 */}
      <line x1="38" y1="20" x2="46" y2="24" stroke="#4fc3f7" strokeWidth="2" strokeLinecap="round" />
      <polyline points="40,26 40,22 46,24" fill="none" stroke="#4fc3f7" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
      {/* 中心节点 */}
      <circle cx="24" cy="24" r="3.5" fill="#1677ff" />
      {/* 上下连接点 */}
      <circle cx="24" cy="13" r="2" fill="#4fc3f7" />
      <circle cx="24" cy="35" r="2" fill="#4fc3f7" />
      <circle cx="14" cy="18" r="2" fill="#4fc3f7" />
      <circle cx="34" cy="18" r="2" fill="#4fc3f7" />
      <circle cx="14" cy="30" r="2" fill="#4fc3f7" />
      <circle cx="34" cy="30" r="2" fill="#4fc3f7" />
      {/* 连接线 — 数据流 */}
      <line x1="14" y1="18" x2="24" y2="24" stroke="#4fc3f7" strokeWidth="1" opacity="0.6" />
      <line x1="34" y1="18" x2="24" y2="24" stroke="#4fc3f7" strokeWidth="1" opacity="0.6" />
      <line x1="14" y1="30" x2="24" y2="24" stroke="#4fc3f7" strokeWidth="1" opacity="0.6" />
      <line x1="34" y1="30" x2="24" y2="24" stroke="#4fc3f7" strokeWidth="1" opacity="0.6" />
      <line x1="24" y1="13" x2="24" y2="24" stroke="#4fc3f7" strokeWidth="1" opacity="0.6" />
      <line x1="24" y1="35" x2="24" y2="24" stroke="#4fc3f7" strokeWidth="1" opacity="0.6" />
    </svg>
  )
}

export default Logo
