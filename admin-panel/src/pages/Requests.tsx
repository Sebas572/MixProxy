import { useEffect, useState } from 'react';

interface Request {
  id: string;
  method: string;
  url: string;
  ip: string;
  timestamp: string;
  status: number;
}

const Requests = () => {
  const [requests, setRequests] = useState<Request[]>([]);

  useEffect(() => {
    const fetchRequests = () => {
      fetch('https://admin-api.developer.space/api/requests')
        .then(res => res.json())
        .then(setRequests)
        .catch(console.error);
    };

    fetchRequests();
    const interval = setInterval(fetchRequests, 2000);

    return () => clearInterval(interval);
  }, []);

  return (
    <div className="requests">
      <h2>Real-time Requests</h2>
      <table>
        <thead>
          <tr>
            <th><p>Time</p></th>
            <th><p>Method</p></th>
            <th><p>URL</p></th>
            <th><p>IP</p></th>
            <th><p>Status</p></th>
          </tr>
        </thead>
        <tbody>
          {requests.map(req => (
            <tr key={req.id}>
              <td>{new Date(req.timestamp).toLocaleTimeString()}</td>
              <td>{req.method}</td>
              <td>{req.url}</td>
              <td>{req.ip}</td>
              <td>{req.status}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Requests;