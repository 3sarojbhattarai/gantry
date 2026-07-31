import { useState } from "react";
import {
  primaryName,
  removeContainer,
  restartContainer,
  startContainer,
  stopContainer,
  prune,
} from "@/lib/api";
import { useConfirm } from "@/components/confirm";
import { keys, useAction, useContainers } from "@/lib/queries";
import { Button, EmptyState, ErrorState, Spinner, StateBadge } from "@/components/ui";
import { useToast } from "@/components/toast";
import type { Container } from "@/lib/types";

export function ContainersView({
  selectedId,
  onSelect,
  onExec,
}: {
  selectedId: string | null;
  onSelect: (id: string | null) => void;
  onExec: (id: string) => void;
}) {
  const [all, setAll] = useState(true);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const { data, isLoading, error } = useContainers(all);
  const run = useAction();
  const confirm = useConfirm();
  const toast = useToast();

  if (isLoading) return <Spinner />;
  if (error) return <ErrorState message={(error as Error).message} />;
  const containers = data ?? [];

  const toggle = (id: string) =>
    setSelected((s) => {
      const next = new Set(s);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });

  const running = (c: Container) => c.state === "running";

  async function removeMany(ids: string[]) {
    if (ids.length === 0) return;
    const ok = await confirm({
      title: `Remove ${ids.length} container${ids.length > 1 ? "s" : ""}?`,
      body: "This cannot be undone.",
      confirmLabel: "Remove",
      danger: true,
    });
    if (!ok) return;
    for (const id of ids) {
      await run(`Removed ${id.slice(0, 12)}`, () => removeContainer(id, true), [keys.containers]);
    }
    setSelected(new Set());
  }

  async function pruneStopped() {
    const preview = await prune("containers", { confirm: false, dryRun: true });
    const n = preview.items?.length ?? 0;
    if (n === 0) {
      toast.success("Nothing to prune");
      return;
    }
    const ok = await confirm({
      title: `Prune ${n} stopped container${n > 1 ? "s" : ""}?`,
      confirmLabel: "Prune",
      danger: true,
    });
    if (!ok) return;
    await run("Pruned containers", () => prune("containers", { confirm: true, dryRun: false }), [
      keys.containers,
    ]);
  }

  return (
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-3 border-b border-slate-800 px-3 py-2">
        <label className="flex items-center gap-1.5 text-xs text-slate-400">
          <input type="checkbox" checked={all} onChange={(e) => setAll(e.target.checked)} />
          Show all
        </label>
        <div className="flex-1" />
        {selected.size > 0 && (
          <Button variant="danger" onClick={() => removeMany([...selected])}>
            Remove selected ({selected.size})
          </Button>
        )}
        <Button onClick={() => removeMany(containers.filter((c) => !running(c)).map((c) => c.id))}>
          Remove all stopped
        </Button>
        <Button variant="danger" onClick={pruneStopped}>
          Prune
        </Button>
      </div>

      <div className="flex-1 overflow-auto">
        {containers.length === 0 ? (
          <EmptyState>No containers</EmptyState>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="sticky top-0 bg-slate-900 text-xs uppercase text-slate-500">
              <tr>
                <th className="w-8 px-3 py-2" />
                <th className="px-3 py-2">Name</th>
                <th className="px-3 py-2">Image</th>
                <th className="px-3 py-2">State</th>
                <th className="px-3 py-2">Status</th>
                <th className="px-3 py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody>
              {containers.map((c) => (
                <tr
                  key={c.id}
                  onClick={() => onSelect(c.id)}
                  className={`cursor-pointer border-t border-slate-800 hover:bg-slate-800/50 ${
                    selectedId === c.id ? "bg-slate-800" : ""
                  }`}
                >
                  <td className="px-3 py-2" onClick={(e) => e.stopPropagation()}>
                    <input
                      type="checkbox"
                      checked={selected.has(c.id)}
                      onChange={() => toggle(c.id)}
                    />
                  </td>
                  <td className="px-3 py-2 font-medium text-slate-200">{primaryName(c.names)}</td>
                  <td className="px-3 py-2 text-slate-400">{c.image}</td>
                  <td className="px-3 py-2">
                    <StateBadge state={c.state} />
                  </td>
                  <td className="px-3 py-2 text-slate-400">{c.status}</td>
                  <td className="px-3 py-2" onClick={(e) => e.stopPropagation()}>
                    <div className="flex justify-end gap-1">
                      {running(c) ? (
                        <>
                          <Button onClick={() => run("Stopped", () => stopContainer(c.id), [keys.containers])}>
                            Stop
                          </Button>
                          <Button onClick={() => run("Restarted", () => restartContainer(c.id), [keys.containers])}>
                            Restart
                          </Button>
                          <Button onClick={() => onExec(c.id)}>Exec</Button>
                        </>
                      ) : (
                        <Button variant="primary" onClick={() => run("Started", () => startContainer(c.id), [keys.containers])}>
                          Start
                        </Button>
                      )}
                      <Button variant="danger" onClick={() => removeMany([c.id])}>
                        Remove
                      </Button>
                    </div>
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
