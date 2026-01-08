import { RefreshCw, Play, Square } from "lucide-react";
import { DataTable } from "@/components/ui/data-table";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useState, useEffect } from "react";
import { Container, Templates, api } from "@/lib/api";

export default function Containers() {
  const columns = [
    { key: "Name", header: "Name" },
    { key: "Id", header: "ID", className: "font-mono text-primary" },
    { key: "State", header: "State" },
    { key: "Actions", header: "Actions", render: (container: Container) => (
      <div className="flex gap-2">
          {container.State === "running" ? (
              <Button size="sm" variant="outline" onClick={async () => {
                  try {
                      await api.stopContainer(container.Id);
                      await fetchData();
                  } catch (error) {
                      console.error('Error stopping container:', error);
                  }
              }}>
                  <Square className="h-4 w-4" />
              </Button>
          ) : (
              <Button size="sm" variant="outline" onClick={async () => {
                  try {
                      await api.startContainer(container.Id);
                      await fetchData();
                  } catch (error) {
                      console.error('Error starting container:', error);
                  }
              }}>
                  <Play className="h-4 w-4" />
              </Button>
          )}
      </div>
    ) },
  ];
  const [containers, setContainers] = useState<Container[]>([]);
  const [loading, setLoading] = useState(true);
  const [updating, setUpdating] = useState(false);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [image, setImage] = useState("");
  const [name, setName] = useState("");
  const [templates, setTemplates] = useState<Templates | null>(null);
  const [templatesLoading, setTemplatesLoading] = useState(true);
  const [templateDialogOpen, setTemplateDialogOpen] = useState(false);
  const [selectedType, setSelectedType] = useState<"databases" | "repositorys">("databases");
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [containerName, setContainerName] = useState("");
  const [envValues, setEnvValues] = useState<Record<string, string>>({});

  const fetchData = async () => {
    try {
      const data = await api.getContainers(true);
      setContainers(data);
      setLoading(false);
    } catch (error) {
      console.error('Error fetching containers:', error);
      setLoading(false);
    }
  };

  const fetchTemplates = async () => {
    try {
      const data = await api.getTemplates();
      setTemplates(data);
    } catch (error) {
      console.error('Error fetching templates:', error);
    } finally {
      setTemplatesLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    fetchTemplates();
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
        <div className="flex gap-2">
          <Button
            onClick={() => setCreateDialogOpen(true)}
            variant="default"
          >
            Create Container
          </Button>
          <Button
            onClick={() => setTemplateDialogOpen(true)}
            variant="outline"
          >
            Create from Template
          </Button>
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

      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Container</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="image">Image</Label>
              <Input id="image" value={image} onChange={(e) => setImage(e.target.value)} placeholder="e.g. nginx:latest" />
            </div>
            <div>
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. my-container" />
            </div>
            <Button onClick={async () => {
              try {
                await api.createContainer(image, name);
                setCreateDialogOpen(false);
                setImage("");
                setName("");
                await fetchData();
              } catch (error) {
                console.error('Error creating container:', error);
              }
            }}>
              Create
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      <Dialog open={templateDialogOpen} onOpenChange={setTemplateDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create Container from Template</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div>
              <Label htmlFor="type">Type</Label>
              <Select value={selectedType} onValueChange={(value: "databases" | "repositorys") => { setSelectedType(value); setSelectedIndex(0); setEnvValues({}); }}>
                <SelectTrigger>
                  <SelectValue placeholder="Select type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="databases">Databases</SelectItem>
                  <SelectItem value="repositorys">Repositorys</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="template">Template</Label>
              <Select value={selectedIndex.toString()} onValueChange={(value) => { setSelectedIndex(parseInt(value)); setEnvValues({}); }}>
                <SelectTrigger>
                  <SelectValue placeholder="Select a template" />
                </SelectTrigger>
                <SelectContent>
                  {templatesLoading ? (
                    <SelectItem disabled value="loading">Loading templates...</SelectItem>
                  ) : templates && templates[selectedType] ? (
                    templates[selectedType].map((tmpl, idx) => (
                      <SelectItem key={idx} value={idx.toString()}>{tmpl.image}</SelectItem>
                    ))
                  ) : (
                    <SelectItem disabled value="none">No templates available</SelectItem>
                  )}
                </SelectContent>
              </Select>
            </div>
            <div>
              <Label htmlFor="name">Name</Label>
              <Input id="name" value={containerName} onChange={(e) => setContainerName(e.target.value)} placeholder="e.g. my-db" />
            </div>
            {templates && templates[selectedType][selectedIndex] && templates[selectedType][selectedIndex].environment.map((envKey) => (
              <div key={envKey}>
                <Label htmlFor={envKey}>{envKey}</Label>
                <Input id={envKey} value={envValues[envKey] || ""} onChange={(e) => setEnvValues({...envValues, [envKey]: e.target.value})} />
              </div>
            ))}
            <Button onClick={async () => {
              try {
                await api.createContainerFromTemplate(selectedType, selectedIndex, containerName, envValues);
                setTemplateDialogOpen(false);
                setContainerName("");
                setEnvValues({});
                await fetchData();
              } catch (error) {
                console.error('Error creating container from template:', error);
              }
            }}>
              Create
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}