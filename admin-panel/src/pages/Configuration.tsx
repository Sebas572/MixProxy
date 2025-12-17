import React, { useState } from "react";
import { Save, RefreshCw, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, Config } from "@/lib/api";

export default function Configuration() {
  const queryClient = useQueryClient();
  const { data: config, isLoading } = useQuery({
    queryKey: ['config'],
    queryFn: api.getConfig,
  });

  const [formData, setFormData] = useState<Config | null>(null);
  const [validationErrors, setValidationErrors] = useState<string[]>([]);

  React.useEffect(() => {
    if (config && !formData) {
      setFormData(config);
    }
  }, [config, formData]);

  const updateConfigMutation = useMutation({
    mutationFn: api.updateConfig,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config'] });
      setValidationErrors([]);
    },
  });

  const validateConfig = (cfg: Config): string[] => {
    const errors: string[] = [];

    // Validate hostname
    if (!cfg.hostname.trim()) {
      errors.push("Hostname is required");
    }

    // Validate admin panel subdomain
    if (!cfg.subdomain_admin_panel.trim()) {
      errors.push("Admin panel subdomain is required");
    }

    // Validate load balancer entries
    const subdomains = new Set<string>();
    cfg.load_balancer.forEach((lb, lbIndex) => {
      // Check unique subdomains
      if (subdomains.has(lb.subdomain)) {
        errors.push(`Duplicate subdomain "${lb.subdomain}"`);
      }
      subdomains.add(lb.subdomain);

      // Validate subdomain
      if (!lb.subdomain.trim()) {
        errors.push(`Load balancer ${lbIndex + 1}: Subdomain is required`);
      }

      // Validate type
      if (!lb.type.trim()) {
        errors.push(`Load balancer ${lbIndex + 1}: Type is required`);
      }

      // Validate VPS entries
      let activeCapacitySum = 0;
      let hasActiveVPS = false;

      lb.vps.forEach((vps, vpsIndex) => {
        // Validate IP
        if (!vps.ip.trim()) {
          errors.push(`Load balancer ${lbIndex + 1}, VPS ${vpsIndex + 1}: IP address is required`);
        } else {
          try {
            new URL(vps.ip);
          } catch {
            errors.push(`Load balancer ${lbIndex + 1}, VPS ${vpsIndex + 1}: Invalid IP/URL format`);
          }
        }

        // Validate capacity
        if (vps.capacity < 0 || vps.capacity > 1) {
          errors.push(`Load balancer ${lbIndex + 1}, VPS ${vpsIndex + 1}: Capacity must be between 0 and 1`);
        }

        if (vps.active) {
          hasActiveVPS = true;
          activeCapacitySum += vps.capacity;
        }
      });

      // Check capacity sum for active VPS
      if (hasActiveVPS && Math.abs(activeCapacitySum - 1) > 0.001) {
        errors.push(`Load balancer ${lbIndex + 1}: Sum of active VPS capacities must be 1 (currently ${activeCapacitySum.toFixed(3)})`);
      }

      // Check if there are any active VPS
      if (!hasActiveVPS) {
        errors.push(`Load balancer ${lbIndex + 1}: At least one VPS must be active`);
      }
    });

    return errors;
  };

  const handleSave = () => {
    if (!formData) return;

    const errors = validateConfig(formData);
    if (errors.length > 0) {
      setValidationErrors(errors);
      return;
    }

    updateConfigMutation.mutate(formData);
  };

  const updateField = (field: keyof Config, value: any) => {
    setFormData(prev => prev ? { ...prev, [field]: value } : null);
  };

  const updateLoadBalancer = (index: number, field: string, value: any) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      (newLB[index] as any)[field] = value;
      return { ...prev, load_balancer: newLB };
    });
  };

  const updateVPS = (lbIndex: number, vpsIndex: number, field: string, value: any) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      const newVPS = [...newLB[lbIndex].vps];
      (newVPS[vpsIndex] as any)[field] = value;
      newLB[lbIndex] = { ...newLB[lbIndex], vps: newVPS };
      return { ...prev, load_balancer: newLB };
    });
  };

  const addVPS = (lbIndex: number) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      newLB[lbIndex].vps.push({ ip: "", capacity: 0.5, active: true });
      return { ...prev, load_balancer: newLB };
    });
  };

  const removeVPS = (lbIndex: number, vpsIndex: number) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      newLB[lbIndex].vps.splice(vpsIndex, 1);
      return { ...prev, load_balancer: newLB };
    });
  };

  const addLoadBalancer = () => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      newLB.push({
        subdomain: "",
        type: "",
        active: true,
        vps: [{ ip: "", capacity: 1, active: true }]
      });
      return { ...prev, load_balancer: newLB };
    });
  };

  const removeLoadBalancer = (index: number) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      newLB.splice(index, 1);
      return { ...prev, load_balancer: newLB };
    });
  };

  if (isLoading || !formData) {
    return <div>Loading...</div>;
  }

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Configuration</h1>
          <p className="text-sm text-muted-foreground">
            Manage your reverse proxy settings
          </p>
        </div>
        <div className="flex gap-3">
          <Button variant="outline" className="gap-2">
            <RefreshCw className="h-4 w-4" />
            Reset
          </Button>
          <Button className="gap-2" onClick={handleSave} disabled={updateConfigMutation.isPending}>
            <Save className="h-4 w-4" />
            {updateConfigMutation.isPending ? 'Saving...' : 'Save Changes'}
          </Button>
        </div>
      </div>

      {/* General Settings */}
      <div className="rounded-xl border border-border bg-card p-6 space-y-6">
        <h2 className="text-lg font-semibold text-foreground">General Settings</h2>

        <div className="grid gap-6 md:grid-cols-2">
          <div className="space-y-2">
            <Label htmlFor="hostname">Hostname</Label>
            <Input
              id="hostname"
              value={formData.hostname}
              onChange={(e) => updateField('hostname', e.target.value)}
              className="bg-secondary border-border font-mono"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="subdomain_admin_panel">Admin Panel Subdomain</Label>
            <Input
              id="subdomain_admin_panel"
              value={formData.subdomain_admin_panel}
              onChange={(e) => updateField('subdomain_admin_panel', e.target.value)}
              className="bg-secondary border-border font-mono mt-0"
            />
          </div>
          <div className="center-switch">
            <Label htmlFor="on_https">HTTPS Enabled</Label>
            <Switch
              checked={formData.on_https}
              onCheckedChange={(checked) => updateField('on_https', checked)}
            />
          </div>
          <div className="center-switch">
            <Label htmlFor="mode_developer">Developer Mode</Label>
            <Switch
              checked={formData.mode_developer}
              onCheckedChange={(checked) => updateField('mode_developer', checked)}
            />
          </div>
        </div>
      </div>

      <Separator className="bg-border" />

      {/* Load Balancer */}
      <div className="rounded-xl border border-border bg-card p-6 space-y-6">
        <h2 className="text-lg font-semibold text-foreground">Load Balancer</h2>

        <div className="space-y-6">
          {formData.load_balancer.map((lb, i) => (
            <div key={i} className="border border-border rounded-lg p-4 space-y-4">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium">Server {i + 1}</h3>
                <Button size="sm" variant="ghost" onClick={() => removeLoadBalancer(i)}>
                  <Trash2 className="h-4 w-4" />
                </Button>
              </div>
              <div className="grid gap-4 md:grid-cols-3">
                <div className="space-y-2">
                  <Label>Subdomain</Label>
                  <Input
                    value={lb.subdomain}
                    onChange={(e) => updateLoadBalancer(i, 'subdomain', e.target.value)}
                    className="bg-secondary border-border font-mono"
                  />
                </div>
                <div className="space-y-2">
                  <Label>Type</Label>
                  <Input
                    value={lb.type}
                    onChange={(e) => updateLoadBalancer(i, 'type', e.target.value)}
                    className="bg-secondary border-border font-mono"
                  />
                </div>
                <div className="center-switch w-24">
                  <Label>Active</Label>
                  <Switch
                    checked={lb.active}
                    onCheckedChange={(checked) => updateLoadBalancer(i, 'active', checked)}
                  />
                </div>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <Label>VPS Servers</Label>
                  <Button size="sm" variant="outline" onClick={() => addVPS(i)}>
                    <Plus className="h-4 w-4 mr-1" />
                    Add VPS
                  </Button>
                </div>
                <div className="space-y-2">
                  {lb.vps.map((vps, j) => (
                    <div key={j} className="flex gap-2 items-center">
                      <Input
                        placeholder="IP Address"
                        value={vps.ip}
                        onChange={(e) => updateVPS(i, j, 'ip', e.target.value)}
                        className="bg-secondary border-border font-mono flex-1"
                      />
                      <Input
                        type="number"
                        step="0.1"
                        placeholder="Capacity"
                        value={vps.capacity}
                        onChange={(e) => updateVPS(i, j, 'capacity', parseFloat(e.target.value))}
                        className="bg-secondary border-border font-mono w-24"
                      />
                      <Switch
                        checked={vps.active}
                        onCheckedChange={(checked) => updateVPS(i, j, 'active', checked)}
                      />
                      <Button size="sm" variant="ghost" onClick={() => removeVPS(i, j)}>
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
          <Button variant="outline" onClick={addLoadBalancer}>
            <Plus className="h-4 w-4 mr-2" />
            Add Load Balancer
          </Button>
        </div>
      </div>
    </div>
  );
}
