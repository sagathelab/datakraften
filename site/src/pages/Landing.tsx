import Layout from '../components/Layout'
import Terminal from '../components/Terminal'
import FeatureCard from '../components/FeatureCard'

export default function Landing() {
  return (
    <Layout variant="landing">
      <section className="hero text-center pt-12 pb-8">
        <pre className="font-share-tech text-magenta text-[10px] sm:text-xs leading-tight animate-logo-fade whitespace-pre">
{`██████╗  █████╗ ████████╗ █████╗ ██╗  ██╗██████╗  █████╗ ███████╗████████╗███████╗███╗   ██╗
██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝████╗  ██║
██║  ██║███████║   ██║   ███████║█████╔╝ ██████╔╝███████║█████╗     ██║   █████╗  ██╔██╗ ██║
██║  ██║██╔══██║   ██║   ██╔══██║██╔═██╗ ██╔══██╗██╔══██║██╔══╝     ██║   ██╔══╝  ██║╚██╗██║
██████╔╝██║  ██║   ██║   ██║  ██║██║  ██╗██║  ██║██║  ██║██║        ██║   ███████╗██║ ╚████║
╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═══╝`}
        </pre>
        <p className="font-share-tech text-text-dim text-sm mt-4">
          The WSL-first developer workstation platform
        </p>
      </section>

      <section className="install-section pt-6 pb-12">
        <h2 className="font-share-tech text-magenta text-center mb-6 text-lg">Install</h2>
        <Terminal />
      </section>

      <section className="features pb-12">
        <h2 className="font-share-tech text-magenta text-center mb-6 text-lg">Features</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <FeatureCard
            title="~ Runtimes"
            desc="Node.js (fnm), Python (uv), .NET SDK -- install and manage with a single command"
            to="/docs/#runtimes"
          />
          <FeatureCard
            title="~ Shell &amp; CLI"
            desc="Fish shell, Starship prompt, Atuin history, fzf, broot, fd, btm -- power-user defaults"
            to="/docs/#shell-cli"
          />
          <FeatureCard
            title="~ Cloud &amp; Dev"
            desc="GitHub CLI, Azure CLI, Docker, Homebrew -- everything for cloud-native development"
            to="/docs/#cloud-dev"
          />
          <FeatureCard
            title="~ Editors &amp; Tools"
            desc="VS Code, Zed, Cursor, Codex CLI, OpenCode, PowerShell -- AI-powered development"
            to="/docs/#editors-tools"
          />
        </div>
      </section>

      <section className="how-to pb-12">
        <h2 className="font-share-tech text-magenta text-center mb-6 text-lg">Get started</h2>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-2 text-xs font-jetbrains">
          <Step num="01" text="Install" />
          <Arrow />
          <Step num="02" text="dk init" />
          <Arrow />
          <Step num="03" text="dk apply" />
          <Arrow />
          <Step num="04" text="dk doctor" />
          <Arrow />
          <Step num="05" text="dk update" />
        </div>
      </section>
    </Layout>
  )
}

function Step({ num, text }: { num: string; text: string }) {
  return (
    <div className="flex flex-col items-center gap-1">
      <span className="text-magenta font-bold">{num}</span>
      <span className="text-text-dim">{text}</span>
    </div>
  )
}

function Arrow() {
  return (
    <svg className="w-5 h-5 text-text-dim hidden sm:block" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M5 12h14M13 5l7 7-7 7" />
    </svg>
  )
}
