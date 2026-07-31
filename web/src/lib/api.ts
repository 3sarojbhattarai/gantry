import type {
  Container,
  ContainerDetails,
  CreateSpec,
  ImageSummary,
  Network,
  PruneReport,
  ResourceKind,
  Volume,
} from "@/lib/types";

export class ApiError extends Error {
  status: number;
  constructor(message: string, status: number) {
    super(message);
    this.status = status;
  }
}

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init);
  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      if (body && typeof body.error === "string") message = body.error;
    } catch {
      /* non-JSON error body */
    }
    throw new ApiError(message, res.status);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

// --- reads ------------------------------------------------------------------

export const getContainers = (all: boolean) =>
  request<Container[]>(`/api/containers?all=${all}`);
export const getContainer = (id: string) =>
  request<ContainerDetails>(`/api/containers/${id}`);
export const getImages = () => request<ImageSummary[]>(`/api/images`);
export const getNetworks = () => request<Network[]>(`/api/networks`);
export const getVolumes = () => request<Volume[]>(`/api/volumes`);

// --- mutations --------------------------------------------------------------

const post = (url: string) => request<void>(url, { method: "POST" });
const del = (url: string) => request<void>(url, { method: "DELETE" });

export const startContainer = (id: string) => post(`/api/containers/${id}/start`);
export const stopContainer = (id: string) => post(`/api/containers/${id}/stop`);
export const restartContainer = (id: string) =>
  post(`/api/containers/${id}/restart`);
export const killContainer = (id: string) => post(`/api/containers/${id}/kill`);

// Destructive: confirm=true is the explicit consent the engine requires.
export const removeContainer = (id: string, force: boolean) =>
  del(`/api/containers/${id}?confirm=true&force=${force}`);
export const removeImage = (id: string, force: boolean) =>
  del(`/api/images/${id}?confirm=true&force=${force}`);
export const removeNetwork = (id: string) =>
  del(`/api/networks/${id}?confirm=true`);
export const removeVolume = (name: string, force: boolean) =>
  del(`/api/volumes/${encodeURIComponent(name)}?confirm=true&force=${force}`);

export const prune = (kind: ResourceKind, opts: { confirm: boolean; dryRun: boolean }) =>
  request<PruneReport>(
    `/api/prune/${kind}?confirm=${opts.confirm}&dryRun=${opts.dryRun}`,
    { method: "POST" },
  );

export const createNetwork = (name: string, driver: string) =>
  request<{ id: string }>(`/api/networks`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name, driver }),
  });

// --- create / exec ----------------------------------------------------------

export const getContainerSpec = (id: string) =>
  request<CreateSpec>(`/api/containers/${id}/spec`);

export const createContainer = (spec: CreateSpec, start: boolean) =>
  request<{ id: string }>(`/api/containers?start=${start}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(spec),
  });

export const exportSpec = (format: "run" | "compose", spec: CreateSpec) =>
  request<{ text: string }>(`/api/export/${format}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(spec),
  });

// --- formatting helpers -----------------------------------------------------

export function shortId(id: string): string {
  const s = id.replace(/^sha256:/, "");
  return s.length > 12 ? s.slice(0, 12) : s;
}

export function humanSize(bytes: number): string {
  if (bytes < 1000) return `${bytes} B`;
  const units = ["kB", "MB", "GB", "TB", "PB"];
  let v = bytes;
  let i = -1;
  do {
    v /= 1000;
    i++;
  } while (v >= 1000 && i < units.length - 1);
  return `${v.toFixed(1)} ${units[i]}`;
}

export function primaryName(names: string[] | null): string {
  return names && names.length > 0 ? names[0] : "<none>";
}

export function primaryTag(tags: string[] | null): string {
  return tags && tags.length > 0 ? tags[0] : "<none>";
}
