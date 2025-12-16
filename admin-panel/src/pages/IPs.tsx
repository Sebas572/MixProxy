import { useEffect, useState } from 'react';

interface IPData {
  ip: string;
  count: number;
  lastSeen: string;
}

const IPs = () => {
  const [ips, setIPs] = useState<IPData[]>([]);

  useEffect(() => {
    const fetchIPs = () => {
      fetch('https://admin-api.developer.space/api/ips')
        .then(res => res.json())
        .then(setIPs)
        .catch(console.error);
    };

    fetchIPs();
    const interval = setInterval(fetchIPs, 2000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="ips">
      <h2>IP Monitoring</h2>
      <table>
        <thead>
          <tr>
            <th><p>IP Address</p></th>
            <th><p>Request Count</p></th>
            <th><p>Last Seen</p></th>
          </tr>
        </thead>
        <tbody>
          {ips.map(ip => (
            <tr key={ip.ip}>
              <td>{ip.ip}</td>
              <td>{ip.count}</td>
              <td>{new Date(ip.lastSeen).toLocaleString()}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default IPs;