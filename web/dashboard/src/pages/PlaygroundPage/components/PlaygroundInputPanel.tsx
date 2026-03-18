import { ManifestInputForm } from '@/components/common/ManifestInputForm';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { cn } from '@/lib/utils';
import { useThemeStore } from '@/stores/themeStore';
import Editor from '@monaco-editor/react';
import { AnimatePresence, motion } from 'framer-motion';
import { BookOpen, FileCode2, FormInput } from 'lucide-react';
import { useCallback } from 'react';
import { InputTab, usePlaygroundStore } from '../store/playgroundStore';
import { ExampleSelector } from './ExampleSelector';

interface PlaygroundInputPanelProps {
  className?: string;
}

export function PlaygroundInputPanel({ className }: PlaygroundInputPanelProps) {
  const {
    functionInfo,
    inputValue,
    inputJson,
    activeInputTab,
    setActiveInputTab,
    setInputValue,
    setInputJson,
  } = usePlaygroundStore();

  const { resolvedTheme } = useThemeStore();

  const handleFormChange = useCallback(
    (value: unknown) => {
      setInputValue(value);
    },
    [setInputValue]
  );

  const handleJsonChange = useCallback(
    (value: string | undefined) => {
      if (value !== undefined) {
        setInputJson(value);
      }
    },
    [setInputJson]
  );

  const inputSpec = functionInfo?.manifest?.input;

  // Build Monaco JSON schema for validation
  const monacoSchema = inputSpec?.schema
    ? {
        uri: 'http://functionfly/input-schema.json',
        fileMatch: ['*'],
        schema: inputSpec.schema,
      }
    : undefined;

  return (
    <div className={cn('flex flex-col h-full', className)}>
      <Tabs
        value={activeInputTab}
        onValueChange={(v) => setActiveInputTab(v as InputTab)}
        className="flex flex-col h-full"
      >
        {/* Tab bar */}
        <div className="flex items-center border-b border-border-subtle px-3 pt-1 bg-bg-secondary shrink-0">
          <TabsList className="h-8 bg-transparent gap-0 p-0">
            <TabsTrigger
              value="form"
              className="h-8 px-3 text-xs rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:text-indigo-400 data-[state=active]:bg-transparent gap-1.5"
            >
              <FormInput className="w-3.5 h-3.5" />
              Form
            </TabsTrigger>
            <TabsTrigger
              value="json"
              className="h-8 px-3 text-xs rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:text-indigo-400 data-[state=active]:bg-transparent gap-1.5"
            >
              <FileCode2 className="w-3.5 h-3.5" />
              JSON
            </TabsTrigger>
            <TabsTrigger
              value="examples"
              className="h-8 px-3 text-xs rounded-none border-b-2 border-transparent data-[state=active]:border-indigo-500 data-[state=active]:text-indigo-400 data-[state=active]:bg-transparent gap-1.5"
            >
              <BookOpen className="w-3.5 h-3.5" />
              Examples
            </TabsTrigger>
          </TabsList>

          <div className="ml-auto text-xs text-text-muted pr-1">Input</div>
        </div>

        {/* Tab content: single keyed child for AnimatePresence mode="wait" */}
        <div className="flex-1 overflow-hidden">
          <TabsContent
            value="form"
            className="h-full m-0 overflow-auto"
            forceMount
            hidden={activeInputTab !== 'form'}
          >
            <AnimatePresence mode="wait">
              {activeInputTab === 'form' && (
                <motion.div
                  key="form"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.15 }}
                  className="p-3 h-full"
                >
                  {inputSpec ? (
                    <ManifestInputForm
                      inputSpec={inputSpec}
                      value={inputValue}
                      onChange={handleFormChange}
                      className="border-0 shadow-none bg-transparent"
                    />
                  ) : (
                    <div className="flex flex-col items-center justify-center h-full text-center py-8">
                      <FormInput className="w-10 h-10 text-text-muted mb-3" />
                      <p className="text-sm text-text-muted">No input schema defined</p>
                      <p className="text-xs text-text-muted mt-1">
                        Switch to JSON tab to enter raw input
                      </p>
                    </div>
                  )}
                </motion.div>
              )}
            </AnimatePresence>
          </TabsContent>

          <TabsContent
            value="json"
            className="h-full m-0"
            forceMount
            hidden={activeInputTab !== 'json'}
          >
            <AnimatePresence mode="wait">
              {activeInputTab === 'json' && (
                <motion.div
                  key="json"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.15 }}
                  className="h-full"
                >
                  <Editor
                    height="100%"
                    language="json"
                    value={inputJson}
                    onChange={handleJsonChange}
                    theme={resolvedTheme === 'dark' ? 'vs-dark' : 'light'}
                    options={{
                      minimap: { enabled: false },
                      fontSize: 13,
                      lineNumbers: 'on',
                      scrollBeyondLastLine: false,
                      wordWrap: 'on',
                      formatOnPaste: true,
                      formatOnType: true,
                      tabSize: 2,
                      automaticLayout: true,
                      padding: { top: 8, bottom: 8 },
                      scrollbar: {
                        verticalScrollbarSize: 6,
                        horizontalScrollbarSize: 6,
                      },
                    }}
                    beforeMount={(monaco) => {
                      if (monacoSchema) {
                        monaco.languages.json.jsonDefaults.setDiagnosticsOptions({
                          validate: true,
                          schemas: [monacoSchema],
                        });
                      }
                    }}
                  />
                </motion.div>
              )}
            </AnimatePresence>
          </TabsContent>

          <TabsContent
            value="examples"
            className="h-full m-0 overflow-auto"
            forceMount
            hidden={activeInputTab !== 'examples'}
          >
            <AnimatePresence mode="wait">
              {activeInputTab === 'examples' && (
                <motion.div
                  key="examples"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1 }}
                  exit={{ opacity: 0 }}
                  transition={{ duration: 0.15 }}
                  className="p-3"
                >
                  <ExampleSelector />
                </motion.div>
              )}
            </AnimatePresence>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  );
}
