import { useEffect, useRef, useState } from 'react'

const LINES = [
  '██████╗  █████╗ ████████╗ █████╗ ██╗  ██╗██████╗  █████╗ ███████╗████████╗███████╗███╗   ██╗',
  '██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝████╗  ██║',
  '██║  ██║███████║   ██║   ███████║█████╔╝ ██████╔╝███████║█████╗     ██║   █████╗  ██╔██╗ ██║',
  '██║  ██║██╔══██║   ██║   ██╔══██║██╔═██╗ ██╔══██╗██╔══██║██╔══╝     ██║   ██╔══╝  ██║╚██╗██║',
  '██████╔╝██║  ██║   ██║   ██║  ██║██║  ██╗██║  ██║██║  ██║██║        ██║   ███████╗██║ ╚████║',
  '╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═══╝',
]

export default function LogoSvg() {
  const shellRef = useRef<HTMLDivElement>(null)
  const preRef = useRef<HTMLPreElement>(null)
  const [scale, setScale] = useState(1)
  const [height, setHeight] = useState(0)

  useEffect(() => {
    const shell = shellRef.current
    const pre = preRef.current
    if (!shell || !pre) {
      return
    }

    const updateScale = () => {
      const shellWidth = shell.clientWidth
      const contentWidth = pre.scrollWidth
      const contentHeight = pre.scrollHeight
      if (!shellWidth || !contentWidth || !contentHeight) {
        return
      }

      const nextScale = Math.min(1, shellWidth / contentWidth)
      setScale(nextScale)
      setHeight(contentHeight * nextScale)
    }

    updateScale()

    const observer = new ResizeObserver(updateScale)
    observer.observe(shell)
    window.addEventListener('resize', updateScale)

    return () => {
      observer.disconnect()
      window.removeEventListener('resize', updateScale)
    }
  }, [])

  return (
    <div
      ref={shellRef}
      className="logo-shell"
      aria-label="Datakraften logo"
      role="img"
      style={height > 0 ? { height: `${height}px` } : undefined}
    >
      <div
        className="logo-stage"
        style={{
          transform: `translateX(-50%) scale(${scale})`,
        }}
      >
        <pre ref={preRef} className="logo-pre">
          {LINES.join('\n')}
        </pre>
      </div>
    </div>
  )
}
