import { CodeBlock } from '@/components/common/CodeBlock';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { motion } from 'framer-motion';
import { FileJson, Terminal, Zap } from 'lucide-react';
import type { FunctionInfo } from './types';
import { generateCodeExamples } from './utils';

interface CodeExamplesSectionProps {
  functionInfo: FunctionInfo;
}

export function CodeExamplesSection({ functionInfo }: CodeExamplesSectionProps) {
  const codeExamples = generateCodeExamples(functionInfo);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5, delay: 0.7 }}
      className="function-page-section"
    >
      <Card className="function-page-code-examples">
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Terminal className="w-5 h-5 text-brand-500" />
            Code Examples
          </CardTitle>
          <CardDescription className="text-text-secondary">
            Use these examples to integrate this function into your application
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Tabs defaultValue="curl" className="w-full">
            <TabsList className="grid w-full grid-cols-3 mb-4">
              <TabsTrigger value="curl" className="gap-2">
                <Terminal className="w-4 h-4" />
                cURL
              </TabsTrigger>
              <TabsTrigger value="javascript" className="gap-2">
                <FileJson className="w-4 h-4" />
                JavaScript
              </TabsTrigger>
              <TabsTrigger value="python" className="gap-2">
                <Zap className="w-4 h-4" />
                Python
              </TabsTrigger>
            </TabsList>
            <TabsContent value="curl">
              <CodeBlock code={codeExamples.curl} language="bash" />
            </TabsContent>
            <TabsContent value="javascript">
              <CodeBlock code={codeExamples.javascript} language="javascript" />
            </TabsContent>
            <TabsContent value="python">
              <CodeBlock code={codeExamples.python} language="python" />
            </TabsContent>
          </Tabs>
        </CardContent>
      </Card>
    </motion.div>
  );
}
