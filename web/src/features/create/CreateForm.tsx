import { useState } from "react";
import { createContainer, exportSpec, getContainerSpec } from "@/lib/api";
import { keys, useAction, useContainers } from "@/lib/queries";
import { buildSpec, emptyForm, specToForm, type FormState } from "@/features/create/specForm";
import { Button } from "@/components/ui";

const RESTART = ["no", "always", "unless-stopped", "on-failure"];

export function CreateForm({ onClose }: { onClose: () => void }) {
  const [form, setForm] = useState<FormState>(emptyForm);
  const [advanced, setAdvanced] = useState(false);
  const [start, setStart] = useState(true);
  const [exported, setExported] = useState<string | null>(null);
  const { data: containers } = useContainers(true);
  const run = useAction();

  const set = (k: keyof FormState) => (e: { target: { value: string } }) =>
    setForm((f) => ({ ...f, [k]: e.target.value }));

  async function prefillFrom(id: string) {
    if (!id) return;
    try {
      const spec = await getContainerSpec(id);
      setForm(specToForm(spec));
    } catch {
      /* ignore */
    }
  }

  async function create() {
    const ok = await run(
      "Created container",
      () => createContainer(buildSpec(form), start),
      [keys.containers],
    );
    if (ok) onClose();
  }

  async function doExport(format: "run" | "compose") {
    try {
      const { text } = await exportSpec(format, buildSpec(form));
      setExported(text);
    } catch {
      /* ignore */
    }
  }

  return (
    <div className="fixed inset-0 z-40 flex items-start justify-center overflow-auto bg-black/60 py-8">
      <div className="w-[40rem] max-w-[92vw] rounded-lg border border-slate-700 bg-slate-900 p-5 shadow-2xl">
        <div className="mb-4 flex items-center justify-between">
          <h2 className="text-base font-semibold text-slate-100">Create container</h2>
          <select
            className="rounded border border-slate-700 bg-slate-800 px-2 py-1 text-xs text-slate-300"
            defaultValue=""
            onChange={(e) => prefillFrom(e.target.value)}
          >
            <option value="">Clone from…</option>
            {(containers ?? []).map((c) => (
              <option key={c.id} value={c.id}>
                {c.names[0] ?? c.id.slice(0, 12)}
              </option>
            ))}
          </select>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <Field label="Image *">
            <input className={inputCls} value={form.image} onChange={set("image")} placeholder="nginx:alpine" />
          </Field>
          <Field label="Name">
            <input className={inputCls} value={form.name} onChange={set("name")} placeholder="web" />
          </Field>
          <Field label="Command" wide>
            <input className={inputCls} value={form.command} onChange={set("command")} placeholder="nginx -g 'daemon off;'" />
          </Field>
          <Field label="Ports (one per line)">
            <textarea className={areaCls} value={form.ports} onChange={set("ports")} placeholder="8080:80/tcp" />
          </Field>
          <Field label="Env (KEY=value per line)">
            <textarea className={areaCls} value={form.env} onChange={set("env")} placeholder="TZ=UTC" />
          </Field>
          <Field label="Restart policy">
            <select className={inputCls} value={form.restartPolicy} onChange={set("restartPolicy")}>
              {RESTART.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </Field>
          <Field label="">
            <label className="flex items-center gap-2 pt-5 text-xs text-slate-400">
              <input type="checkbox" checked={start} onChange={(e) => setStart(e.target.checked)} />
              Start after create
            </label>
          </Field>
        </div>

        <button
          className="mt-3 text-xs text-sky-400 hover:underline"
          onClick={() => setAdvanced((a) => !a)}
        >
          {advanced ? "▾ Hide advanced" : "▸ Advanced (volumes, networks, labels)"}
        </button>

        {advanced && (
          <div className="mt-3 grid grid-cols-2 gap-3">
            <Field label="Volumes (src:dst[:ro] per line)">
              <textarea className={areaCls} value={form.volumes} onChange={set("volumes")} />
            </Field>
            <Field label="Networks (one per line)">
              <textarea className={areaCls} value={form.networks} onChange={set("networks")} />
            </Field>
            <Field label="Labels (k=v per line)">
              <textarea className={areaCls} value={form.labels} onChange={set("labels")} />
            </Field>
            <div className="grid grid-cols-2 gap-2">
              <Field label="Working dir">
                <input className={inputCls} value={form.workingDir} onChange={set("workingDir")} />
              </Field>
              <Field label="User">
                <input className={inputCls} value={form.user} onChange={set("user")} />
              </Field>
            </div>
          </div>
        )}

        {exported !== null && (
          <pre className="mt-3 max-h-40 overflow-auto rounded border border-slate-700 bg-slate-950 p-2 font-mono text-xs text-slate-300">
            {exported}
          </pre>
        )}

        <div className="mt-5 flex items-center justify-end gap-2">
          <Button onClick={() => doExport("run")}>Export run</Button>
          <Button onClick={() => doExport("compose")}>Export compose</Button>
          <div className="flex-1" />
          <Button onClick={onClose}>Cancel</Button>
          <Button variant="primary" onClick={create} disabled={!form.image.trim()}>
            Create
          </Button>
        </div>
      </div>
    </div>
  );
}

const inputCls =
  "w-full rounded border border-slate-700 bg-slate-800 px-2 py-1 text-sm text-slate-200 placeholder:text-slate-600";
const areaCls = `${inputCls} h-16 resize-none font-mono text-xs`;

function Field({ label, wide, children }: { label: string; wide?: boolean; children: React.ReactNode }) {
  return (
    <div className={wide ? "col-span-2" : ""}>
      {label && <label className="mb-1 block text-xs text-slate-400">{label}</label>}
      {children}
    </div>
  );
}
