import { Activity, Globe, Zap } from "lucide-react";
import { StatCard } from "@/components/ui/stat-card";
import { trafficOriginData } from "@/data/mockData";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/status-badge";
import { MethodBadge } from "@/components/ui/method-badge";
import { useState, useEffect } from "react";
import { Stats, RequestLog, api } from "@/lib/api";

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
  const [stats, setStats] = useState<Stats | null>(null);
  const [requests, setRequests] = useState<RequestLog[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const text = await api.getLogs();
        const lines = text.trim().split('\n').filter(line => line.trim());
        const logs = lines.map(line => JSON.parse(line));
        const totalRequests = logs.length;
        const uniqueIPs = new Set(logs.map((l: any) => l.ip)).size;
        const now = new Date();
        const oneMinAgo = new Date(now.getTime() - 60000);
        const activeConnections = new Set(logs.filter((l: any) => new Date(l.time) > oneMinAgo).map((l: any) => l.ip)).size;
        setStats({ totalRequests, activeConnections, uniqueIPs } as Stats);
        const recentRequests = logs.slice(-5).reverse().map((l: any, index: number) => ({
          id: index.toString(),
          timestamp: l.time,
          method: l.method,
          url: l.url,
          ip: l.ip,
          status: l.status,
        }));
        setRequests(recentRequests);
        setLoading(false);
      } catch (error) {
        console.error('Error fetching logs:', error);
        setLoading(false);
      }
    };
    fetchData();
    const interval = setInterval(fetchData, 5000);
    return () => clearInterval(interval);
  }, []);

  if (loading) {
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
        <DataTable data={requests as any[]} columns={recentRequestsColumns} />
      </div>
    </div>
  );
}
