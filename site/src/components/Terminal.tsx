import { useState } from 'react'

const tabs = [
  { id: 'dk', label: 'dk CLI', command: 'curl -fsSL https://datakraften.no/install | bash' },
  {
    id: 'bash',
    label: 'Bash (legacy)',
    command: 'bash <(curl -fsSL https://datakraften.no/wsl/debian.sh)',
  },
]

export default function Terminal() {
  const [activeTab, setActiveTab] = useState('dk')
  const [showToast, setShowToast] = useState(false)

  const activeCommand = tabs.find((t) => t.id === activeTab)?.command ?? ''

  const copyCommand = () => {
    navigator.clipboard.writeText(activeCommand).then(() => {
      setShowToast(true)
      setTimeout(() => setShowToast(false), 2000)
    })
  }

  return (
    <div className="rounded-lg overflow-hidden border border-fuchsia-500/30 bg-[#1a1a2e] max-w-2xl mx-auto">
      <div className="flex items-center gap-1.5 px-3 py-2 bg-black/30">
        <span className="w-3 h-3 rounded-full bg-red-500/80" />
        <span className="w-3 h-3 rounded-full bg-yellow-500/80" />
        <span className="w-3 h-3 rounded-full bg-green-500/80" />
        <span className="ml-4 text-sm text-text-dim font-jetbrains">
          {activeTab === 'dk' ? 'dk-cli' : 'bash-legacy'}
        </span>
      </div>
      <div className="flex border-b border-fuchsia-500/20">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm font-jetbrains transition-colors ${
              activeTab === tab.id
                ? 'bg-fuchsia-500/10 text-magenta border-b-2 border-magenta'
                : 'text-text-dim hover:text-text-primary'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>
      <div className="p-4 font-jetbrains text-base leading-relaxed flex justify-between items-start">
        <div>
          <span className="text-prompt">$</span>
          <span className="ml-2 text-text-primary">{activeCommand}</span>
          <span className="inline-block w-[0.15em] h-[1em] bg-magenta ml-3 animate-cursor-blink align-text-top relative top-[2px]" />
        </div>
        <button
          onClick={copyCommand}
          className="text-text-dim hover:text-magenta transition-colors flex-shrink-0 ml-4"
          title="Copy command"
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            strokeLinejoin="round"
          >
            <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
          </svg>
        </button>
      </div>
      <div className="px-4 pb-2 text-sm text-text-dim">
        {activeTab === 'dk'
          ? 'Install the Datakraften CLI'
          : 'Legacy bash installer for WSL Debian/Ubuntu'}
      </div>

      {showToast && (
        <div className="fixed bottom-6 left-1/2 -translate-x-1/2 bg-[#1a1a2e] border border-fuchsia-500/30 px-4 py-2 rounded text-base text-text-primary font-jetbrains z-50 animate-blink">
          ✓ Copied to clipboard
        </div>
      )}
    </div>
  )
}
