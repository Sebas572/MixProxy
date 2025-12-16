import { useEffect, useState } from 'react';

interface LoadBalancerEntry {
  vps: { ip: string; capacity: number; active: boolean }[];
  type: string;
  subdomain: string;
  active: boolean;
}

interface Config {
  hostname: string;
  on_https: boolean;
  mode_developer: boolean;
  load_balancer: LoadBalancerEntry[];
}

const ConfigPage = () => {
  const [config, setConfig] = useState<Config | null>(null);

  useEffect(() => {
    fetch('https://admin-api.developer.space/api/config')
      .then(res => res.json())
      .then(setConfig);
  }, []);

  if (!config) return <div>Loading...</div>;

  return (
    <div className="config">
      <h2>Configuration</h2>
      <pre>{JSON.stringify(config, null, 2)}</pre>
    </div>
  );
};

export default ConfigPage;