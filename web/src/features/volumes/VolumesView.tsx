import { removeVolume } from "@/lib/api";
import { useConfirm } from "@/components/confirm";
import { keys, useAction, useVolumes } from "@/lib/queries";
import { Button, EmptyState, ErrorState, Spinner, Toolbar } from "@/components/ui";
import { PruneButton } from "@/components/PruneButton";

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
