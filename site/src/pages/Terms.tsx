import Layout from '../components/Layout'

export default function Terms() {
  return (
    <Layout variant="docs" title="Terms of Use">
      <div className="py-8 text-base text-text-primary leading-relaxed space-y-4">
        <h1 className="text-3xl font-bold text-magenta font-share-tech mb-4">Terms of Use</h1>
        <p>Last updated: 2026</p>

        <h2 className="text-xl font-bold text-text-primary mt-6 mb-2">License</h2>
        <p>
          Datakraften is open source software released under the Apache License 2.0. You are free to
          use, modify, and distribute it in accordance with the license terms.
        </p>

        <h2 className="text-xl font-bold text-text-primary mt-6 mb-2">Disclaimer</h2>
        <p>
          Datakraften is provided "as is", without warranty of any kind, express or implied. The
          tools it installs are subject to their own licenses and terms.
        </p>

        <h2 className="text-xl font-bold text-text-primary mt-6 mb-2">Third-Party Tools</h2>
        <p>
          Datakraften orchestrates the installation of third-party tools. Each tool has its own
          license, terms, and privacy policy. By using a tool installed via Datakraften, you agree
          to that tool's terms.
        </p>

        <h2 className="text-xl font-bold text-text-primary mt-6 mb-2">Limitation of Liability</h2>
        <p>
          In no event shall the Datakraften authors be liable for any claim, damages, or other
          liability arising from the use of this software or the tools it installs.
        </p>

        <h2 className="text-xl font-bold text-text-primary mt-6 mb-2">Contact</h2>
        <p>
          For questions about these terms, open an issue on the{' '}
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
