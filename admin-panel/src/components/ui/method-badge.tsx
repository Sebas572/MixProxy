import { cn } from "@/lib/utils";

interface MethodBadgeProps {
  method: string;
}

export function MethodBadge({ method }: MethodBadgeProps) {
  const getMethodColor = (m: string) => {
    switch (m.toUpperCase()) {
      case "GET":
        return "bg-success/20 text-success";
      case "POST":
        return "bg-primary/20 text-primary";
      case "PUT":
        return "bg-warning/20 text-warning";
      case "DELETE":
        return "bg-destructive/20 text-destructive";
      case "PATCH":
        return "bg-purple-500/20 text-purple-400";
      default:
        return "bg-muted text-muted-foreground";
    }
  };

  return (
    <span
      className={cn(
        "inline-flex items-center rounded px-2 py-0.5 text-xs font-bold uppercase",
        getMethodColor(method)
      )}
    >
      {method}
    </span>
  );
}
