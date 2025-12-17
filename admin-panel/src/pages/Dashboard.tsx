import { Activity, Globe, Zap } from "lucide-react";
import { StatCard } from "@/components/ui/stat-card";
import { trafficOriginData } from "@/data/mockData";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/status-badge";
import { MethodBadge } from "@/components/ui/method-badge";
import { useQuery } from "@tanstack/react-query";
import { api, Stats, RequestLog } from "@/lib/api";

const recentRequestsColumns = [
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

export default function Dashboard() {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['stats'],
    queryFn: api.getStats,
    refetchInterval: 5000, // Refresh every 5 seconds
  });

  const { data: requests, isLoading: requestsLoading } = useQuery({
    queryKey: ['requests'],
    queryFn: api.getRequests,
    refetchInterval: 5000,
  });

  if (statsLoading || requestsLoading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-foreground">Dashboard</h1>
        <p className="text-sm text-muted-foreground">
          Monitor your reverse proxy in real-time
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid gap-6 md:grid-cols-3">
        <StatCard
          title="Total Requests"
          value={stats?.totalRequests || 0}
          icon={Activity}
          trend={{ value: 12.5, isPositive: true }}
        />
        <StatCard
          title="Active Connections"
          value={stats?.activeConnections || 0}
          icon={Zap}
          trend={{ value: 3.2, isPositive: true }}
        />
        <StatCard
          title="Unique IPs"
          value={stats?.uniqueIPs || 0}
          icon={Globe}
          trend={{ value: 8.1, isPositive: true }}
        />
      </div>

      {/* Recent Requests */}
      <div className="space-y-4">
        <h2 className="text-lg font-semibold text-foreground">Recent Requests</h2>
        <DataTable data={(requests || []).slice(0, 5) as any[]} columns={recentRequestsColumns} />
      </div>
    </div>
  );
}
