import type { ReactNode } from 'react'
import Layout from '../components/Layout'
import Terminal from '../components/Terminal'
import FeatureCard from '../components/FeatureCard'
import LogoSvg from '../components/LogoSvg'
import { Link } from 'react-router-dom'

export default function Landing() {
  return (
    <Layout variant="landing">
      <section className="hero text-center pt-12 pb-10">
        <LogoSvg />
        <p className="font-share-tech text-text-dim uppercase tracking-[0.1em] text-base sm:text-lg mt-4">
          MAKE YOUR MACHINE SHIP-READY FOR CODING AND COLLABORATION
        </p>
        <p className="font-share-tech text-magenta uppercase tracking-[0.15em] text-sm sm:text-base mt-1 animate-tagline-glow">
          A BOOTSTRAP TOOL — FROM ONE DEVELOPER TO ANOTHER
        </p>
      </section>

      <section className="pb-14">
        <div className="grid grid-cols-1 lg:grid-cols-3 border-y border-fuchsia-500/20 divide-y lg:divide-y-0 lg:divide-x divide-fuchsia-500/15">
          <ValuePoint
            icon={<BoltIcon />}
            title="FAST SETUP"
            body="Bootstrap runtimes, shell, editors, and AI tools without stitching together setup guides by hand."
          />
          <ValuePoint
            icon={<LayersIcon />}
            title="CONSISTENT ENVIRONMENTS"
            body="Share one YAML config and give every developer the same workstation, tools, and defaults."
          />
          <ValuePoint
            icon={<RefreshIcon />}
            title="KEEP UP TO DATE"
            body="Stay current and move faster with a development environment that is easy to refresh as your tools evolve."
          />
        </div>
      </section>

      <section className="install-section pt-2 pb-14">
        <h2 className="font-share-tech text-magenta text-center mb-6 text-2xl">Install</h2>
        <Terminal />
      </section>

      <section className="features pb-12">
        <h2 className="font-share-tech text-magenta text-center mb-6 text-2xl">Features</h2>
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          <FeatureCard
            title="~ BUILD FAST"
            packages="Node.js (fnm), Python (uv), Go, .NET SDK"
            tagline="Modern runtimes ready for real development work"
            to="/docs#runtimes"
          />
          <FeatureCard
            title="~ UPGRADE YOUR SHELL"
            packages="Fish shell, Starship, Atuin, fzf, broot, fd, btm"
            tagline="A better terminal experience out of the box"
            to="/docs#shell-cli"
          />
          <FeatureCard
            title="~ WORK CLOUD-READY"
            packages="GitHub CLI, Azure CLI, Docker, Homebrew"
            tagline="Core tooling for modern cloud-native workflows"
            to="/docs#cloud-dev"
          />
          <FeatureCard
            title="~ CODE WITH AI"
            packages="VS Code, Zed, Cursor, Codex CLI, OpenCode, GitHub Copilot"
            tagline="Editors and AI tools ready from day one"
            to="/docs#editors-tools"
          />
        </div>
      </section>

      <section className="what-you-get pb-16 max-w-2xl mx-auto px-4">
        <h2 className="font-share-tech text-magenta text-center mb-8 text-2xl">What you get</h2>
        <div className="space-y-6 text-text-body text-sm sm:text-base leading-relaxed">
          <p>
            <span className="text-magenta font-bold">~</span> A complete, AI-ready development
            environment — from shell to editor to cloud tools — installed and configured with a
            single command. No more hour-long setup guides or scattered READMEs.
          </p>
          <p>
            <span className="text-magenta font-bold">~</span> Consistency across your team. Share a
            YAML config and every developer gets the same toolchain, runtimes, and shell config.
            Onboarding goes from days to minutes.
          </p>
          <p>
            <span className="text-magenta font-bold">~</span> Idempotent and transparent. Run it as
            many times as you want — it only installs what's missing. Everything is orchestrated
            through battle-tested tools you already know.
          </p>
        </div>
      </section>

      <section className="how-to pb-16">
        <h2 className="font-share-tech text-magenta text-center mb-8 text-2xl">Get started</h2>
        <div className="flex flex-col sm:flex-row items-center justify-center gap-4 sm:gap-3 font-jetbrains">
          <Step num="01" text="Install" desc="Get the CLI" to="/docs/dk" />
          <Arrow />
          <Step num="02" text="Configure" desc="Run dk init" to="/docs/dk#init" />
          <Arrow />
          <Step num="03" text="Apply" desc="Build your environment" to="/docs/dk#apply" />
        </div>
        <p className="mt-6 text-sm sm:text-base text-text-dim text-center leading-relaxed">
          Then verify with{' '}
          <Link to="/docs/dk#doctor" className="text-magenta hover:underline">
            dk doctor
          </Link>{' '}
          and keep it fresh with{' '}
          <Link to="/docs/dk#update" className="text-magenta hover:underline">
            dk update
          </Link>
          .
        </p>
      </section>
    </Layout>
  )
}

function Step({ num, text, desc, to }: { num: string; text: string; desc: string; to: string }) {
  return (
    <Link
      to={to}
      className="flex flex-col items-center gap-1.5 p-4 border border-fuchsia-500/20 rounded-lg min-w-[180px] bg-bg-card hover:border-magenta hover:bg-fuchsia-500/5 transition-all"
    >
      <span className="text-2xl font-bold text-magenta leading-none">{num}</span>
      <span className="text-base text-text-primary">{text}</span>
      <span className="text-sm text-text-dim text-center">{desc}</span>
    </Link>
  )
}

function ValuePoint({ icon, title, body }: { icon: ReactNode; title: string; body: string }) {
  return (
    <div className="flex items-start gap-4 px-2 py-6 lg:px-6">
      <div className="flex h-11 w-11 flex-shrink-0 items-center justify-center rounded-lg border border-fuchsia-500/30 bg-fuchsia-500/8 text-magenta">
        {icon}
      </div>
      <div>
        <h3 className="font-share-tech text-base text-magenta mb-1">{title}</h3>
        <p className="text-sm sm:text-base text-text-body leading-relaxed">{body}</p>
      </div>
    </div>
  )
}

function Arrow() {
  return (
    <svg
      className="w-6 h-6 text-text-dim hidden sm:block flex-shrink-0"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
    >
      <path d="M5 12h14M13 5l7 7-7 7" />
    </svg>
  )
}

function BoltIcon() {
  return (
    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z" />
    </svg>
  )
}

function LayersIcon() {
  return (
    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="m12 3 9 5-9 5-9-5 9-5Z" />
      <path d="m3 12 9 5 9-5" />
      <path d="m3 16 9 5 9-5" />
    </svg>
  )
}

function RefreshIcon() {
  return (
    <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M21 12a9 9 0 1 1-2.64-6.36" />
      <path d="M21 3v6h-6" />
    </svg>
  )
}
