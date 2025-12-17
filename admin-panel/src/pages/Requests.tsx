import { Search, Filter } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/status-badge";
import { MethodBadge } from "@/components/ui/method-badge";
import { trafficOriginData } from "@/data/mockData";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useQuery } from "@tanstack/react-query";
import { api, RequestLog } from "@/lib/api";

const columns = [
  { key: "timestamp", header: "Time", className: "text-muted-foreground" },
  {
    key: "method",
    header: "Method",
    render: (item: any) => <MethodBadge method={item.method} />,
  },
  { key: "url", header: "URL", className: "text-primary" },
  { key: "ip", header: "IP" },
  {
    key: "status",
    header: "Status",
    render: (item: any) => <StatusBadge status={item.status} />,
  },
];

export default function Requests() {
  const { data: requests, isLoading } = useQuery({
    queryKey: ['requests'],
    queryFn: api.getRequests,
    refetchInterval: 5000,
  });

  if (isLoading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Requests</h1>
          <p className="text-sm text-muted-foreground">
            View and analyze all incoming requests
          </p>
        </div>
      </div>

      {/* Search and Filters */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Search by URL, IP, or method..."
            className="pl-10 bg-secondary border-border"
          />
        </div>
        <Button variant="outline" className="gap-2">
          <Filter className="h-4 w-4" />
          Filters
        </Button>
      </div>

      {/* Requests Table */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-lg font-semibold text-foreground">All Requests</h2>
          <span className="text-sm text-muted-foreground">
            Showing {(requests || []).length} requests
          </span>
        </div>
        <DataTable data={(requests || []) as any[]} columns={columns} />
      </div>
    </div>
  );
}
