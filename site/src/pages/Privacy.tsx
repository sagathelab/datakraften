import Layout from '../components/Layout'

export default function Privacy() {
  return (
    <Layout variant="docs" title="Privacy Policy">
      <div className="py-8 text-sm text-text-primary leading-relaxed space-y-4">
        <h1 className="text-3xl font-bold text-magenta font-share-tech mb-4">Privacy Policy</h1>
        <p>Last updated: 2026</p>

        <h2 className="text-lg font-bold text-text-primary mt-6 mb-2">Information We Collect</h2>
        <p>
          Datakraften CLI does not collect or transmit any personal data. It operates entirely on
          your local machine, installing and configuring tools you choose.
        </p>
        <p>
          The website (datakraften.no) is a static site hosted on GitHub Pages and does not use
          cookies, trackers, or analytics of any kind.
        </p>

        <h2 className="text-lg font-bold text-text-primary mt-6 mb-2">Third-Party Services</h2>
        <p>
          When you use Datakraften to install tools, those tools may collect data according to their
          own privacy policies. This includes but is not limited to:
        </p>
        <ul className="list-disc pl-5 space-y-1">
          <li>GitHub CLI -- if you authenticate, GitHub's privacy policy applies</li>
          <li>Azure CLI -- if you authenticate, Microsoft's privacy policy applies</li>
          <li>
            AI tools (Codex, Claude Code, Gemini CLI, OpenCode, GitHub Copilot) -- if you use them,
            their respective privacy policies apply
          </li>
        </ul>

        <h2 className="text-lg font-bold text-text-primary mt-6 mb-2">Data Storage</h2>
        <p>
          Datakraften stores only local configuration files in ~/.config/datakraften/ and state in
          ~/.local/state/datakraften/. No data is sent to external servers by the Datakraften CLI
          itself.
        </p>

        <h2 className="text-lg font-bold text-text-primary mt-6 mb-2">Contact</h2>
        <p>
          For privacy-related questions, open an issue on the{' '}
          <a
            href="https://github.com/sagathelab/datakraften"
            target="_blank"
            rel="noopener noreferrer"
            className="text-magenta hover:underline"
          >
            GitHub repository
          </a>
          .
        </p>
      </div>
    </Layout>
  )
}
