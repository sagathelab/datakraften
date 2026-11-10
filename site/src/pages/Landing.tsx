import { Link } from 'react-router-dom'
import Layout from '../components/Layout'
import Terminal from '../components/Terminal'

export default function Landing() {
  return (
    <Layout variant="landing">
      <div className="product-page">
        <section className="product-hero">
          <div>
            <p className="product-eyebrow">OPEN SOURCE · LINUX · MACOS · WSL</p>
            <h1>A better starting point for your development machine.</h1>
            <p className="product-lead">
              Datakraften brings your runtimes, shell, editors and AI tools together in a
              configuration you can understand and share.
            </p>
            <div className="product-actions">
              <a className="product-primary" href="#get-started">
                Get started <span aria-hidden="true">↗</span>
              </a>
              <a className="product-secondary" href="https://github.com/sagathelab/datakraften">
                View on GitHub ↗
              </a>
            </div>
            <p className="product-caption">
              Built on the tools you already use. Apache 2.0 licensed.
            </p>
          </div>
          <div className="product-demo" aria-label="Example Datakraften workflow">
            <div className="product-demo-header">
              <span>WORKSTATION / PREVIEW</span>
              <span>dk</span>
            </div>
            <pre>
              <code>
                <span>$ dk init</span>
                {'\n'} Configure your workstation{'\n\n'}
                <span>$ dk apply --dry-run</span>
                {'\n'} 01 / 08 · System packages{'\n'} 02 / 08 · Developer tools{'\n'} 03 / 08 ·
                Runtimes{'\n'} …{'\n'} Preview complete; no changes made{'\n\n'}
                <span>$ dk doctor</span>
                {'\n'} Environment diagnostics
              </code>
            </pre>
            <p>Illustrative output. Results depend on your machine and configuration.</p>
          </div>
        </section>
        <section className="product-section product-intro">
          <p className="product-eyebrow">LESS SETUP. MORE MAKING.</p>
          <h2>One place for the tools that make your machine yours.</h2>
          <p>
            Stop piecing together a workstation from scattered instructions. Describe the tools you
            need, preview the plan, and let dk coordinate the setup.
          </p>
        </section>
        <section className="product-grid" aria-label="How it works">
          <article>
            <span>01 / CONFIGURE</span>
            <h3>Start with your setup.</h3>
            <p>
              Generate a default configuration or create your own. Choose the tools that fit your
              work.
            </p>
          </article>
          <article>
            <span>02 / PREVIEW</span>
            <h3>See before you change.</h3>
            <p>
              Run a dry run to review intended actions. Some checks depend on your local
              environment.
            </p>
          </article>
          <article>
            <span>03 / VERIFY</span>
            <h3>Know where you stand.</h3>
            <p>
              Apply your configuration, then use doctor to inspect your workstation and identify
              missing tools.
            </p>
          </article>
        </section>
        <section className="product-section product-split">
          <div>
            <p className="product-eyebrow">BUILT FOR PEOPLE AND TEAMS</p>
            <h2>A shared starting point. Room to make it yours.</h2>
            <p>
              Use a local configuration for your own machine, or point a team at a shared YAML file.
              Platform support and installed tools can still vary between machines.
            </p>
            <Link to="/docs/teams">Explore team configuration →</Link>
          </div>
          <pre>
            <code>
              {
                '# Local config\nsource: team\nurl: https://example.com/team.yaml\n\n# Remote team.yaml\nai_tools:\n  codex:\n    enabled: true\n    manager: npm'
              }
            </code>
          </pre>
        </section>
        <section className="product-section product-trust">
          <p className="product-eyebrow">UNDERSTAND THE TRADEOFFS</p>
          <h2>Your machine. Your choices.</h2>
          <p>
            Datakraften orchestrates existing tools including your system package manager, Homebrew,
            fnm and uv. Some system operations require sudo. Review the configuration and preview
            actions before applying them; availability and behavior vary across Linux, macOS and
            WSL.
          </p>
          <div className="product-actions">
            <Link to="/docs/dk">Read the CLI guide →</Link>
            <Link to="/docs/config">Explore configuration →</Link>
            <a href="https://github.com/sagathelab/datakraften">Inspect the source →</a>
          </div>
        </section>
        <section id="get-started" className="product-section product-start">
          <p className="product-eyebrow">READY WHEN YOU ARE</p>
          <h2>Start with the CLI.</h2>
          <p>
            This installs dk. Run <code>dk init</code> and preview with{' '}
            <code>dk apply --dry-run</code> before applying your configuration.
          </p>
          <Terminal />
          <p>
            <Link to="/docs/install">Installation details and requirements →</Link>
          </p>
        </section>
      </div>
    </Layout>
  )
}
