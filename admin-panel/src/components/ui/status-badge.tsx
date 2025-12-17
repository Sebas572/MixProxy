import { cn } from "@/lib/utils";

interface StatusBadgeProps {
  status: number;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const getStatusColor = (code: number) => {
    if (code >= 200 && code < 300) return "bg-success/20 text-success border-success/30";
    if (code >= 300 && code < 400) return "bg-primary/20 text-primary border-primary/30";
    if (code >= 400 && code < 500) return "bg-warning/20 text-warning border-warning/30";
    return "bg-destructive/20 text-destructive border-destructive/30";
  };

  return (
    <span
      className={cn(
        "inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-semibold",
        getStatusColor(status)
      )}
    >
      {status}
    </span>
  );
}
