/**
 * Diff algorithm utilities using diff-match-patch
 */

import { diff_match_patch, DIFF_DELETE, DIFF_INSERT, DIFF_EQUAL } from "diff-match-patch";
import type { WordDiff, DiffHunk, DiffStats, DiffError } from "../types";

/**
 * Compute word-level diffs for a modified line
 */
export function computeWordDiff(oldText: string, newText: string): WordDiff[] {
  const dmp = new diff_match_patch();
  const diffs = dmp.diff_main(oldText, newText);

  // Clean up the diffs for better readability
  dmp.diff_cleanupSemantic(diffs);
  dmp.diff_cleanupEfficiency(diffs);

  const wordDiffs: WordDiff[] = [];

  diffs.forEach(([op, text]) => {
    const type = op === DIFF_DELETE ? "removed" : op === DIFF_INSERT ? "added" : "unchanged";
    if (text) {
      wordDiffs.push({ text, type });
    }
  });

  return wordDiffs;
}

/**
 * Compute the diff between two strings using diff-match-patch algorithm
 */
export function computeDiff(
  before: string,
  after: string,
  options: {
    enableWordDiffs?: boolean;
    timeout?: number;
    editCost?: number;
  } = {}
): { hunks: DiffHunk[]; stats: DiffStats; error?: DiffError } {
  try {
    const dmp = new diff_match_patch();

    // Configure the diff algorithm
    if (options.timeout !== undefined) {
      dmp.Diff_Timeout = options.timeout / 1000; // Convert to seconds
    }
    if (options.editCost !== undefined) {
      dmp.Diff_EditCost = options.editCost;
    }

    // Compute line-based diff first
    const lineDiffs = dmp.diff_main(before, after);
    dmp.diff_cleanupSemantic(lineDiffs);

    const beforeLines = before.split("\n");
    const afterLines = after.split("\n");

    const hunks: DiffHunk[] = [];
    const allLines: { content: string; type: "added" | "removed" | "unchanged"; oldLineNum?: number; newLineNum?: number }[] = [];

    let oldLineNum = 1;
    let newLineNum = 1;
    let currentHunkLines: typeof allLines = [];
    let inHunk = false;

    // Process each diff segment
    lineDiffs.forEach(([op, text]) => {
      const lines = text.split("\n");

      lines.forEach((line, index) => {
        // Skip empty lines at segment boundaries (except for the last segment)
        if (line === "" && index < lines.length - 1) return;

        const isLastLine = index === lines.length - 1;
        const lineType = op === DIFF_DELETE ? "removed" : op === DIFF_INSERT ? "added" : "unchanged";

        if (lineType !== "unchanged") {
          if (!inHunk) {
            inHunk = true;
            currentHunkLines = [];
          }

          const lineData: { content: string; type: "added" | "removed"; oldLineNum?: number; newLineNum?: number } = {
            content: line,
            type: lineType as "added" | "removed",
          };

          if (lineType === "removed") {
            lineData.oldLineNum = oldLineNum++;
          } else if (lineType === "added") {
            lineData.newLineNum = newLineNum++;
          }

          currentHunkLines.push(lineData);
          allLines.push(lineData);
        } else {
          if (inHunk) {
            // End current hunk
            if (currentHunkLines.length > 0) {
              hunks.push({
                oldRange: { start: Math.max(1, oldLineNum - currentHunkLines.length), count: currentHunkLines.length },
                newRange: { start: Math.max(1, newLineNum - currentHunkLines.length), count: currentHunkLines.length },
                lines: currentHunkLines.map(line => ({
                  content: line.content,
                  type: line.type,
                  oldLineNumber: (line as any).oldLineNum,
                  newLineNumber: (line as any).newLineNum,
                  isInChangeBlock: true,
                })),
              });
            }
            inHunk = false;
            currentHunkLines = [];
          }

          // Add unchanged line
          allLines.push({
            content: line,
            type: "unchanged",
            oldLineNum: oldLineNum++,
            newLineNum: newLineNum++,
          });
        }
      });
    });

    // Handle any remaining hunk
    if (inHunk && currentHunkLines.length > 0) {
      hunks.push({
        oldRange: { start: Math.max(1, oldLineNum - currentHunkLines.length), count: currentHunkLines.length },
        newRange: { start: Math.max(1, newLineNum - currentHunkLines.length), count: currentHunkLines.length },
        lines: currentHunkLines.map(line => ({
          content: line.content,
          type: line.type,
          oldLineNumber: (line as any).oldLineNum,
          newLineNumber: (line as any).newLineNum,
          isInChangeBlock: true,
        })),
      });
    }

    // Calculate statistics
    const stats: DiffStats = {
      added: 0,
      removed: 0,
      modified: 0,
      total: hunks.length,
      totalLines: allLines.length,
    };

    hunks.forEach((hunk) => {
      hunk.lines.forEach((line) => {
        if (line.type === "added") stats.added++;
        else if (line.type === "removed") stats.removed++;
        else if (line.type === "modified") stats.modified++;
      });
    });

    return { hunks, stats };
  } catch (error) {
    const diffError: DiffError = {
      message: error instanceof Error ? error.message : "Unknown diff computation error",
      code: "DIFF_COMPUTATION_FAILED",
      recoverable: true,
    };
    return { hunks: [], stats: { added: 0, removed: 0, modified: 0, total: 0, totalLines: 0 }, error: diffError };
  }
}