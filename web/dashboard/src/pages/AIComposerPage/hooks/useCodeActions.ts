import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';

interface UseCodeActionsOptions {
  getCode: () => string;
  getSelectedLine: () => number | undefined;
  onCodeUpdate: (code: string) => void;
}

export function useCodeActions({ getCode, getSelectedLine, onCodeUpdate }: UseCodeActionsOptions) {
  const explainCodeMutation = useMutation({
    mutationFn: async (params: { code: string; lineNumber?: number }) => {
      return new Promise<string>((resolve) => {
        setTimeout(() => {
          resolve(
            `This function ${params.lineNumber ? `at line ${params.lineNumber}` : ''} processes data by first validating inputs, then performing the core computation, and finally returning a formatted result. It uses error handling to ensure robustness.`
          );
        }, 800);
      });
    },
    onSuccess: (explanation) => {
      toast.success('Explanation ready', { description: explanation });
    },
  });

  const addCommentsMutation = useMutation({
    mutationFn: async (code: string) => {
      return new Promise<string>((resolve) => {
        setTimeout(() => {
          resolve('// Added comprehensive documentation comments\n' + code);
        }, 1000);
      });
    },
    onSuccess: (newCode) => {
      onCodeUpdate(newCode);
      toast.success('Comments added successfully');
    },
  });

  const addLoggingMutation = useMutation({
    mutationFn: async (code: string) => {
      return new Promise<string>((resolve) => {
        setTimeout(() => {
          resolve('// Added structured logging\n' + code);
        }, 1000);
      });
    },
    onSuccess: (newCode) => {
      onCodeUpdate(newCode);
      toast.success('Logging instrumentation added');
    },
  });

  const securityAuditMutation = useMutation({
    mutationFn: async (code: string) => {
      return new Promise<
        Array<{ severity: 'high' | 'medium' | 'low'; issue: string; line?: number }>
      >((resolve) => {
        setTimeout(() => {
          resolve([{ severity: 'low', issue: 'Consider adding input validation', line: 1 }]);
        }, 1200);
      });
    },
    onSuccess: (issues) => {
      if (issues.length === 0) {
        toast.success('No security issues found!');
      } else {
        const highCount = issues.filter((i) => i.severity === 'high').length;
        const mediumCount = issues.filter((i) => i.severity === 'medium').length;
        toast.warning(`Security audit found ${issues.length} issue(s)`, {
          description: `${highCount} high, ${mediumCount} medium priority`,
        });
      }
    },
  });

  const handleExplain = () => {
    explainCodeMutation.mutate({ code: getCode(), lineNumber: getSelectedLine() });
  };

  const handleAddComments = () => {
    addCommentsMutation.mutate(getCode());
  };

  const handleAddLogging = () => {
    addLoggingMutation.mutate(getCode());
  };

  const handleSecurityAudit = () => {
    securityAuditMutation.mutate(getCode());
  };

  return {
    explainPending: explainCodeMutation.isPending,
    commentsPending: addCommentsMutation.isPending,
    loggingPending: addLoggingMutation.isPending,
    securityPending: securityAuditMutation.isPending,
    handleExplain,
    handleAddComments,
    handleAddLogging,
    handleSecurityAudit,
  };
}
