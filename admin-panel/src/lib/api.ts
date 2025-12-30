const getHostname = () => {
  return window.location.hostname.split(".").slice(1).join(".");
};

const API_BASE = `https://admin-api.${getHostname()}`;

export interface Stats {
  totalRequests: number;
  activeConnections: number;
  uniqueIPs: number;
}

export interface RequestLog {
  id: string;
  method: string;
  url: string;
  ip: string;
  timestamp: string;
  status: number;
}

export interface IPStat {
  ip: string;
  count: number;
  lastSeen: string;
}

export interface VPSEntry {
  ip: string;
  capacity: number;
  active: boolean;
}

export interface LoadBalancerEntry {
  vps: VPSEntry[];
  type: string;
  subdomain?: string;
  active: boolean;
  cache_enabled: boolean;
  cache_paths: string[];
  whitelist_enabled: boolean;
  blacklist_enabled: boolean;
}

export interface Reason {
  Content: string;
  Time: string;
  Date: string;
}

export interface Config {
  hostname: string;
  subdomain_admin_panel: string;
  on_https: boolean;
  mode_developer: boolean;
  load_balancer: LoadBalancerEntry[];
  root_load_balancer?: LoadBalancerEntry;
}

export const api = {
  async getStats(): Promise<Stats> {
    const res = await fetch(`${API_BASE}/api/stats`);
    if (!res.ok) throw new Error('Failed to fetch stats');
    return res.json();
  },

  async getRequests(): Promise<RequestLog[]> {
    const res = await fetch(`${API_BASE}/api/requests`);
    if (!res.ok) throw new Error('Failed to fetch requests');
    return res.json();
  },

  async getIPs(): Promise<IPStat[]> {
    const res = await fetch(`${API_BASE}/api/ips`);
    if (!res.ok) throw new Error('Failed to fetch IPs');
    return res.json();
  },

  async getLogs(date?: string): Promise<string> {
    const url = date ? `${API_BASE}/api/logs?date=${date}` : `${API_BASE}/api/logs`;
    const res = await fetch(url);
    if (!res.ok) throw new Error('Failed to fetch logs');
    return res.text();
  },

  async getLogList(): Promise<string[]> {
    const res = await fetch(`${API_BASE}/api/logs/list`);
    if (!res.ok) throw new Error('Failed to fetch log list');
    const logListSort = (await res.json()).sort((a, b) => {
      const d1 = new Date(a.split("-").slice(1).join("/"));
      const d2 = new Date(b.split("-").slice(1).join("/"));

      return d2.getTime() - d1.getTime();
    })

    return logListSort
  },

  async getConfig(): Promise<Config> {
    const res = await fetch(`${API_BASE}/api/config`);
    if (!res.ok) throw new Error('Failed to fetch config');
    return res.json();
  },

  async updateConfig(config: Config): Promise<void> {
    const res = await fetch(`${API_BASE}/api/config`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(config),
    });
    if (!res.ok) throw new Error('Failed to update config');
  },

  async reload(): Promise<void> {
    const res = await fetch(`${API_BASE}/api/reload`, { method: 'POST' });
    if (!res.ok) throw new Error('Failed to reload');
  },

  // Whitelist
  async getWhitelistEnabled(subdomain: string): Promise<boolean> {
    const res = await fetch(`${API_BASE}/api/whitelist/enabled/${subdomain}`);
    if (!res.ok) throw new Error('Failed to get whitelist enabled');
    const data = await res.json();
    return data.enabled;
  },

  async setWhitelistEnabled(subdomain: string, enabled: boolean): Promise<void> {
    const res = await fetch(`${API_BASE}/api/whitelist/enabled/${subdomain}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
    });
    if (!res.ok) throw new Error('Failed to set whitelist enabled');
  },

  async getEnabledWhitelists(): Promise<string[]> {
    const res = await fetch(`${API_BASE}/api/whitelist/enabled`);
    if (!res.ok) throw new Error('Failed to get enabled whitelists');
    return res.json();
  },

  async getWhitelistIPs(subdomain: string): Promise<Record<string, Reason>> {
    const res = await fetch(`${API_BASE}/api/whitelist/ips/${subdomain}`);
    if (!res.ok) throw new Error('Failed to get whitelist IPs');
    return res.json();
  },

  async addWhitelistIP(subdomain: string, ip: string, reason: Reason, duration: string): Promise<void> {
    const res = await fetch(`${API_BASE}/api/whitelist/ip`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ subdomain, ip, reason, duration }),
    });
    if (!res.ok) throw new Error('Failed to add whitelist IP');
  },

  async removeWhitelistIP(subdomain: string, ip: string): Promise<void> {
    const res = await fetch(`${API_BASE}/api/whitelist/ip/${subdomain}/${ip}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to remove whitelist IP');
  },

  async getEnabledBlacklists(): Promise<string[]> {
    const res = await fetch(`${API_BASE}/api/blacklist/enabled`);
    if (!res.ok) throw new Error('Failed to get enabled blacklists');
    return res.json();
  },

  // Blacklist
  async getBlacklistEnabled(subdomain: string): Promise<boolean> {
    const res = await fetch(`${API_BASE}/api/blacklist/enabled/${subdomain}`);
    if (!res.ok) throw new Error('Failed to get blacklist enabled');
    const data = await res.json();
    return data.enabled;
  },

  async setBlacklistEnabled(subdomain: string, enabled: boolean): Promise<void> {
    const res = await fetch(`${API_BASE}/api/blacklist/enabled/${subdomain}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled }),
    });
    if (!res.ok) throw new Error('Failed to set blacklist enabled');
  },

  async getBlacklistIPs(subdomain: string): Promise<Record<string, Reason>> {
    const res = await fetch(`${API_BASE}/api/blacklist/ips/${subdomain}`);
    if (!res.ok) throw new Error('Failed to get blacklist IPs');
    return res.json();
  },

  async addBlacklistIP(subdomain: string, ip: string, reason: Reason, duration: string): Promise<void> {
    const res = await fetch(`${API_BASE}/api/blacklist/ip`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ subdomain, ip, reason, duration }),
    });
    if (!res.ok) throw new Error('Failed to add blacklist IP');
  },

  async removeBlacklistIP(subdomain: string, ip: string): Promise<void> {
    const res = await fetch(`${API_BASE}/api/blacklist/ip/${subdomain}/${ip}`, { method: 'DELETE' });
    if (!res.ok) throw new Error('Failed to remove blacklist IP');
  },
};