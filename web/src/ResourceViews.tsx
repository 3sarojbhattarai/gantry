import { useState } from "react";
import {
  createNetwork,
  humanSize,
  primaryTag,
  prune,
  removeImage,
  removeNetwork,
  removeVolume,
  shortId,
} from "@/lib/api";
import { useConfirm } from "@/components/confirm";
import { keys, useAction, useImages, useNetworks, useVolumes } from "@/lib/queries";
import { Button, EmptyState, ErrorState, Spinner } from "@/components/ui";
import type { ResourceKind } from "@/lib/types";

function PruneButton({ kind, hasDryRun }: { kind: ResourceKind; hasDryRun: boolean }) {
  const run = useAction();
  const confirm = useConfirm();
  async function onClick() {
    let title = `Prune unused ${kind}?`;
    if (hasDryRun) {
      const preview = await prune(kind, { confirm: false, dryRun: true });
      const n = preview.items?.length ?? 0;
      title = `Prune ${n} unused ${kind}?`;
    }
    const ok = await confirm({ title, confirmLabel: "Prune", danger: true });
    if (!ok) return;
    await run(`Pruned ${kind}`, () => prune(kind, { confirm: true, dryRun: false }), [[kind]]);
  }
  return (
    <Button variant="danger" onClick={onClick}>
      Prune
    </Button>
  );
}

function Toolbar({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex items-center gap-2 border-b border-slate-800 px-3 py-2">
      <div className="flex-1" />
      {children}
    </div>
  );
}

export function ImagesView() {
  const { data, isLoading, error } = useImages();
  const run = useAction();
  const confirm = useConfirm();
  if (isLoading) return <Spinner />;
  if (error) return <ErrorState message={(error as Error).message} />;
  const images = data ?? [];

  async function remove(id: string) {
    const ok = await confirm({ title: "Remove image?", confirmLabel: "Remove", danger: true });
    if (!ok) return;
    await run("Removed image", () => removeImage(id, true), [keys.images]);
  }

  return (
    <div className="flex h-full flex-col">
      <Toolbar>
        <PruneButton kind="images" hasDryRun />
      </Toolbar>
      <div className="flex-1 overflow-auto">
        {images.length === 0 ? (
          <EmptyState>No images</EmptyState>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="sticky top-0 bg-slate-900 text-xs uppercase text-slate-500">
              <tr>
                <th className="px-3 py-2">Repository:Tag</th>
                <th className="px-3 py-2">ID</th>
                <th className="px-3 py-2">Size</th>
                <th className="px-3 py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {images.map((im) => (
                <tr key={im.id} className="border-t border-slate-800 hover:bg-slate-800/50">
                  <td className="px-3 py-2 font-medium text-slate-200">{primaryTag(im.repoTags)}</td>
                  <td className="px-3 py-2 font-mono text-xs text-slate-400">{shortId(im.id)}</td>
                  <td className="px-3 py-2 text-slate-400">{humanSize(im.size)}</td>
                  <td className="px-3 py-2 text-right">
                    <Button variant="danger" onClick={() => remove(im.id)}>
                      Remove
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export function NetworksView() {
  const { data, isLoading, error } = useNetworks();
  const run = useAction();
  const confirm = useConfirm();
  const [name, setName] = useState("");
  if (isLoading) return <Spinner />;
  if (error) return <ErrorState message={(error as Error).message} />;
  const networks = data ?? [];

  async function remove(id: string) {
    const ok = await confirm({ title: "Remove network?", confirmLabel: "Remove", danger: true });
    if (!ok) return;
    await run("Removed network", () => removeNetwork(id), [keys.networks]);
  }
  async function create() {
    if (!name.trim()) return;
    const ok = await run("Created network", () => createNetwork(name.trim(), "bridge"), [keys.networks]);
    if (ok) setName("");
  }

  return (
    <div className="flex h-full flex-col">
      <Toolbar>
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="new-network"
          className="rounded border border-slate-700 bg-slate-800 px-2 py-1 text-xs text-slate-200"
        />
        <Button variant="primary" onClick={create}>
          Create
        </Button>
        <PruneButton kind="networks" hasDryRun={false} />
      </Toolbar>
      <div className="flex-1 overflow-auto">
        {networks.length === 0 ? (
          <EmptyState>No networks</EmptyState>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="sticky top-0 bg-slate-900 text-xs uppercase text-slate-500">
              <tr>
                <th className="px-3 py-2">Name</th>
                <th className="px-3 py-2">Driver</th>
                <th className="px-3 py-2">Scope</th>
                <th className="px-3 py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {networks.map((n) => (
                <tr key={n.id} className="border-t border-slate-800 hover:bg-slate-800/50">
                  <td className="px-3 py-2 font-medium text-slate-200">{n.name}</td>
                  <td className="px-3 py-2 text-slate-400">{n.driver}</td>
                  <td className="px-3 py-2 text-slate-400">{n.scope}</td>
                  <td className="px-3 py-2 text-right">
                    <Button variant="danger" onClick={() => remove(n.id)}>
                      Remove
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

export function VolumesView() {
  const { data, isLoading, error } = useVolumes();
  const run = useAction();
  const confirm = useConfirm();
  if (isLoading) return <Spinner />;
  if (error) return <ErrorState message={(error as Error).message} />;
  const volumes = data ?? [];

  async function remove(nm: string) {
    const ok = await confirm({ title: `Remove volume ${nm}?`, confirmLabel: "Remove", danger: true });
    if (!ok) return;
    await run("Removed volume", () => removeVolume(nm, true), [keys.volumes]);
  }

  return (
    <div className="flex h-full flex-col">
      <Toolbar>
        <PruneButton kind="volumes" hasDryRun={false} />
      </Toolbar>
      <div className="flex-1 overflow-auto">
        {volumes.length === 0 ? (
          <EmptyState>No volumes</EmptyState>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="sticky top-0 bg-slate-900 text-xs uppercase text-slate-500">
              <tr>
                <th className="px-3 py-2">Name</th>
                <th className="px-3 py-2">Driver</th>
                <th className="px-3 py-2">Mountpoint</th>
                <th className="px-3 py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {volumes.map((v) => (
                <tr key={v.name} className="border-t border-slate-800 hover:bg-slate-800/50">
                  <td className="px-3 py-2 font-medium text-slate-200">{v.name}</td>
                  <td className="px-3 py-2 text-slate-400">{v.driver}</td>
                  <td className="px-3 py-2 font-mono text-xs text-slate-400">{v.mountpoint}</td>
                  <td className="px-3 py-2 text-right">
                    <Button variant="danger" onClick={() => remove(v.name)}>
                      Remove
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
