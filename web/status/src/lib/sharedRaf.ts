let callbacks: Set<(timestamp: number) => void> = new Set();
let rafId: number | null = null;

function loop(timestamp: number) {
  callbacks.forEach((cb) => cb(timestamp));
  if (callbacks.size > 0) {
    rafId = requestAnimationFrame(loop);
  } else {
    rafId = null;
  }
}

export function subscribeSharedRaf(cb: (timestamp: number) => void): () => void {
  callbacks.add(cb);
  if (rafId === null) {
    rafId = requestAnimationFrame(loop);
  }
  return () => {
    callbacks.delete(cb);
  };
}
