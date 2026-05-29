import {
  Pagination,
  PaginationContent,
  PaginationEllipsis,
  PaginationItem,
  PaginationLink,
  PaginationNext,
  PaginationPrevious,
} from '@/components/ui/pagination';
import { cn } from '@/lib/utils';
import type { MouseEvent } from 'react';
import { getVisiblePages } from '../utils';

interface DiscoveryPaginationProps {
  page: number;
  totalPages: number;
  totalCount: number;
  rangeStart: number;
  rangeEnd: number;
  onPageChange: (page: number) => void;
  className?: string;
}

export function DiscoveryPagination({
  page,
  totalPages,
  totalCount,
  rangeStart,
  rangeEnd,
  onPageChange,
  className,
}: DiscoveryPaginationProps) {
  if (totalPages <= 1) return null;

  const visiblePages = getVisiblePages(page, totalPages);

  const goTo = (next: number) => (e: MouseEvent) => {
    e.preventDefault();
    if (next >= 1 && next <= totalPages && next !== page) {
      onPageChange(next);
    }
  };

  return (
    <div className={cn('discovery-pagination', className)}>
      <p className="discovery-pagination-summary">
        Showing{' '}
        <span className="font-medium text-foreground">
          {rangeStart}–{rangeEnd}
        </span>{' '}
        of <span className="font-medium text-foreground">{totalCount.toLocaleString()}</span>{' '}
        functions
      </p>

      <Pagination>
        <PaginationContent>
          <PaginationItem>
            <PaginationPrevious
              href="#"
              onClick={goTo(page - 1)}
              aria-disabled={page <= 1}
              className={cn(page <= 1 && 'pointer-events-none opacity-40')}
            />
          </PaginationItem>

          {visiblePages.map((p, i) =>
            p === 'ellipsis' ? (
              <PaginationItem key={`ellipsis-${i}`}>
                <PaginationEllipsis />
              </PaginationItem>
            ) : (
              <PaginationItem key={p}>
                <PaginationLink href="#" isActive={p === page} onClick={goTo(p)}>
                  {p}
                </PaginationLink>
              </PaginationItem>
            )
          )}

          <PaginationItem>
            <PaginationNext
              href="#"
              onClick={goTo(page + 1)}
              aria-disabled={page >= totalPages}
              className={cn(page >= totalPages && 'pointer-events-none opacity-40')}
            />
          </PaginationItem>
        </PaginationContent>
      </Pagination>
    </div>
  );
}
