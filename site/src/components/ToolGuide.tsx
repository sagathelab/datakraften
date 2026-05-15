import { Fragment, useState, useCallback } from 'react'
import { Link } from 'react-router-dom'
import Layout from './Layout'

const dkLinks: Record<string, string> = {
  init: '/docs/dk#init',
  apply: '/docs/dk#apply',
  doctor: '/docs/dk#doctor',
  status: '/docs/dk#status',
  update: '/docs/dk#update',
  profile: '/docs/dk#profile',
}

interface ToolSection {
  title?: string
  body: string
  id?: string
}

interface ToolGuideProps {
  title: string
  subtitle: string
  sections: ToolSection[]
  website?: string
}

function renderInlineCode(text: string) {
  const parts = text.split(/(`[^`]+`|~[^~]+~)/g)
  return parts.map((part, i) => {
    if (part.startsWith('`') && part.endsWith('`')) {
      const code = part.slice(1, -1)
      const dkMatch = code.match(/^dk (\w+)/)
      if (dkMatch && dkLinks[dkMatch[1]]) {
        return (
          <Link key={i} to={dkLinks[dkMatch[1]]} className="inline-code inline-code--link">
            {code}
          </Link>
        )
      }
      return <code key={i} className="inline-code">{code}</code>
    }
    if (part.startsWith('~') && part.endsWith('~')) {
      return <mark key={i} className="inline-highlight">{part.slice(1, -1)}</mark>
    }
    return <span key={i}>{part}</span>
  })
}

function CodeBlock({ lines }: { lines: string[] }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = useCallback(() => {
    const text = lines.map(l => {
      if (l.startsWith('$ ')) return l.slice(2)
      if (l.startsWith('> ')) return l.slice(2)
      if (l.startsWith('// ')) return l.slice(3)
      if (l.startsWith('# ')) return l.slice(2)
      return l
    }).join('\n')
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }, [lines])

  return (
    <div className="code-block">
      <button
        onClick={handleCopy}
        className={`yaml-copy ${copied ? 'yaml-copy--copied' : ''}`}
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
      {lines.filter(l => l.trim() !== '').map((line, li) => {
        if (line.startsWith('$ ') || line === '$') {
          return (
            <div key={li}>
              <span className="code-prompt">$</span>
              <span className="ml-2">{renderInlineCode(line.slice(2))}</span>
            </div>
          )
        }
        if (line.startsWith('> ')) {
          return (
            <div key={li} className="code-output">{renderInlineCode(line.slice(2))}</div>
          )
        }
        if (line.startsWith('// ') || line.startsWith('# ') || line.startsWith('<!-- ')) {
          const comment = line.replace(/^\/\/ |^# |^<!-- /, '')
          return (
            <div key={li} className="code-comment">{renderInlineCode(comment)}</div>
          )
        }
        return <div key={li}>{renderInlineCode(line)}</div>
      })}
    </div>
  )
}

function YamlBlock({ lines }: { lines: string[] }) {
  const [copied, setCopied] = useState(false)

  const handleCopy = useCallback(() => {
    const text = lines.map(l => l.replace(/^\| /, '')).join('\n')
    navigator.clipboard.writeText(text).then(() => {
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    })
  }, [lines])

  return (
    <div className="yaml-block">
      <button
        onClick={handleCopy}
        className={`yaml-copy ${copied ? 'yaml-copy--copied' : ''}`}
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
      {lines.filter(l => l.trim() !== '').map((line, li) => {
        const content = line.replace(/^\| /, '').replace(/^\|$/, '')
        return (
          <div key={li} className="yaml-line">
            {renderInlineCode(content)}
          </div>
        )
      })}
    </div>
  )
}

function renderBody(body: string) {
  const lines = body.split('\n')
  const blocks: { type: 'code' | 'text' | 'tip' | 'note' | 'list' | 'yaml'; lines: string[] }[] = []
  let current: { type: 'code' | 'text' | 'tip' | 'note' | 'list' | 'yaml'; lines: string[] } | null = null

  const codePrefixes = ['$ ', '# ', '// ', '<!-- ', '> ']
  const isCode = (l: string) => codePrefixes.some(p => l.startsWith(p)) || l === '$'
  const isTip = (l: string) => l.startsWith('TIP:')
  const isNote = (l: string) => l.startsWith('NOTE:')
  const isEmpty = (l: string) => l.trim() === ''
  const isListItem = (l: string) => l.startsWith('- ')
  const isYaml = (l: string) => l.startsWith('| ') || l === '|'

  for (const line of lines) {
    if (isTip(line)) {
      if (current && current.type !== 'tip') { blocks.push(current); current = null }
      if (!current) current = { type: 'tip', lines: [] }
      current.lines.push(line.slice(5).trim())
    } else if (isNote(line)) {
      if (current && current.type !== 'note') { blocks.push(current); current = null }
      if (!current) current = { type: 'note', lines: [] }
      current.lines.push(line.slice(5).trim())
    } else if (isYaml(line)) {
      if (current && current.type !== 'yaml') { blocks.push(current); current = null }
      if (!current) current = { type: 'yaml', lines: [] }
      current.lines.push(line)
    } else if (isCode(line)) {
      if (current && current.type !== 'code') { blocks.push(current); current = null }
      if (!current) current = { type: 'code', lines: [] }
      current.lines.push(line)
    } else if (isListItem(line)) {
      if (current && current.type !== 'list') { blocks.push(current); current = null }
      if (!current) current = { type: 'list', lines: [] }
      current.lines.push(line.slice(2))
    } else if (isEmpty(line)) {
      if (current && (current.type === 'code' || current.type === 'yaml')) {
        current.lines.push(line)
      } else {
        if (current) { blocks.push(current); current = null }
      }
    } else {
      if (current && current.type !== 'text') { blocks.push(current); current = null }
      if (!current) current = { type: 'text', lines: [] }
      current.lines.push(line)
    }
  }
  if (current) blocks.push(current)

  return blocks.map((block, bi) => {
    if (block.type === 'yaml') {
      return <YamlBlock key={bi} lines={block.lines} />
    }

    if (block.type === 'code') {
      return <CodeBlock key={bi} lines={block.lines} />
    }

    if (block.type === 'tip') {
      return (
        <div key={bi} className="tip-box">
          <span className="text-magenta font-bold">&#9654;</span>
          {' '}
          {block.lines.map((l, li) => (
            <Fragment key={li}>
              {li > 0 && <br />}
              {renderInlineCode(l)}
            </Fragment>
          ))}
        </div>
      )
    }

    if (block.type === 'note') {
      return (
        <div key={bi} className="note-box">
          <span className="text-text-dim font-bold">~</span>
          {' '}
          {block.lines.map((l, li) => (
            <Fragment key={li}>
              {li > 0 && <br />}
              {renderInlineCode(l)}
            </Fragment>
          ))}
        </div>
      )
    }

    if (block.type === 'list') {
      return (
        <ul key={bi} className="list-disc list-inside text-sm text-text-body leading-relaxed mb-3 space-y-1">
          {block.lines.map((l, li) => (
            <li key={li}>{renderInlineCode(l)}</li>
          ))}
        </ul>
      )
    }

    return (
      <p key={bi} className="text-sm text-text-body leading-relaxed mb-3">
        {renderInlineCode(block.lines.join(' '))}
      </p>
    )
  })
}

export default function ToolGuide({ title, subtitle, sections, website }: ToolGuideProps) {
  const toc = sections.filter(s => s.title)

  return (
    <Layout variant="docs" title={title}>
      <div className="flex gap-10 py-8">
        <div className="doc-content flex-1 min-w-0">
          <h1 className="text-3xl font-bold text-magenta font-share-tech mb-2">{title}</h1>
          <p className="text-base text-text-dim mb-8">{subtitle}</p>

          {sections.map((section, i) => (
            <div key={i} className="mb-6">
              {section.title && (
                <h2 id={section.id || section.title.toLowerCase().replace(/\s+/g, '-')} className="text-lg font-bold text-text-primary mb-2 scroll-mt-16">{section.title}</h2>
              )}
              {renderBody(section.body)}
            </div>
          ))}

          {website && (
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
          )}
        </div>

        {toc.length > 1 && (
          <aside className="hidden lg:block w-56 flex-shrink-0">
            <nav className="sticky top-20 space-y-1.5">
              <span className="text-xs text-text-dim uppercase tracking-wider font-semibold">On this page</span>
              {toc.map((s, i) => (
                <a
                  key={i}
                  href={`#${s.id || s.title!.toLowerCase().replace(/\s+/g, '-')}`}
                  className="block text-sm text-text-dim hover:text-magenta transition-colors leading-relaxed"
                >
                  {s.title}
                </a>
              ))}
            </nav>
          </aside>
        )}
      </div>
    </Layout>
  )
}
