// Wire types mirroring the JSON the Go API returns (see internal/docker/types.go).

export interface Port {
  ip: string;
  private: number;
  public: number;
  type: string;
}

export interface Container {
  id: string;
  names: string[];
  image: string;
  imageId: string;
  command: string;
  created: string;
  state: string;
  status: string;
  ports: Port[] | null;
  labels: Record<string, string> | null;
}

export interface ContainerDetails extends Container {
  path: string;
  args: string[] | null;
  env: string[] | null;
  platform: string;
  restartCount: number;
  exitCode: number;
  error: string;
  startedAt: string;
  finishedAt: string;
}

export interface ImageSummary {
  id: string;
  repoTags: string[] | null;
  created: string;
  size: number;
  labels: Record<string, string> | null;
}

export interface Network {
  id: string;
  name: string;
  driver: string;
  scope: string;
  created: string;
  internal: boolean;
  labels: Record<string, string> | null;
}

export interface Volume {
  name: string;
  driver: string;
  mountpoint: string;
  created: string;
  labels: Record<string, string> | null;
}

export interface Stats {
  containerId: string;
  cpuPercent: number;
  memUsage: number;
  memLimit: number;
  memPercent: number;
  netRx: number;
  netTx: number;
  blockRead: number;
  blockWrite: number;
  pids: number;
}

export interface DaemonEvent {
  type: string;
  action: string;
  actor: string;
  name: string;
  time: string;
}

export interface PruneReport {
  items: string[] | null;
  spaceReclaimed: number;
  dryRun: boolean;
}

export type ResourceKind = "containers" | "images" | "networks" | "volumes";

export interface PortMapping {
  host: string;
  container: number;
  proto: string;
}

export interface CreateSpec {
  image: string;
  name: string;
  command: string[] | null;
  env: string[] | null;
  ports: PortMapping[] | null;
  restartPolicy: string;
  volumes: string[] | null;
  networks: string[] | null;
  labels: Record<string, string> | null;
  workingDir: string;
  user: string;
}
