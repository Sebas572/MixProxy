import { Search, Shield, AlertTriangle, RefreshCw } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
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
  const [updating, setUpdating] = useState(false);
  const [logFiles, setLogFiles] = useState<string[]>([]);
  const [selectedLog, setSelectedLog] = useState<string>("");

  const fetchData = async (date?: string) => {
    try {
      const text = await api.getLogs(date);
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

  useEffect(() => {
    const init = async () => {
      try {
        const files = await api.getLogList();
        setLogFiles(files);
        if (files.length > 0) {
          const latestFile = files[0];
          setSelectedLog(latestFile);
          const date = latestFile.replace('log-', '');
          await fetchData(date);
        }
      } catch (error) {
        console.error('Error initializing:', error);
        setLoading(false);
      }
    };
    init();
  }, []);

  useEffect(() => {
    if (selectedLog) {
      const date = selectedLog.replace('log-', '');
      fetchData(date);
    }
  }, [selectedLog]);

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">IP Addresses</h1>
          <p className="text-sm text-muted-foreground">
            Monitor and manage IP addresses accessing your proxy
          </p>
        </div>
        <div className="flex items-center gap-4">
          <Select value={selectedLog} onValueChange={setSelectedLog}>
            <SelectTrigger className="w-48">
              <SelectValue placeholder="Select log file" />
            </SelectTrigger>
            <SelectContent>
              {logFiles.map((file) => (
                <SelectItem key={file} value={file}>
                  {file}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          <Button
            onClick={async () => {
              setUpdating(true);
              const date = selectedLog.replace('log-', '');
              await fetchData(date);
              setUpdating(false);
            }}
            disabled={updating || !selectedLog}
            variant="outline"
            size="sm"
          >
            <RefreshCw className="mr-2 h-4 w-4" />
            {updating ? "Updating..." : "Update Logs"}
          </Button>
        </div>
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
