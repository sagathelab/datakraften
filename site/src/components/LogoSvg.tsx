const LINES = [
  '██████╗  █████╗ ████████╗ █████╗ ██╗  ██╗██████╗  █████╗ ███████╗████████╗███████╗███╗   ██╗',
  '██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝████╗  ██║',
  '██║  ██║███████║   ██║   ███████║█████╔╝ ██████╔╝███████║█████╗     ██║   █████╗  ██╔██╗ ██║',
  '██║  ██║██╔══██║   ██║   ██╔══██║██╔═██╗ ██╔══██╗██╔══██║██╔══╝     ██║   ██╔══╝  ██║╚██╗██║',
  '██████╔╝██║  ██║   ██║   ██║  ██║██║  ██╗██║  ██║██║  ██║██║        ██║   ███████╗██║ ╚████║',
  '╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═══╝',
]

export default function LogoSvg() {
  return (
    <svg
      className="logo-svg"
      viewBox="0 0 920 200"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Datakraften logo"
    >
      <foreignObject x="0" y="0" width="920" height="200">
        {/* @ts-expect-error: xmlns needed for HTML inside foreignObject */}
        <div xmlns="http://www.w3.org/1999/xhtml" className="logo-fo">
          <pre className="logo-pre">{LINES.join('\n')}</pre>
        </div>
      </foreignObject>
    </svg>
  )
}
