import Layout from '../components/Layout'
import Terminal from '../components/Terminal'
import FeatureCard from '../components/FeatureCard'
import { Link } from 'react-router-dom'

export default function Landing() {
  return (
    <Layout variant="landing">
      <section className="hero text-center pt-12 pb-8">
        <pre className="logo-ascii" aria-label="Datakraften logo">
{`██████╗  █████╗ ████████╗ █████╗ ██╗  ██╗██████╗  █████╗ ███████╗████████╗███████╗███╗   ██╗
██╔══██╗██╔══██╗╚══██╔══╝██╔══██╗██║ ██╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██╔════╝████╗  ██║
██║  ██║███████║   ██║   ███████║█████╔╝ ██████╔╝███████║█████╗     ██║   █████╗  ██╔██╗ ██║
██║  ██║██╔══██║   ██║   ██╔══██║██╔═██╗ ██╔══██╗██╔══██║██╔══╝     ██║   ██╔══╝  ██║╚██╗██║
██████╔╝██║  ██║   ██║   ██║  ██║██║  ██╗██║  ██║██║  ██║██║        ██║   ███████╗██║ ╚████║
╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝        ╚═╝   ╚══════╝╚═╝  ╚═══╝`}
        </pre>
        <p className="font-share-tech text-text-dim uppercase tracking-[0.1em] text-base sm:text-lg mt-4">
          Bootstrap modern developer environments
        </p>
        <p className="font-share-tech text-magenta uppercase tracking-[0.15em] text-sm sm:text-base mt-1 animate-tagline-glow">
          for AI-powered development
        </p>
      </section>

      <section className="install-section pt-6 pb-12">
        <h2 className="font-share-tech text-magenta text-center mb-6 text-xl">Install</h2>
        <Terminal />
      </section>

      <section className="features pb-12">
        <h2 className="font-share-tech text-magenta text-center mb-6 text-xl">Features</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <FeatureCard
            title="~ Runtimes"
            packages="Node.js (fnm), Python (uv), .NET SDK"
            tagline="install and manage with a single command"
            to="/docs#runtimes"
          />
          <FeatureCard
            title="~ Shell & CLI"
            packages="Fish shell, Starship, Atuin, fzf, broot, fd, btm"
            tagline="power-user defaults for your terminal"
            to="/docs#shell-cli"
          />
          <FeatureCard
            title="~ Cloud & Dev"
            packages="GitHub CLI, Azure CLI, Docker, Homebrew"
            tagline="everything for cloud-native development"
            to="/docs#cloud-dev"
          />
          <FeatureCard
            title="~ Editors & Tools"
            packages="VS Code, Zed, Cursor, Codex CLI, OpenCode, GitHub Copilot, PowerShell"
            tagline="AI-powered development from day one"
            to="/docs#editors-tools"
          />
        </div>
      </section>

      <section className="how-to pb-16">
        <h2 className="font-share-tech text-magenta text-center mb-8 text-xl">Get started</h2>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-3 font-jetbrains">
          <Step num="01" text="Install" desc="Bootstrap your workstation" to="/docs/dk" />
          <Arrow />
          <Step num="02" text="dk init" desc="Generate your config" to="/docs/dk#init" />
          <Arrow />
          <Step num="03" text="dk apply" desc="Install everything" to="/docs/dk#apply" />
          <Arrow />
          <Step num="04" text="dk doctor" desc="Verify your setup" to="/docs/dk#doctor" />
          <Arrow />
          <Step num="05" text="dk update" desc="Stay up to date" to="/docs/dk#update" />
        </div>
      </section>
    </Layout>
  )
}

function Step({ num, text, desc, to }: { num: string; text: string; desc: string; to: string }) {
  return (
    <Link to={to} className="flex flex-col items-center gap-1.5 p-4 border border-fuchsia-500/20 rounded-lg min-w-[140px] bg-bg-card hover:border-magenta hover:bg-fuchsia-500/5 transition-all">
      <span className="text-2xl font-bold text-magenta leading-none">{num}</span>
      <span className="text-base text-text-primary">{text}</span>
      <span className="text-sm text-text-dim text-center">{desc}</span>
    </Link>
  )
}

function Arrow() {
  return (
    <svg className="w-6 h-6 text-text-dim hidden sm:block flex-shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M5 12h14M13 5l7 7-7 7" />
    </svg>
  )
}
