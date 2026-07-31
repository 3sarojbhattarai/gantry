import { useEffect, useState } from "react";
import type { Stats } from "@/lib/types";

// useLogStream tails a container's logs over SSE, keeping the last 1000 lines.
export function useLogStream(id: string | null): string[] {
  const [lines, setLines] = useState<string[]>([]);
  useEffect(() => {
    setLines([]);
    if (!id) return;
    const es = new EventSource(`/api/containers/${id}/logs?follow=true&tail=200`);
    es.onmessage = (e) =>
      setLines((l) => {
        const next = [...l, e.data];
        return next.length > 1000 ? next.slice(next.length - 1000) : next;
      });
    return () => es.close();
  }, [id]);
  return lines;
}

// useStatsStream streams a container's live resource stats over SSE.
export function useStatsStream(id: string | null): Stats | null {
  const [stats, setStats] = useState<Stats | null>(null);
  useEffect(() => {
    setStats(null);
    if (!id) return;
    const es = new EventSource(`/api/containers/${id}/stats`);
    es.onmessage = (e) => {
      try {
        setStats(JSON.parse(e.data));
      } catch {
        /* ignore malformed frame */
      }
    };
    return () => es.close();
  }, [id]);
  return stats;
}
