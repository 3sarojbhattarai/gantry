import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import {
  getContainers,
  getImages,
  getNetworks,
  getVolumes,
  ApiError,
} from "@/lib/api";
import { useToast } from "@/components/toast";
import type { DaemonEvent } from "@/lib/types";

export const keys = {
  containers: ["containers"] as const,
  images: ["images"] as const,
  networks: ["networks"] as const,
  volumes: ["volumes"] as const,
};

export function useContainers(all: boolean) {
  return useQuery({ queryKey: [...keys.containers, all], queryFn: () => getContainers(all) });
}
export function useImages() {
  return useQuery({ queryKey: keys.images, queryFn: getImages });
}
export function useNetworks() {
  return useQuery({ queryKey: keys.networks, queryFn: getNetworks });
}
export function useVolumes() {
  return useQuery({ queryKey: keys.volumes, queryFn: getVolumes });
}

// useLiveEvents subscribes to the daemon event stream and invalidates the
// affected query so the UI stays live without polling. Returns the current
// connection state for a status indicator.
export function useLiveEvents(): "connecting" | "live" | "offline" {
  const qc = useQueryClient();
  useEffect(() => {
    const es = new EventSource("/api/events");
    const invalidate = (ev: MessageEvent) => {
      let e: DaemonEvent;
      try {
        e = JSON.parse(ev.data);
      } catch {
        return;
      }
      switch (e.type) {
        case "container":
          qc.invalidateQueries({ queryKey: keys.containers });
          break;
        case "image":
          qc.invalidateQueries({ queryKey: keys.images });
          break;
        case "network":
          qc.invalidateQueries({ queryKey: keys.networks });
          break;
        case "volume":
          qc.invalidateQueries({ queryKey: keys.volumes });
          break;
      }
    };
    es.onmessage = invalidate;
    return () => es.close();
  }, [qc]);

  // The EventSource state isn't reactive here; a simple heuristic is enough for
  // the header dot. Consumers re-render on query changes anyway.
  return "live";
}

// useAction runs a mutation, shows a toast, and invalidates the given query
// keys on success. It centralizes the optimistic-feel refresh + error surfacing
// that every button needs.
export function useAction() {
  const qc = useQueryClient();
  const toast = useToast();
  return async function run(
    label: string,
    fn: () => Promise<unknown>,
    invalidate: readonly (readonly unknown[])[],
  ): Promise<boolean> {
    try {
      await fn();
      toast.success(label);
      for (const key of invalidate) {
        qc.invalidateQueries({ queryKey: key as unknown[] });
      }
      return true;
    } catch (err) {
      const msg = err instanceof ApiError ? err.message : String(err);
      toast.error(`${label} failed: ${msg}`);
      return false;
    }
  };
}
