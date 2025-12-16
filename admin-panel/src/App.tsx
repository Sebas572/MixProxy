import { BrowserRouter as Router, Routes, Route, Link } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Requests from './pages/Requests';
import IPs from './pages/IPs';
import Config from './pages/Config';
import './App.css';

function App() {
  return (
    <Router>
      <div className="app">
        <nav className="navbar">
          <h1>Proxy Monitor</h1>
          <ul>
            <li><Link to="/">Dashboard</Link></li>
            <li><Link to="/requests">Requests</Link></li>
            <li><Link to="/ips">IPs</Link></li>
            <li><Link to="/config">Configuration</Link></li>
          </ul>
        </nav>
        <main>
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/requests" element={<Requests />} />
            <Route path="/ips" element={<IPs />} />
            <Route path="/config" element={<Config />} />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
