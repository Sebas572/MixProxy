import React, { useState } from "react";
import { Save, RefreshCw, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, Config } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";

export default function Configuration() {
  const queryClient = useQueryClient();
  const { toast } = useToast();
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
      toast({
        title: "Success",
        description: "Configuration saved successfully.",
      });
    },
    onError: () => {
      toast({
        title: "Error",
        description: "Failed to save configuration.",
        variant: "destructive",
      });
    },
  });

  const reloadMutation = useMutation({
    mutationFn: api.reload,
    onSuccess: () => {
      toast({
        title: "Reset",
        description: "Reset initiated, system is reloading.",
      });
    },
    onError: () => {
      toast({
        title: "Error",
        description: "Failed to initiate reset.",
        variant: "destructive",
      });
    },
  });


  const setWhitelistEnabledMutation = useMutation({
    mutationFn: ({ subdomain, enabled }: { subdomain: string, enabled: boolean }) => api.setWhitelistEnabled(subdomain, enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['enabled-whitelists'] });
    },
  });

  const setBlacklistEnabledMutation = useMutation({
    mutationFn: ({ subdomain, enabled }: { subdomain: string, enabled: boolean }) => api.setBlacklistEnabled(subdomain, enabled),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['enabled-blacklists'] });
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
      if (!lb.subdomain || !lb.subdomain.trim()) {
        errors.push(`Load balancer ${lbIndex + 1}: Subdomain is required`);
      }

      // Type is set by switch, always valid

      // Validate VPS entries
      let totalCapacitySum = 0;
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

        totalCapacitySum += vps.capacity;

        if (vps.active) {
          hasActiveVPS = true;
          activeCapacitySum += vps.capacity;
        }
      });

      // Check total capacity sum
      if (Math.abs(totalCapacitySum - 1) > 0.001) {
        errors.push(`Load balancer ${lbIndex + 1}: Sum of all VPS capacities must be 1 (currently ${totalCapacitySum.toFixed(3)})`);
      }

      // Check capacity sum for active VPS
      if (hasActiveVPS && Math.abs(activeCapacitySum - 1) > 0.001) {
        errors.push(`Load balancer ${lbIndex + 1}: Sum of active VPS capacities must be 1 (currently ${activeCapacitySum.toFixed(3)})`);
      }

      // Check if there are any active VPS
      if (!hasActiveVPS) {
        errors.push(`Load balancer ${lbIndex + 1}: At least one VPS must be active`);
      }

      // Validate cache paths
      if (lb.cache_enabled) {
        const paths = lb.cache_paths || [];
        if (paths.length === 0) {
          errors.push(`Load balancer ${lbIndex + 1}: Cache enabled but no cache paths specified`);
        } else {
          paths.forEach((path, pathIndex) => {
            if (!path.trim()) {
              errors.push(`Load balancer ${lbIndex + 1}, Cache Path ${pathIndex + 1}: Path cannot be empty`);
            } else if (!path.startsWith('/')) {
              errors.push(`Load balancer ${lbIndex + 1}, Cache Path ${pathIndex + 1}: Path must start with '/'`);
            }
          });
        }
      }
    });

    // Validate root load balancer
    if (cfg.root_load_balancer) {
      // Type is set by switch, always valid

      // Validate VPS entries
      let totalCapacitySum = 0;
      let activeCapacitySum = 0;
      let hasActiveVPS = false;

      cfg.root_load_balancer.vps.forEach((vps, vpsIndex) => {
        // Validate IP
        if (!vps.ip.trim()) {
          errors.push(`Root Load Balancer, VPS ${vpsIndex + 1}: IP address is required`);
        } else {
          try {
            new URL(vps.ip);
          } catch {
            errors.push(`Root Load Balancer, VPS ${vpsIndex + 1}: Invalid IP/URL format`);
          }
        }

        // Validate capacity
        if (vps.capacity < 0 || vps.capacity > 1) {
          errors.push(`Root Load Balancer, VPS ${vpsIndex + 1}: Capacity must be between 0 and 1`);
        }

        totalCapacitySum += vps.capacity;

        if (vps.active) {
          hasActiveVPS = true;
          activeCapacitySum += vps.capacity;
        }
      });

      // Check total capacity sum
      if (Math.abs(totalCapacitySum - 1) > 0.001) {
        errors.push(`Root Load Balancer: Sum of all VPS capacities must be 1 (currently ${totalCapacitySum.toFixed(3)})`);
      }

      // Check capacity sum for active VPS
      if (hasActiveVPS && Math.abs(activeCapacitySum - 1) > 0.001) {
        errors.push(`Root Load Balancer: Sum of active VPS capacities must be 1 (currently ${activeCapacitySum.toFixed(3)})`);
      }

      // Check if there are any active VPS
      if (!hasActiveVPS) {
        errors.push(`Root Load Balancer: At least one VPS must be active`);
      }

      // Validate cache paths for root
      if (cfg.root_load_balancer.cache_enabled) {
        const paths = cfg.root_load_balancer.cache_paths || [];
        if (paths.length === 0) {
          errors.push(`Root Load Balancer: Cache enabled but no cache paths specified`);
        } else {
          paths.forEach((path, pathIndex) => {
            if (!path.trim()) {
              errors.push(`Root Load Balancer, Cache Path ${pathIndex + 1}: Path cannot be empty`);
            } else if (!path.startsWith('/')) {
              errors.push(`Root Load Balancer, Cache Path ${pathIndex + 1}: Path must start with '/'`);
            }
          });
        }
      }
    }

    return errors;
  };

  const handleSave = () => {
    if (!formData) return;

    const errors = validateConfig(formData);
    if (errors.length > 0) {
      setValidationErrors(errors);
      toast({
        title: "Validation Error",
        description: errors.join("; "),
        variant: "destructive",
      });
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
        cache_enabled: false,
        cache_paths: [],
        whitelist_enabled: false,
        blacklist_enabled: false,
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

  const updateCachePath = (lbIndex: number, pathIndex: number, value: string) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      const newPaths = [...(newLB[lbIndex].cache_paths || [])];
      newPaths[pathIndex] = value;
      newLB[lbIndex] = { ...newLB[lbIndex], cache_paths: newPaths };
      return { ...prev, load_balancer: newLB };
    });
  };

  const addCachePath = (lbIndex: number) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      const newPaths = [...(newLB[lbIndex].cache_paths || []), ""];
      newLB[lbIndex] = { ...newLB[lbIndex], cache_paths: newPaths };
      return { ...prev, load_balancer: newLB };
    });
  };

  const removeCachePath = (lbIndex: number, pathIndex: number) => {
    setFormData(prev => {
      if (!prev) return null;
      const newLB = [...prev.load_balancer];
      const newPaths = [...(newLB[lbIndex].cache_paths || [])];
      newPaths.splice(pathIndex, 1);
      newLB[lbIndex] = { ...newLB[lbIndex], cache_paths: newPaths };
      return { ...prev, load_balancer: newLB };
    });
  };

  const updateRootLB = (field: string, value: any) => {
    setFormData(prev => {
      if (!prev || !prev.root_load_balancer) return prev;
      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, [field]: value } };
    });
  };

  const updateRootVPS = (vpsIndex: number, field: string, value: any) => {
    setFormData(prev => {
      if (!prev || !prev.root_load_balancer) return prev;
      const newVPS = [...prev.root_load_balancer.vps];
      (newVPS[vpsIndex] as any)[field] = value;
      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, vps: newVPS } };
    });
  };

  const addRootVPS = () => {
    setFormData(prev => {
      if (!prev) return null;
      if (!prev.root_load_balancer) return prev;
      const newVPS = (prev.root_load_balancer.vps === null ? [] : [...prev.root_load_balancer.vps]);
      newVPS.push({ ip: "", capacity: 0.5, active: true });
      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, vps: newVPS } };
    });
  };

  const removeRootVPS = (vpsIndex: number) => {
    setFormData(prev => {
      if (!prev || !prev.root_load_balancer) return prev;
      const newVPS = [...prev.root_load_balancer.vps];
      newVPS.splice(vpsIndex, 1);
      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, vps: newVPS } };
    });
  };

  const addRootLoadBalancer = () => {
    setFormData(prev => {
      if (!prev) return null;
      return { ...prev, root_load_balancer: { vps: [{ ip: "", capacity: 1, active: true }], type: "https", active: true, cache_enabled: false, cache_paths: [], whitelist_enabled: false, blacklist_enabled: false } };
    });
  };

  const removeRootLoadBalancer = () => {
    setFormData(prev => {
      if (!prev) return null;
      const { root_load_balancer, ...rest } = prev;
      return rest;
    });
  };

  const updateRootCachePath = (pathIndex: number, value: string) => {
    setFormData(prev => {
      if (!prev || !prev.root_load_balancer) return prev;
      const newPaths = [...(prev.root_load_balancer.cache_paths || [])];
      newPaths[pathIndex] = value;
      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, cache_paths: newPaths } };
    });
  };

  const addRootCachePath = () => {
    setFormData(prev => {
      if (!prev || !prev.root_load_balancer) return prev;
      const newPaths = [...(prev.root_load_balancer.cache_paths || []), ""];
      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, cache_paths: newPaths } };
    });
  };

  const removeRootCachePath = (pathIndex: number) => {
    setFormData(prev => {
      if (!prev || !prev.root_load_balancer) return prev;
      const newPaths = [...(prev.root_load_balancer.cache_paths || [])];
      newPaths.splice(pathIndex, 1);
      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, cache_paths: newPaths } };
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
          <Button variant="outline" className="gap-2" onClick={() => { setFormData(config); reloadMutation.mutate(); }} disabled={reloadMutation.isPending}>
            <RefreshCw className="h-4 w-4" />
            {reloadMutation.isPending ? 'Resetting...' : 'Reset'}
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
              required
              id="hostname"
              value={formData.hostname}
              onChange={(e) => updateField('hostname', e.target.value)}
              className="bg-secondary border-border font-mono"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="subdomain_admin_panel">Admin Panel Subdomain</Label>
            <Input
              required
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

      {/* Root Domain Load Balancer */}
      <div className="rounded-xl border border-border bg-card p-6 space-y-6">
        <h2 className="text-lg font-semibold text-foreground">Root Domain Load Balancer</h2>

        {formData.root_load_balancer ? (
          <div className="space-y-4">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-medium">Root Domain</h3>
              <Button size="sm" variant="ghost" onClick={removeRootLoadBalancer}>
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
            <div className="grid gap-6 md:grid-cols-4">
              <div className="center-switch">
                <Label>HTTPS Enabled</Label>
                <Switch
                  checked={formData.root_load_balancer.type === "https"}
                  onCheckedChange={(checked) => updateRootLB('type', checked ? "https" : "http")}
                />
              </div>
              <div className="center-switch">
                <Label>Active</Label>
                <Switch
                  checked={formData.root_load_balancer.active}
                  onCheckedChange={(checked) => updateRootLB('active', checked)}
                />
              </div>
              <div className="center-switch">
                <Label>Whitelist</Label>
                <Switch
                  checked={formData?.root_load_balancer?.whitelist_enabled || false}
                  onCheckedChange={(checked) => {
                    setFormData(prev => {
                      if (!prev || !prev.root_load_balancer) return prev;
                      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, whitelist_enabled: checked, blacklist_enabled: checked ? false : prev.root_load_balancer.blacklist_enabled } };
                    });
                    if (checked) {
                      setBlacklistEnabledMutation.mutate({ subdomain: "", enabled: false });
                    }
                    setWhitelistEnabledMutation.mutate({ subdomain: "", enabled: checked });
                  }}
                />
              </div>
              <div className="center-switch">
                <Label>Blacklist</Label>
                <Switch
                  checked={formData?.root_load_balancer?.blacklist_enabled || false}
                  onCheckedChange={(checked) => {
                    setFormData(prev => {
                      if (!prev || !prev.root_load_balancer) return prev;
                      return { ...prev, root_load_balancer: { ...prev.root_load_balancer, blacklist_enabled: checked, whitelist_enabled: checked ? false : prev.root_load_balancer.whitelist_enabled } };
                    });
                    if (checked) {
                      setWhitelistEnabledMutation.mutate({ subdomain: "", enabled: false });
                    }
                    setBlacklistEnabledMutation.mutate({ subdomain: "", enabled: checked });
                  }}
                />
              </div>
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label>VPS Servers</Label>
                <Button size="sm" variant="outline" onClick={addRootVPS}>
                  <Plus className="h-4 w-4 mr-1" />
                  Add VPS
                </Button>
              </div>
              <div className="space-y-2">
                {formData.root_load_balancer.vps !== null && (formData.root_load_balancer.vps.map((vps, j) => (
                  <div key={j} className="flex gap-2 items-center">
                    <Input
                      required
                      placeholder="IP Address"
                      value={vps.ip}
                      onChange={(e) => updateRootVPS(j, 'ip', e.target.value)}
                      className="bg-secondary border-border font-mono flex-1"
                    />
                    <Input
                      type="number"
                      step="0.1"
                      placeholder="Capacity"
                      value={vps.capacity}
                      onChange={(e) => updateRootVPS(j, 'capacity', parseFloat(e.target.value))}
                      className="bg-secondary border-border font-mono w-24"
                    />
                    <Switch
                      checked={vps.active}
                      onCheckedChange={(checked) => updateRootVPS(j, 'active', checked)}
                    />
                    <Button size="sm" variant="ghost" onClick={() => removeRootVPS(j)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                )))}
              </div>
            </div>

            <div className="space-y-2">
              <div className="center-switch">
                <Label>Cache Enabled</Label>
                <Switch
                  checked={formData.root_load_balancer.cache_enabled}
                  onCheckedChange={(checked) => updateRootLB('cache_enabled', checked)}
                />
              </div>
              {formData.root_load_balancer.cache_enabled && (
                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label>Cache Paths</Label>
                    <Button size="sm" variant="outline" onClick={addRootCachePath}>
                      <Plus className="h-4 w-4 mr-1" />
                      Add Path
                    </Button>
                  </div>
                  <div className="space-y-2">
                    {(formData.root_load_balancer.cache_paths || []).map((path, k) => (
                      <div key={k} className="flex gap-2 items-center">
                        <Input
                          placeholder="/path or /path/*"
                          value={path}
                          onChange={(e) => updateRootCachePath(k, e.target.value)}
                          className="bg-secondary border-border font-mono flex-1"
                        />
                        <Button size="sm" variant="ghost" onClick={() => removeRootCachePath(k)}>
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        ) : (
          <Button variant="outline" onClick={addRootLoadBalancer}>
            <Plus className="h-4 w-4 mr-2" />
            Add Root Domain Load Balancer
          </Button>
        )}
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
              <div className="grid gap-4 md:grid-cols-5">
                <div className="space-y-2">
                  <Label>Subdomain</Label>
                  <Input
                    required
                    value={lb.subdomain}
                    onChange={(e) => updateLoadBalancer(i, 'subdomain', e.target.value)}
                    className="bg-secondary border-border font-mono"
                  />
                </div>
                <div className="center-switch">
                  <Label>HTTPS</Label>
                  <Switch
                    checked={lb.type === "https"}
                    onCheckedChange={(checked) => updateLoadBalancer(i, 'type', checked ? "https" : "http")}
                  />
                </div>
                <div className="center-switch">
                  <Label>Active</Label>
                  <Switch
                    checked={lb.active}
                    onCheckedChange={(checked) => updateLoadBalancer(i, 'active', checked)}
                  />
                </div>
                <div className="center-switch">
                  <Label>Whitelist</Label>
                  <Switch
                    checked={lb.whitelist_enabled}
                    onCheckedChange={(checked) => {
                      setFormData(prev => {
                        if (!prev) return null;
                        const newLB = prev.load_balancer.map(lbItem => {
                          if (lbItem.subdomain === lb.subdomain) {
                            return { ...lbItem, whitelist_enabled: checked, blacklist_enabled: checked ? false : lbItem.blacklist_enabled };
                          }
                          return lbItem;
                        });
                        return { ...prev, load_balancer: newLB };
                      });
                      if (checked) {
                        setBlacklistEnabledMutation.mutate({ subdomain: lb.subdomain, enabled: false });
                      }
                      setWhitelistEnabledMutation.mutate({ subdomain: lb.subdomain, enabled: checked });
                    }}
                  />
                </div>
                <div className="center-switch">
                  <Label>Blacklist</Label>
                  <Switch
                    checked={lb.blacklist_enabled}
                    onCheckedChange={(checked) => {
                      setFormData(prev => {
                        if (!prev) return null;
                        const newLB = prev.load_balancer.map(lbItem => {
                          if (lbItem.subdomain === lb.subdomain) {
                            return { ...lbItem, blacklist_enabled: checked, whitelist_enabled: checked ? false : lbItem.whitelist_enabled };
                          }
                          return lbItem;
                        });
                        return { ...prev, load_balancer: newLB };
                      });
                      if (checked) {
                        setWhitelistEnabledMutation.mutate({ subdomain: lb.subdomain, enabled: false });
                      }
                      setBlacklistEnabledMutation.mutate({ subdomain: lb.subdomain, enabled: checked });
                    }}
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
                        required
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

              <div className="space-y-2">
                <div className="center-switch">
                  <Label>Cache Enabled</Label>
                  <Switch
                    checked={lb.cache_enabled}
                    onCheckedChange={(checked) => updateLoadBalancer(i, 'cache_enabled', checked)}
                  />
                </div>
                {lb.cache_enabled && (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <Label>Cache Paths</Label>
                      <Button size="sm" variant="outline" onClick={() => addCachePath(i)}>
                        <Plus className="h-4 w-4 mr-1" />
                        Add Path
                      </Button>
                    </div>
                    <div className="space-y-2">
                      {(lb.cache_paths || []).map((path, k) => (
                        <div key={k} className="flex gap-2 items-center">
                          <Input
                            placeholder="/path or /path/*"
                            value={path}
                            onChange={(e) => updateCachePath(i, k, e.target.value)}
                            className="bg-secondary border-border font-mono flex-1"
                          />
                          <Button size="sm" variant="ghost" onClick={() => removeCachePath(i, k)}>
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
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
