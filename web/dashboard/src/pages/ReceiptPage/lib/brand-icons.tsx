// Brand icons — inline SVG components for social/brand marks that aren't
// in lucide-react. Paths are sourced from the simple-icons package
// (already a dep) which is CC0-licensed. Adding brand icons inline
// avoids pulling in another icon package just for two logos.
//
// Only include icons that we actually use. Add more as needed.

import type { SVGProps } from "react";

type IconProps = SVGProps<SVGSVGElement> & { size?: number; title?: string };

/**
 * X (formerly Twitter) brand icon. ViewBox 0 0 24 24.
 * Source: simple-icons/icons/x.svg
 */
export function XBrandIcon({ size = 16, title = "X", ...rest }: IconProps) {
  return (
    <svg
      role="img"
      viewBox="0 0 24 24"
      width={size}
      height={size}
      aria-label={title}
      xmlns="http://www.w3.org/2000/svg"
      {...rest}
    >
      <title>{title}</title>
      <path
        fill="currentColor"
        d="M14.234 10.162 22.977 0h-2.072l-7.591 8.824L7.251 0H.258l9.168 13.343L.258 24H2.33l8.016-9.318L16.749 24h6.993zm-2.837 3.299-.929-1.329L3.076 1.56h3.182l5.965 8.532.929 1.329 7.754 11.09h-3.182z"
      />
    </svg>
  );
}
