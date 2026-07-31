import { useState } from "react";
import { createNetwork, removeNetwork } from "@/lib/api";
import { useConfirm } from "@/components/confirm";
import { keys, useAction, useNetworks } from "@/lib/queries";
import { Button, EmptyState, ErrorState, Spinner, Toolbar } from "@/components/ui";
import { PruneButton } from "@/components/PruneButton";

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
