import React, { useState } from "react";
import { Plus, Trash2, ShieldCheck } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, Reason } from "@/lib/api";
import { useToast } from "@/hooks/use-toast";

export default function Whitelist() {
  const queryClient = useQueryClient();
  const { toast } = useToast();
  const [selectedSubdomain, setSelectedSubdomain] = useState<string>("");
  const [newIP, setNewIP] = useState("");
  const [newReason, setNewReason] = useState("");
  const [newDuration, setNewDuration] = useState("1h");

  const { data: enabledSubdomains } = useQuery({
    queryKey: ['enabled-whitelists'],
    queryFn: api.getEnabledWhitelists,
  });

  const { data: ips } = useQuery({
    queryKey: ['whitelist-ips', selectedSubdomain],
    queryFn: () => api.getWhitelistIPs(selectedSubdomain),
    enabled: !!selectedSubdomain,
  });


  const addIPMutation = useMutation({
    mutationFn: () => {
      const reason: Reason = {
        Content: newReason,
        Time: new Date().toISOString(),
        Date: new Date().toISOString().split('T')[0],
      };
      return api.addWhitelistIP(selectedSubdomain, newIP, reason, newDuration);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['whitelist-ips', selectedSubdomain] });
      setNewIP("");
      setNewReason("");
      setNewDuration("1h");
      toast({
        title: "Success",
        description: "IP added to whitelist",
      });
    },
    onError: () => {
      toast({
        title: "Error",
        description: "Failed to add IP",
        variant: "destructive",
      });
    },
  });

  const removeIPMutation = useMutation({
    mutationFn: (ip: string) => api.removeWhitelistIP(selectedSubdomain, ip),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['whitelist-ips', selectedSubdomain] });
      toast({
        title: "Success",
        description: "IP removed from whitelist",
      });
    },
    onError: () => {
      toast({
        title: "Error",
        description: "Failed to remove IP",
        variant: "destructive",
      });
    },
  });

  return (
    <div className="space-y-8">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-foreground">Whitelist</h1>
          <p className="text-sm text-muted-foreground">
            Manage whitelisted IPs per subdomain
          </p>
        </div>
      </div>

      {/* Enabled Subdomains List */}
      <div className="rounded-xl border border-border bg-card p-6 space-y-4">
        <h2 className="text-lg font-semibold text-foreground">Enabled Whitelists</h2>
        <div className="grid gap-2 md:grid-cols-3">
          {enabledSubdomains?.map((subdomain) => (
            <Button
              key={subdomain}
              variant={selectedSubdomain === subdomain ? "default" : "outline"}
              onClick={() => setSelectedSubdomain(subdomain)}
              className="justify-start"
            >
              <ShieldCheck className="h-4 w-4 mr-2" />
              {subdomain}
            </Button>
          ))}
          {(!enabledSubdomains || enabledSubdomains.length === 0) && (
            <div className="text-center text-muted-foreground py-8 col-span-3">
              No whitelists enabled
            </div>
          )}
        </div>
      </div>

      {selectedSubdomain && (
        <>
          {/* Add IP */}
          <div className="rounded-xl border border-border bg-card p-6 space-y-4">
            <h2 className="text-lg font-semibold text-foreground">Add IP to Whitelist for {selectedSubdomain}</h2>
            <div className="grid gap-4 md:grid-cols-4">
              <div className="space-y-2">
                <Label>IP Address</Label>
                <Input
                  value={newIP}
                  onChange={(e) => setNewIP(e.target.value)}
                  placeholder="192.168.1.1"
                  className="bg-secondary border-border font-mono"
                />
              </div>
              <div className="space-y-2">
                <Label>Reason</Label>
                <Input
                  value={newReason}
                  onChange={(e) => setNewReason(e.target.value)}
                  placeholder="Reason for whitelisting"
                  className="bg-secondary border-border"
                />
              </div>
              <div className="space-y-2">
                <Label>Duration</Label>
                <Input
                  value={newDuration}
                  onChange={(e) => setNewDuration(e.target.value)}
                  placeholder="1h"
                  className="bg-secondary border-border font-mono"
                />
              </div>
              <div className="flex items-end">
                <Button
                  onClick={() => addIPMutation.mutate()}
                  disabled={addIPMutation.isPending || !newIP || !newReason}
                  className="w-full"
                >
                  <Plus className="h-4 w-4 mr-2" />
                  Add IP
                </Button>
              </div>
            </div>
          </div>

          {/* IP Table */}
          <div className="rounded-xl border border-border bg-card p-6 space-y-4">
            <h2 className="text-lg font-semibold text-foreground">Whitelisted IPs for {selectedSubdomain}</h2>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>IP Address</TableHead>
                  <TableHead>Reason</TableHead>
                  <TableHead>Added</TableHead>
                  <TableHead>Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {ips && Object.entries(ips).map(([ip, reason]) => (
                  <TableRow key={ip}>
                    <TableCell className="font-mono">{ip}</TableCell>
                    <TableCell>{reason.Content}</TableCell>
                    <TableCell>{new Date(reason.Time).toLocaleString()}+</TableCell>
                    <TableCell>
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => removeIPMutation.mutate(ip)}
                        disabled={removeIPMutation.isPending}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
                {(!ips || Object.keys(ips).length === 0) && (
                  <TableRow>
                    <TableCell colSpan={4} className="text-center text-muted-foreground py-8">
                      No whitelisted IPs for this subdomain
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>
        </>
      )}
    </div>
  );
}