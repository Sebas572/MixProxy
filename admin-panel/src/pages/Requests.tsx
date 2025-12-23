import { Search, Filter, RefreshCw } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { StatusBadge } from "@/components/ui/status-badge";
import { MethodBadge } from "@/components/ui/method-badge";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useState, useEffect } from "react";
import { RequestLog, api } from "@/lib/api";

const columns = [
  { key: "timestamp", header: "Time", className: "text-muted-foreground" },
  {
    key: "method",
    header: "Method",
    render: (item: any) => <MethodBadge method={item.method} />,
  },
  { key: "url", header: "URL", className: "text-primary" },
  { key: "ip", header: "IP" },
  { key: "subdomain", header: "subdomain" },
  {
    key: "status",
    header: "Status",
    render: (item: any) => <StatusBadge status={item.status} />,
  },
];

export default function Requests() {
  const [requests, setRequests] = useState<RequestLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [logFiles, setLogFiles] = useState<string[]>([]);
  const [selectedLog, setSelectedLog] = useState<string>("");

  const fetchData = async (date?: string) => {
    try {
      const text = await api.getLogs(date);
      const lines = text.trim().split('\n').filter(line => line.trim());
      const logs = lines.map(line => JSON.parse(line));
      const allRequests = logs.map((l: any, index: number) => ({
        id: index.toString(),
        timestamp: l.time,
        method: l.method,
        url: l.url,
        ip: l.ip,
        subdomain: l.sub,
        status: l.status,
      }));
      setRequests(allRequests);
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
          const latestFile = files[0]; // Assume sorted, take latest
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
          <h1 className="text-2xl font-bold text-foreground">Requests</h1>
          <p className="text-sm text-muted-foreground">
            View and analyze all incoming requests
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
            Showing {requests.length} requests
          </span>
        </div>
        <DataTable data={requests as any[]} columns={columns} />
      </div>
    </div>
  );
}
