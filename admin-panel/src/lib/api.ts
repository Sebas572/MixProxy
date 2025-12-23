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
};