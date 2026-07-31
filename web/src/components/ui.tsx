import type { ButtonHTMLAttributes, ReactNode } from "react";

type Variant = "default" | "primary" | "danger";

const variants: Record<Variant, string> = {
  default: "border-slate-700 bg-slate-800 text-slate-200 hover:bg-slate-700",
  primary: "border-sky-700 bg-sky-700 text-white hover:bg-sky-600",
  danger: "border-red-800 bg-red-900/60 text-red-200 hover:bg-red-800",
};

export function Button({
  variant = "default",
  className = "",
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant }) {
  return (
    <button
      className={`rounded border px-2 py-1 text-xs font-medium transition-colors disabled:opacity-40 ${variants[variant]} ${className}`}
      {...props}
    />
  );
}

export function StateBadge({ state }: { state: string }) {
  const color =
    state === "running"
      ? "bg-emerald-900 text-emerald-300"
      : state === "paused"
        ? "bg-amber-900 text-amber-300"
        : "bg-slate-700 text-slate-300";
  return (
    <span className={`inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs ${color}`}>
      <span className="h-1.5 w-1.5 rounded-full bg-current" />
      {state}
    </span>
  );
}

export function Spinner() {
  return (
    <div className="flex items-center justify-center p-8 text-slate-500">
      <span className="h-5 w-5 animate-spin rounded-full border-2 border-slate-600 border-t-slate-300" />
    </div>
  );
}

export function EmptyState({ children }: { children: ReactNode }) {
  return <div className="p-8 text-center text-sm text-slate-500">{children}</div>;
}

export function ErrorState({ message }: { message: string }) {
  return <div className="p-8 text-center text-sm text-red-400">{message}</div>;
}

export function Toolbar({ children }: { children: ReactNode }) {
  return (
    <div className="flex items-center gap-2 border-b border-slate-800 px-3 py-2">
      <div className="flex-1" />
      {children}
    </div>
  );
}
