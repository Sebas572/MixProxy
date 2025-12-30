import { NavLink, useLocation } from "react-router-dom";
import { LayoutDashboard, Activity, Globe, Settings, Shield, ShieldCheck, Ban } from "lucide-react";
import { cn } from "@/lib/utils";

const navItems = [
  { to: "/", icon: LayoutDashboard, label: "Dashboard" },
  { to: "/requests", icon: Activity, label: "Requests" },
  { to: "/ips", icon: Globe, label: "IPs" },
  { to: "/whitelist", icon: ShieldCheck, label: "Whitelist" },
  { to: "/blacklist", icon: Ban, label: "Blacklist" },
  { to: "/configuration", icon: Settings, label: "Configuration" },
];

export function Sidebar() {
  const location = useLocation();

  return (
    <aside className="fixed left-0 top-0 z-40 h-screen w-64 border-r border-sidebar-border bg-sidebar">
      <div className="flex h-full flex-col">
        {/* Logo */}
        <div className="flex h-16 items-center gap-3 border-b border-sidebar-border px-6">
          <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10 glow-primary">
            <Shield className="h-5 w-5 text-primary" />
          </div>
          <div>
            <h1 className="text-sm font-semibold text-foreground">MixProxy</h1>
            <p className="text-xs text-muted-foreground">Admin Panel</p>
          </div>
        </div>

        {/* Navigation */}
        <nav className="flex-1 space-y-1 p-4">
          {navItems.map((item) => {
            const isActive = location.pathname === item.to;
            return (
              <NavLink
                key={item.to}
                to={item.to}
                className={cn(
                  "flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200",
                  isActive
                    ? "bg-primary/10 text-primary glow-primary"
                    : "text-sidebar-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
                )}
              >
                <item.icon className={cn("h-5 w-5", isActive && "text-primary")} />
                {item.label}
              </NavLink>
            );
          })}
        </nav>

        {/* Footer */}
        <div className="border-t border-sidebar-border p-4">
          <div className="flex items-center gap-2">
            <div className="h-2 w-2 rounded-full bg-success animate-pulse" />
            <span className="text-xs text-muted-foreground">System Online</span>
          </div>
        </div>
      </div>
    </aside>
  );
}
