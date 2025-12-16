import { useEffect, useState } from 'react';

interface Stats {
  totalRequests: number;
  activeConnections: number;
  uniqueIPs: number;
}

const Dashboard = () => {
  const [stats, setStats] = useState<Stats>({ totalRequests: 0, activeConnections: 0, uniqueIPs: 0 });

  useEffect(() => {
    const fetchStats = () => {
      fetch('https://admin-api.developer.space/api/stats')
        .then(res => res.json())
        .then(setStats)
        .catch(console.error);
    };

    fetchStats();
    const interval = setInterval(fetchStats, 2000); // Poll every 2 seconds

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="dashboard">
      <h2>Dashboard</h2>
      <div className="stats">
        <div className="stat-card">
          <h3>Total Requests</h3>
          <p>{stats.totalRequests}</p>
        </div>
        <div className="stat-card">
          <h3>Active Connections</h3>
          <p>{stats.activeConnections}</p>
        </div>
        <div className="stat-card">
          <h3>Unique IPs</h3>
          <p>{stats.uniqueIPs}</p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;