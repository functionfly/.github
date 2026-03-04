/**
 * ThreadReplyComposer - Reply input with code editor
 */

import { useState, useCallback } from 'react';
import { cn } from '@/lib/utils';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  Code,
  Paperclip,
  Send,
  Bold,
  Italic,
  Link,
  List,
  X,
  FileCode,
} from 'lucide-react';

interface CodeBlock {
  id: string;
  language: string;
  code: string;
  filename?: string;
}

interface ThreadReplyComposerProps {
  threadId: string;
  onSubmit: (content: string, codeBlocks: CodeBlock[]) => Promise<void>;
  className?: string;
}

const languages = [
  { value: 'python', label: 'Python' },
  { value: 'javascript', label: 'JavaScript' },
  { value: 'typescript', label: 'TypeScript' },
  { value: 'rust', label: 'Rust' },
  { value: 'go', label: 'Go' },
  { value: 'java', label: 'Java' },
  { value: 'cpp', label: 'C++' },
  { value: 'csharp', label: 'C#' },
  { value: 'ruby', label: 'Ruby' },
  { value: 'php', label: 'PHP' },
];

export function ThreadReplyComposer({
  threadId,
  onSubmit,
  className,
}: ThreadReplyComposerProps) {
  const [content, setContent] = useState('');
  const [codeBlocks, setCodeBlocks] = useState<CodeBlock[]>([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [showCodeEditor, setShowCodeEditor] = useState(false);
  const [currentCode, setCurrentCode] = useState('');
  const [currentLanguage, setCurrentLanguage] = useState('python');
  const [currentFilename, setCurrentFilename] = useState('');

  const handleSubmit = async () => {
    if (!content.trim() && codeBlocks.length === 0) return;

    setIsSubmitting(true);
    try {
      await onSubmit(content, codeBlocks);
      setContent('');
      setCodeBlocks([]);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleAddCodeBlock = () => {
    if (!currentCode.trim()) return;

    const newBlock: CodeBlock = {
      id: Date.now().toString(),
      language: currentLanguage,
      code: currentCode,
      filename: currentFilename || undefined,
    };

    setCodeBlocks((prev) => [...prev, newBlock]);
    setCurrentCode('');
    setCurrentFilename('');
    setShowCodeEditor(false);
  };

  const handleRemoveCodeBlock = (id: string) => {
    setCodeBlocks((prev) => prev.filter((block) => block.id !== id));
  };

  const insertMarkdown = useCallback((before: string, after: string = '') => {
    const textarea = document.getElementById('reply-textarea') as HTMLTextAreaElement;
    if (!textarea) return;

    const start = textarea.selectionStart;
    const end = textarea.selectionEnd;
    const selectedText = content.substring(start, end);
    const newText = content.substring(0, start) + before + selectedText + after + content.substring(end);

    setContent(newText);

    // Restore focus and selection
    setTimeout(() => {
      textarea.focus();
      textarea.setSelectionRange(start + before.length, end + before.length);
    }, 0);
  }, [content]);

  return (
    <div className={cn('space-y-4', className)}>
      {/* Existing Code Blocks */}
      {codeBlocks.length > 0 && (
        <div className="space-y-2">
          {codeBlocks.map((block) => (
            <div
              key={block.id}
              className="rounded-lg border border-slate-800 bg-slate-900/50"
            >
              <div className="flex items-center justify-between border-b border-slate-800 px-3 py-2">
                <div className="flex items-center gap-2">
                  <FileCode className="h-4 w-4 text-slate-500" />
                  <span className="text-sm text-slate-400">
                    {block.filename || `${block.language}.${getExtension(block.language)}`}
                  </span>
                </div>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 text-slate-500 hover:text-slate-300"
                  onClick={() => handleRemoveCodeBlock(block.id)}
                >
                  <X className="h-4 w-4" />
                </Button>
              </div>
              <pre className="max-h-48 overflow-auto p-3 text-sm text-slate-300">
                <code>{block.code.slice(0, 500)}{block.code.length > 500 && '...'}</code>
              </pre>
            </div>
          ))}
        </div>
      )}

      {/* Toolbar */}
      <div className="flex flex-wrap items-center gap-1 border-b border-slate-800 pb-2">
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0 text-slate-400 hover:text-slate-200"
          onClick={() => insertMarkdown('**', '**')}
        >
          <Bold className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0 text-slate-400 hover:text-slate-200"
          onClick={() => insertMarkdown('*', '*')}
        >
          <Italic className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0 text-slate-400 hover:text-slate-200"
          onClick={() => insertMarkdown('[', '](url)')}
        >
          <Link className="h-4 w-4" />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 w-8 p-0 text-slate-400 hover:text-slate-200"
          onClick={() => insertMarkdown('\n- ')}
        >
          <List className="h-4 w-4" />
        </Button>
        <div className="mx-2 h-4 w-px bg-slate-800" />
        <Button
          variant="ghost"
          size="sm"
          className={cn(
            'h-8 gap-1.5 text-slate-400 hover:text-slate-200',
            showCodeEditor && 'bg-slate-800 text-slate-200'
          )}
          onClick={() => setShowCodeEditor(!showCodeEditor)}
        >
          <Code className="h-4 w-4" />
          Add Code
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-8 gap-1.5 text-slate-400 hover:text-slate-200"
        >
          <Paperclip className="h-4 w-4" />
          Attach
        </Button>
      </div>

      {/* Code Editor */}
      {showCodeEditor && (
        <div className="rounded-lg border border-slate-800 bg-slate-900/50 p-4 space-y-3">
          <div className="flex flex-wrap gap-2">
            <Select value={currentLanguage} onValueChange={setCurrentLanguage}>
              <SelectTrigger className="w-40 border-slate-700 bg-slate-800 text-slate-200">
                <SelectValue />
              </SelectTrigger>
              <SelectContent className="bg-slate-900 border-slate-800">
                {languages.map((lang) => (
                  <SelectItem
                    key={lang.value}
                    value={lang.value}
                    className="text-slate-200 focus:bg-slate-800 focus:text-white"
                  >
                    {lang.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <input
              type="text"
              placeholder="Filename (optional)"
              value={currentFilename}
              onChange={(e) => setCurrentFilename(e.target.value)}
              className="flex-1 min-w-[150px] rounded-md border border-slate-700 bg-slate-800 px-3 py-1.5 text-sm text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-indigo-500"
            />
          </div>
          <Textarea
            value={currentCode}
            onChange={(e) => setCurrentCode(e.target.value)}
            placeholder="Paste your code here..."
            className="min-h-[150px] border-slate-700 bg-slate-950 font-mono text-sm text-slate-200 placeholder:text-slate-600 focus-visible:ring-indigo-500"
          />
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowCodeEditor(false)}
              className="text-slate-400 hover:text-slate-200"
            >
              Cancel
            </Button>
            <Button
              size="sm"
              onClick={handleAddCodeBlock}
              disabled={!currentCode.trim()}
              className="bg-indigo-600 hover:bg-indigo-500"
            >
              <Code className="mr-2 h-4 w-4" />
              Add Code Block
            </Button>
          </div>
        </div>
      )}

      {/* Main Textarea */}
      <Textarea
        id="reply-textarea"
        value={content}
        onChange={(e) => setContent(e.target.value)}
        placeholder="Write your reply... Use markdown for formatting"
        className="min-h-[150px] border-slate-800 bg-slate-900 text-slate-200 placeholder:text-slate-500 focus-visible:ring-indigo-500"
      />

      {/* Submit Button */}
      <div className="flex justify-end">
        <Button
          onClick={handleSubmit}
          disabled={(!content.trim() && codeBlocks.length === 0) || isSubmitting}
          className="bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50"
        >
          {isSubmitting ? (
            <>
              <div className="mr-2 h-4 w-4 animate-spin rounded-full border-2 border-white/30 border-t-white" />
              Posting...
            </>
          ) : (
            <>
              <Send className="mr-2 h-4 w-4" />
              Post Reply
            </>
          )}
        </Button>
      </div>
    </div>
  );
}

function getExtension(language: string): string {
  const extensions: Record<string, string> = {
    python: 'py',
    javascript: 'js',
    typescript: 'ts',
    rust: 'rs',
    go: 'go',
    java: 'java',
    cpp: 'cpp',
    csharp: 'cs',
    ruby: 'rb',
    php: 'php',
  };
  return extensions[language] || language;
}
