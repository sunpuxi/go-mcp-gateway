import { useState, useEffect, useRef } from 'react'

/**
 * 数字计数动画 Hook
 * @param end 目标值
 * @param duration 动画时长 (ms)
 * @param start 是否启动（可用于延迟触发）
 */
export function useCountUp(end: number, duration = 800, start = true) {
  const [value, setValue] = useState(0)
  const rafRef = useRef<number | null>(null)

  useEffect(() => {
    if (!start) {
      setValue(end)
      return
    }

    let startTime: number | null = null
    const startValue = 0

    const animate = (timestamp: number) => {
      if (startTime === null) startTime = timestamp
      const elapsed = timestamp - startTime
      const progress = Math.min(elapsed / duration, 1)
      // easeOutCubic
      const eased = 1 - Math.pow(1 - progress, 3)
      setValue(Math.round(startValue + (end - startValue) * eased))
      if (progress < 1) {
        rafRef.current = requestAnimationFrame(animate)
      }
    }

    rafRef.current = requestAnimationFrame(animate)

    return () => {
      if (rafRef.current !== null) {
        cancelAnimationFrame(rafRef.current)
      }
    }
  }, [end, duration, start])

  return value
}
