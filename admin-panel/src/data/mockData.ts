export const dashboardStats = {
  totalRequests: 1247893,
  activeConnections: 342,
  uniqueIps: 8924,
};

export const requestsData = [
  { time: "2024-01-15 14:23:45", method: "GET", url: "/api/users", ip: "192.168.1.105", status: 200 },
  { time: "2024-01-15 14:23:44", method: "POST", url: "/api/auth/login", ip: "45.33.32.156", status: 401 },
  { time: "2024-01-15 14:23:43", method: "GET", url: "/api/products", ip: "104.26.10.78", status: 200 },
  { time: "2024-01-15 14:23:42", method: "DELETE", url: "/api/users/123", ip: "172.67.74.152", status: 403 },
  { time: "2024-01-15 14:23:41", method: "PUT", url: "/api/settings", ip: "31.13.24.53", status: 200 },
  { time: "2024-01-15 14:23:40", method: "GET", url: "/api/analytics", ip: "157.240.1.35", status: 500 },
  { time: "2024-01-15 14:23:39", method: "POST", url: "/api/webhook", ip: "185.199.108.153", status: 200 },
  { time: "2024-01-15 14:23:38", method: "GET", url: "/health", ip: "34.117.59.81", status: 200 },
  { time: "2024-01-15 14:23:37", method: "PATCH", url: "/api/users/456", ip: "151.101.1.140", status: 200 },
  { time: "2024-01-15 14:23:36", method: "GET", url: "/api/cache/clear", ip: "104.18.32.7", status: 302 },
];

export const ipsData = [
  { ip: "192.168.1.105", requestCount: 15420, lastSeen: "2024-01-15 14:23:45" },
  { ip: "45.33.32.156", requestCount: 8932, lastSeen: "2024-01-15 14:23:44" },
  { ip: "104.26.10.78", requestCount: 7541, lastSeen: "2024-01-15 14:23:43" },
  { ip: "172.67.74.152", requestCount: 6234, lastSeen: "2024-01-15 14:23:42" },
  { ip: "31.13.24.53", requestCount: 5102, lastSeen: "2024-01-15 14:23:41" },
  { ip: "157.240.1.35", requestCount: 4521, lastSeen: "2024-01-15 14:23:40" },
  { ip: "185.199.108.153", requestCount: 3892, lastSeen: "2024-01-15 14:23:39" },
  { ip: "34.117.59.81", requestCount: 3241, lastSeen: "2024-01-15 14:23:38" },
  { ip: "151.101.1.140", requestCount: 2845, lastSeen: "2024-01-15 14:23:37" },
  { ip: "104.18.32.7", requestCount: 2156, lastSeen: "2024-01-15 14:23:36" },
];

export const trafficOriginData = [
  { country: "United States", lat: 37.0902, lng: -95.7129, requests: 425000 },
  { country: "Germany", lat: 51.1657, lng: 10.4515, requests: 189000 },
  { country: "Japan", lat: 36.2048, lng: 138.2529, requests: 156000 },
  { country: "Brazil", lat: -14.235, lng: -51.9253, requests: 134000 },
  { country: "United Kingdom", lat: 55.3781, lng: -3.436, requests: 98000 },
  { country: "India", lat: 20.5937, lng: 78.9629, requests: 87000 },
  { country: "Australia", lat: -25.2744, lng: 133.7751, requests: 65000 },
  { country: "Canada", lat: 56.1304, lng: -106.3468, requests: 54000 },
];
