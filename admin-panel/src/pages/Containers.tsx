import { RefreshCw } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { Button } from "@/components/ui/button";
import { useState, useEffect } from "react";
import { Container, api } from "@/lib/api";

const columns = [
  { key: "Name", header: "Name" },
  { key: "Id", header: "ID", className: "font-mono text-primary" },
];

export default function Containers() {
  const [containers, setContainers] = useState<Container[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);

  const fetchData = async () => {
    try {
      const data = await api.getContainers();
      setContainers(data);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching containers:', error);
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  if (loading) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Containers</h1>
          <p className="text-sm text-muted-foreground">
            View all running containers in the network
          </p>
        </div>
        <Button
          onClick={async () => {
            setUpdating(true);
            await fetchData();
            setUpdating(false);
          }}
          disabled={updating}
          variant="outline"
          size="sm"
        >
          <RefreshCw className={`mr-2 h-4 w-4 ${updating ? 'animate-spin' : ''}`} />
          {updating ? "Updating..." : "Refresh"}
        </Button>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-3">
        <div className="rounded-xl border border-border bg-card p-4">
          <p className="text-sm text-muted-foreground">Total Containers</p>
          <p className="text-2xl font-bold text-foreground">{containers.length}</p>
        </div>
      </div>

      {/* Table */}
      <DataTable data={containers as any[]} columns={columns} />
    </div>
  );
}