import { useState } from "react";
import { ContainersView } from "@/features/containers/ContainersView";
import { ContainerDetail } from "@/features/containers/ContainerDetail";
import { CreateForm } from "@/features/create/CreateForm";
import { ExecTerminal } from "@/features/containers/ExecTerminal";
import { ImagesView } from "@/features/images/ImagesView";
import { NetworksView } from "@/features/networks/NetworksView";
import { VolumesView } from "@/features/volumes/VolumesView";
import { useLiveEvents } from "@/lib/queries";
import { Button } from "@/components/ui";
import type { ResourceKind } from "@/lib/types";

const TABS: { key: ResourceKind; label: string }[] = [
  { key: "containers", label: "Containers" },
  { key: "images", label: "Images" },
  { key: "networks", label: "Networks" },
  { key: "volumes", label: "Volumes" },
];

export function App() {
  const [tab, setTab] = useState<ResourceKind>("containers");
  const [selected, setSelected] = useState<string | null>(null);
  const [execId, setExecId] = useState<string | null>(null);
  const [creating, setCreating] = useState(false);
  const live = useLiveEvents();

  return (
    <div className="flex h-full flex-col">
      <header className="flex items-center gap-4 border-b border-slate-800 bg-slate-900 px-4 py-2">
        <span className="text-lg font-semibold text-slate-100">gantry</span>
        <nav className="flex gap-1">
          {TABS.map((t) => (
            <button
              key={t.key}
              onClick={() => {
                setTab(t.key);
                if (t.key !== "containers") {
                  setSelected(null);
                  setExecId(null);
                }
              }}
              className={`rounded px-3 py-1 text-sm ${
                tab === t.key ? "bg-slate-700 text-white" : "text-slate-400 hover:text-slate-200"
              }`}
            >
              {t.label}
            </button>
          ))}
        </nav>
        <div className="flex-1" />
        {tab === "containers" && (
          <Button variant="primary" onClick={() => setCreating(true)}>
            + New container
          </Button>
        )}
        <span className="flex items-center gap-1.5 text-xs text-slate-400">
          <span
            className={`h-2 w-2 rounded-full ${live === "live" ? "bg-emerald-500" : "bg-slate-500"}`}
          />
          {live}
        </span>
      </header>

      <main className="flex min-h-0 flex-1">
        <section className="min-w-0 flex-1">
          {tab === "containers" && (
            <ContainersView
              selectedId={selected}
              onSelect={(id) => {
                setSelected(id);
                setExecId(null);
              }}
              onExec={(id) => {
                setExecId(id);
                setSelected(null);
              }}
            />
          )}
          {tab === "images" && <ImagesView />}
          {tab === "networks" && <NetworksView />}
          {tab === "volumes" && <VolumesView />}
        </section>

        {tab === "containers" && execId && (
          <aside className="w-[34rem] shrink-0">
            <ExecTerminal id={execId} onClose={() => setExecId(null)} />
          </aside>
        )}
        {tab === "containers" && !execId && selected && (
          <aside className="w-[28rem] shrink-0">
            <ContainerDetail id={selected} onClose={() => setSelected(null)} />
          </aside>
        )}
      </main>

      {creating && <CreateForm onClose={() => setCreating(false)} />}
    </div>
  );
}
