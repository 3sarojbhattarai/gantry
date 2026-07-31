import { humanSize, primaryTag, removeImage, shortId } from "@/lib/api";
import { useConfirm } from "@/components/confirm";
import { keys, useAction, useImages } from "@/lib/queries";
import { Button, EmptyState, ErrorState, Spinner, Toolbar } from "@/components/ui";
import { PruneButton } from "@/components/PruneButton";

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
