import type { CreateSpec, PortMapping } from "@/lib/types";

// FormState holds the create form's raw text fields. List fields (env, ports,
// …) are edited as one-per-line text and parsed into a CreateSpec on submit.
export interface FormState {
  image: string;
  name: string;
  command: string;
  env: string;
  ports: string;
  volumes: string;
  networks: string;
  labels: string;
  restartPolicy: string;
  workingDir: string;
  user: string;
}

export const emptyForm: FormState = {
  image: "",
  name: "",
  command: "",
  env: "",
  ports: "",
  volumes: "",
  networks: "",
  labels: "",
  restartPolicy: "no",
  workingDir: "",
  user: "",
};

function lines(text: string): string[] {
  return text
    .split("\n")
    .map((l) => l.trim())
    .filter(Boolean);
}

// parsePort accepts "8080:80/tcp", "8080:80", "80/tcp", or "80".
export function parsePort(s: string): PortMapping | null {
  let proto = "tcp";
  let rest = s.trim();
  const slash = rest.lastIndexOf("/");
  if (slash >= 0) {
    proto = rest.slice(slash + 1) || "tcp";
    rest = rest.slice(0, slash);
  }
  const parts = rest.split(":");
  let host = "";
  let container = parts[0];
  if (parts.length === 2) {
    host = parts[0];
    container = parts[1];
  }
  const c = parseInt(container, 10);
  if (!c) return null;
  return { host: host.trim(), container: c, proto };
}

export function parsePorts(text: string): PortMapping[] {
  return lines(text)
    .map(parsePort)
    .filter((p): p is PortMapping => p !== null);
}

export function parseLabels(text: string): Record<string, string> {
  const out: Record<string, string> = {};
  for (const line of lines(text)) {
    const i = line.indexOf("=");
    if (i > 0) out[line.slice(0, i)] = line.slice(i + 1);
  }
  return out;
}

export function buildSpec(f: FormState): CreateSpec {
  return {
    image: f.image.trim(),
    name: f.name.trim(),
    command: f.command.trim() ? f.command.trim().split(/\s+/) : null,
    env: lines(f.env),
    ports: parsePorts(f.ports),
    restartPolicy: f.restartPolicy,
    volumes: lines(f.volumes),
    networks: lines(f.networks),
    labels: parseLabels(f.labels),
    workingDir: f.workingDir.trim(),
    user: f.user.trim(),
  };
}

// specToForm fills the form from a spec (the --from prefill).
export function specToForm(s: CreateSpec): FormState {
  const ports = (s.ports ?? [])
    .map((p) => (p.host ? `${p.host}:${p.container}/${p.proto}` : `${p.container}/${p.proto}`))
    .join("\n");
  const labels = Object.entries(s.labels ?? {})
    .map(([k, v]) => `${k}=${v}`)
    .join("\n");
  return {
    image: s.image ?? "",
    name: s.name ?? "",
    command: (s.command ?? []).join(" "),
    env: (s.env ?? []).join("\n"),
    ports,
    volumes: (s.volumes ?? []).join("\n"),
    networks: (s.networks ?? []).join("\n"),
    labels,
    restartPolicy: s.restartPolicy || "no",
    workingDir: s.workingDir ?? "",
    user: s.user ?? "",
  };
}
