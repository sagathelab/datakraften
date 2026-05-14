import Layout from './Layout'

interface ToolSection {
  title?: string
  body: string
}

interface ToolGuideProps {
  title: string
  subtitle: string
  sections: ToolSection[]
  website: string
}

export default function ToolGuide({ title, subtitle, sections, website }: ToolGuideProps) {
  return (
    <Layout variant="docs" title={title}>
      <div className="doc-content py-8">
        <h1 className="text-xl font-bold text-magenta font-share-tech mb-2">{title}</h1>
        <p className="text-sm text-text-dim mb-8">{subtitle}</p>

        {sections.map((section, i) => (
          <div key={i} className="mb-6">
            {section.title && (
              <h2 className="text-base font-bold text-text-primary mb-2">{section.title}</h2>
            )}
            <div className="text-xs text-text-primary leading-relaxed space-y-2">
              {section.body.split('\n').map((line, j) => {
                if (line.startsWith('$ ')) {
                  return (
                    <div key={j} className="font-jetbrains">
                      <span className="text-prompt">$</span>
                      <span className="ml-2">{line.slice(2)}</span>
                    </div>
                  )
                }
                if (line.startsWith('# ')) {
                  return (
                    <div key={j} className="font-jetbrains text-comment">{line}</div>
                  )
                }
                if (line.startsWith('// ') || line.startsWith('<!-- ')) {
                  return (
                    <div key={j} className="font-jetbrains text-comment">{line}</div>
                  )
                }
                if (line.startsWith('> ')) {
                  return (
                    <div key={j} className="font-jetbrains text-output">{line.slice(2)}</div>
                  )
                }
                if (line.startsWith('TIP:')) {
                  return (
                    <div key={j} className="bg-fuchsia-500/5 border border-fuchsia-500/20 rounded p-3 text-text-dim">
                      <span className="text-magenta">&#9654;</span> {line.slice(5)}
                    </div>
                  )
                }
                return <p key={j}>{line}</p>
              })}
            </div>
          </div>
        ))}

        <div className="mt-8 pt-4 border-t border-fuchsia-500/20">
          <a
            href={website}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs text-magenta hover:underline"
          >
            Official site &rarr;
          </a>
        </div>
      </div>
    </Layout>
  )
}
