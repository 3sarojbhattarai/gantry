import { prune } from "@/lib/api";
import { useConfirm } from "@/components/confirm";
import { useAction } from "@/lib/queries";
import { Button } from "@/components/ui";
import type { ResourceKind } from "@/lib/types";

// PruneButton previews (when supported) then confirms and runs a prune. Shared
// by the image/network/volume views.
export function PruneButton({ kind, hasDryRun }: { kind: ResourceKind; hasDryRun: boolean }) {
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
