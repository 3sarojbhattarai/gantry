import { useEffect, useRef } from "react";
import { humanSize } from "@/lib/api";
import { useLogStream, useStatsStream } from "@/features/containers/streams";
import { Button } from "@/components/ui";

export function ContainerDetail({ id, onClose }: { id: string; onClose: () => void }) {
  const lines = useLogStream(id);
  const stats = useStatsStream(id);
  const logRef = useRef<HTMLDivElement>(null);

  // Auto-scroll to the newest log line.
  useEffect(() => {
    const el = logRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [lines]);

  return (
    <div className="flex h-full flex-col border-l border-slate-800 bg-slate-900">
      <div className="flex items-center justify-between border-b border-slate-800 px-3 py-2">
        <span className="font-mono text-xs text-slate-400">{id.slice(0, 12)}</span>
        <Button onClick={onClose}>Close</Button>
      </div>

      <div className="grid grid-cols-2 gap-2 border-b border-slate-800 p-3 text-xs">
        <Stat label="CPU" value={stats ? `${stats.cpuPercent.toFixed(1)}%` : "…"} />
        <Stat
          label="Memory"
          value={stats ? `${humanSize(stats.memUsage)} / ${humanSize(stats.memLimit)}` : "…"}
        />
        <Stat
          label="Net I/O"
          value={stats ? `↓ ${humanSize(stats.netRx)}  ↑ ${humanSize(stats.netTx)}` : "…"}
        />
        <Stat label="PIDs" value={stats ? String(stats.pids) : "…"} />
      </div>

      <div className="px-3 py-1.5 text-xs uppercase tracking-wide text-slate-500">Logs</div>
      <div
        ref={logRef}
        className="flex-1 overflow-auto whitespace-pre-wrap break-all px-3 pb-3 font-mono text-xs leading-relaxed text-slate-300"
      >
        {lines.length === 0 ? (
          <span className="text-slate-600">Waiting for log output…</span>
        ) : (
          lines.map((l, i) => <div key={i}>{l}</div>)
        )}
      </div>
    </div>
  );
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div>
      <div className="text-[10px] uppercase tracking-wide text-slate-500">{label}</div>
      <div className="text-slate-200">{value}</div>
    </div>
  );
}
