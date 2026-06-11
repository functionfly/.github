import { useEffect } from 'react';
import { Helmet } from 'react-helmet-async';

interface UsePageTitleOptions {
  suffix?: string;
  includeSuffix?: boolean;
}

const DEFAULT_SUFFIX = 'FunctionFly';
const SEPARATOR = ' | ';

/**
 * Hook to set the page title dynamically.
 *
 * @param title - The page title to display
 * @param options.includeSuffix - Whether to append the suffix (default: true)
 * @param options.suffix - Custom suffix to use instead of default
 *
 * @example
 * // Simple usage
 * usePageTitle('Dashboard');
 * // Sets title to "Dashboard | FunctionFly"
 *
 * @example
 * // With custom suffix
 * usePageTitle('Agent Details', { suffix: 'My App' });
 * // Sets title to "Agent Details | My App"
 *
 * @example
 * // Without suffix
 * usePageTitle('Settings', { includeSuffix: false });
 * // Sets title to "Settings"
 */
export function usePageTitle(
  title: string,
  options: UsePageTitleOptions = { includeSuffix: true }
) {
  const { includeSuffix = true, suffix = DEFAULT_SUFFIX } = options;

  const fullTitle = includeSuffix && suffix ? `${title}${SEPARATOR}${suffix}` : title;

  useEffect(() => {
    document.title = fullTitle;
  }, [fullTitle]);

  return (
    <Helmet>
      <title>{fullTitle}</title>
    </Helmet>
  );
}

/**
 * Helper to generate a page title with the default suffix
 */
export function formatPageTitle(title: string, suffix: string = DEFAULT_SUFFIX): string {
  return `${title}${SEPARATOR}${suffix}`;
}
