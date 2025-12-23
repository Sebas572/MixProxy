import { Search, Shield, AlertTriangle } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { useState, useEffect } from "react";
import { IPStat, api } from "@/lib/api";

const columns = [
  {
    key: "ip",
    header: "IP Address",
    render: (item: any) => (
      <span className="font-mono text-primary">{item.ip}</span>
    ),
  },
  {
    key: "count",
    header: "Request Count",
    render: (item: any) => (
      <span className="font-mono">{item.count.toLocaleString()}</span>
    ),
  },
  { key: "lastSeen", header: "Last Seen", className: "text-muted-foreground" },
  {
    key: "actions",
    header: "Actions",
    render: () => (
      <div className="flex gap-2">
        <Button variant="ghost" size="sm" className="h-7 text-xs">
          <Shield className="mr-1 h-3 w-3" />
          Whitelist
        </Button>
        <Button variant="ghost" size="sm" className="h-7 text-xs text-destructive hover:text-destructive">
          <AlertTriangle className="mr-1 h-3 w-3" />
          Block
        </Button>
      </div>
    ),
  },
];

export default function IPs() {
  const [ips, setIps] = useState<IPStat[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const text = await api.getLogs();
        const lines = text.trim().split('\n').filter(line => line.trim());
        const logs = lines.map(line => JSON.parse(line));
        const ipMap = new Map<string, {ip: string, count: number, lastSeen: string}>();
        logs.forEach((l: any) => {
          if (!ipMap.has(l.ip)) {
            ipMap.set(l.ip, {ip: l.ip, count: 0, lastSeen: l.time});
          }
          const entry = ipMap.get(l.ip)!;
          entry.count++;
          if (new Date(l.time) > new Date(entry.lastSeen)) {
            entry.lastSeen = l.time;
          }
        });
        setIps(Array.from(ipMap.values()));
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
        <h1 className="text-2xl font-bold text-foreground">IP Addresses</h1>
        <p className="text-sm text-muted-foreground">
          Monitor and manage IP addresses accessing your proxy
        </p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-4">
          <p className="text-sm text-muted-foreground">Total IPs</p>
          <p className="text-2xl font-bold text-foreground">{ips.length.toLocaleString()}</p>
        </div>
        {/* Coming soon */}
        {/* <div className="rounded-xl border border-success/30 bg-success/10 p-4">
          <p className="text-sm text-success">Whitelisted</p>
          <p className="text-2xl font-bold text-success">24</p>
        </div>
        <div className="rounded-xl border border-destructive/30 bg-destructive/10 p-4">
          <p className="text-sm text-destructive">Blocked</p>
          <p className="text-2xl font-bold text-destructive">156</p>
        </div> */}
      </div>

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Search IP addresses..."
          className="pl-10 bg-secondary border-border"
        />
      </div>

      {/* Table */}
      <DataTable data={ips as any[]} columns={columns} />
    </div>
  );
}
