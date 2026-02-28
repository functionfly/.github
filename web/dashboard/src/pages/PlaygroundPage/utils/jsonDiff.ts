/**
 * JSON diff utilities for comparing two JSON values.
 * Returns a structured diff that can be rendered in the DiffViewer.
 */

export type DiffType = 'added' | 'removed' | 'changed' | 'unchanged';

export interface DiffNode {
  key: string;
  path: string;
  type: DiffType;
  leftValue?: unknown;
  rightValue?: unknown;
  children?: DiffNode[];
}

function getType(value: unknown): string {
  if (value === null) return 'null';
  if (Array.isArray(value)) return 'array';
  return typeof value;
}

function isPrimitive(value: unknown): boolean {
  const t = getType(value);
  return t !== 'object' && t !== 'array';
}

export function diffJson(left: unknown, right: unknown, path = ''): DiffNode[] {
  const nodes: DiffNode[] = [];

  const leftType = getType(left);
  const rightType = getType(right);

  // If types differ, it's a change at this level
  if (leftType !== rightType) {
    return [
      {
        key: path.split('.').pop() || 'root',
        path,
        type: 'changed',
        leftValue: left,
        rightValue: right,
      },
    ];
  }

  if (leftType === 'object' && left !== null && right !== null) {
    const leftObj = left as Record<string, unknown>;
    const rightObj = right as Record<string, unknown>;
    const allKeys = new Set([...Object.keys(leftObj), ...Object.keys(rightObj)]);

    for (const key of allKeys) {
      const childPath = path ? `${path}.${key}` : key;
      const hasLeft = key in leftObj;
      const hasRight = key in rightObj;

      if (!hasLeft) {
        nodes.push({
          key,
          path: childPath,
          type: 'added',
          rightValue: rightObj[key],
        });
      } else if (!hasRight) {
        nodes.push({
          key,
          path: childPath,
          type: 'removed',
          leftValue: leftObj[key],
        });
      } else {
        const childDiff = diffJson(leftObj[key], rightObj[key], childPath);
        if (childDiff.length === 1 && childDiff[0].path === childPath) {
          nodes.push(childDiff[0]);
        } else if (childDiff.length > 0) {
          const hasChanges = childDiff.some((n) => n.type !== 'unchanged');
          nodes.push({
            key,
            path: childPath,
            type: hasChanges ? 'changed' : 'unchanged',
            leftValue: leftObj[key],
            rightValue: rightObj[key],
            children: childDiff,
          });
        } else {
          nodes.push({
            key,
            path: childPath,
            type: 'unchanged',
            leftValue: leftObj[key],
            rightValue: rightObj[key],
          });
        }
      }
    }
  } else if (leftType === 'array') {
    const leftArr = left as unknown[];
    const rightArr = right as unknown[];
    const maxLen = Math.max(leftArr.length, rightArr.length);

    for (let i = 0; i < maxLen; i++) {
      const childPath = `${path}[${i}]`;
      if (i >= leftArr.length) {
        nodes.push({
          key: `[${i}]`,
          path: childPath,
          type: 'added',
          rightValue: rightArr[i],
        });
      } else if (i >= rightArr.length) {
        nodes.push({
          key: `[${i}]`,
          path: childPath,
          type: 'removed',
          leftValue: leftArr[i],
        });
      } else {
        const childDiff = diffJson(leftArr[i], rightArr[i], childPath);
        if (childDiff.length === 1 && childDiff[0].path === childPath) {
          nodes.push(childDiff[0]);
        } else if (childDiff.length > 0) {
          const hasChanges = childDiff.some((n) => n.type !== 'unchanged');
          nodes.push({
            key: `[${i}]`,
            path: childPath,
            type: hasChanges ? 'changed' : 'unchanged',
            leftValue: leftArr[i],
            rightValue: rightArr[i],
            children: childDiff,
          });
        } else {
          nodes.push({
            key: `[${i}]`,
            path: childPath,
            type: 'unchanged',
            leftValue: leftArr[i],
            rightValue: rightArr[i],
          });
        }
      }
    }
  } else {
    // Primitive comparison
    if (isPrimitive(left) && isPrimitive(right)) {
      const type: DiffType = left === right ? 'unchanged' : 'changed';
      return [
        {
          key: path.split('.').pop() || 'root',
          path,
          type,
          leftValue: left,
          rightValue: right,
        },
      ];
    }
  }

  return nodes;
}

export function countDiffChanges(nodes: DiffNode[]): { added: number; removed: number; changed: number } {
  let added = 0;
  let removed = 0;
  let changed = 0;

  function traverse(node: DiffNode) {
    if (node.type === 'added') added++;
    else if (node.type === 'removed') removed++;
    else if (node.type === 'changed') changed++;
    if (node.children) node.children.forEach(traverse);
  }

  nodes.forEach(traverse);
  return { added, removed, changed };
}
