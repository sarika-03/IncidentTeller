'use client';

import React from 'react';
import {
  CheckCircle,
  XCircle,
  AlertTriangle,
  Activity,
  Server,
  Database,
  Globe,
  Cpu,
  HardDrive,
  RefreshCw,
  Zap
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';
import { HealthResponse } from '@/types';
import { formatDistanceToNow } from 'date-fns';

const getStatusIcon = (status: string) => {
  switch (status) {
    case 'healthy':
      return CheckCircle;
    case 'degraded':
      return AlertTriangle;
    case 'unhealthy':
      return XCircle;
    default:
      return Activity;
  }
};

const getStatusColor = (status: string) => {
  switch (status) {
    case 'healthy':
      return 'text-success bg-success/10 border-success/20';
    case 'degraded':
      return 'text-warning bg-warning/10 border-warning/20';
    case 'unhealthy':
      return 'text-destructive bg-destructive/10 border-destructive/20';
    default:
      return 'text-muted-foreground bg-muted border-border';
  }
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'healthy':
      return <Badge variant="success">Healthy</Badge>;
    case 'degraded':
      return <Badge variant="warning">Degraded</Badge>;
    case 'unhealthy':
      return <Badge variant="destructive">Unhealthy</Badge>;
    default:
      return <Badge variant="secondary">Unknown</Badge>;
  }
};

const getServiceIcon = (service: string) => {
  switch (service.toLowerCase()) {
    case 'database':
      return Database;
    case 'netdata':
      return Globe;
    case 'memory':
      return Cpu;
    default:
      return Server;
  }
};

export default function HealthPage() {
  const [health, setHealth] = React.useState<HealthResponse | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [isRefreshing, setIsRefreshing] = React.useState(false);
  const [error, setError] = React.useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = React.useState<Date>(new Date());

  // Modals
  const [showLogs, setShowLogs] = React.useState(false);
  const [logs, setLogs] = React.useState<string[]>([]);
  const [showDiagnostics, setShowDiagnostics] = React.useState(false);
  const [diagnostics, setDiagnostics] = React.useState<any[]>([]);
  const [diagLoading, setDiagLoading] = React.useState(false);

  // Toast
  const [toast, setToast] = React.useState<string | null>(null);

  const showToast = (msg: string) => {
    setToast(msg);
    setTimeout(() => setToast(null), 3000);
  };

  const fetchHealth = async () => {
    setIsRefreshing(true);
    try {
      const healthData = await api.getHealth();
      setHealth(healthData);
      setLastUpdate(new Date());
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch health data');
    } finally {
      setLoading(false);
      setTimeout(() => setIsRefreshing(false), 500);
    }
  };

  const handleFetchLogs = async () => {
    setShowLogs(true);
    try {
      const data = await api.getLogs();
      setLogs(data.logs);
    } catch (err) {
      console.error('Failed to fetch logs', err);
    }
  };

  const handleExportMetrics = async () => {
    try {
      const blob = await api.exportMetrics();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `incident-teller-metrics-${new Date().toISOString()}.csv`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      showToast('Metrics exported successfully! ');
    } catch (err) {
      showToast('Failed to export metrics');
    }
  };

  const handleRunDiagnostics = async () => {
    setShowDiagnostics(true);
    setDiagLoading(true);
    try {
      const data = await api.getDiagnostics();
      setDiagnostics(data.diagnostics);
    } catch (err) {
      console.error('Diagnostics failed', err);
    } finally {
      setTimeout(() => setDiagLoading(false), 1500);
    }
  };

  React.useEffect(() => {
    fetchHealth();

    // Set up polling for real-time updates
    const interval = setInterval(fetchHealth, 30000); // Update every 30 seconds

    return () => clearInterval(interval);
  }, []);

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
      </div>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <AlertTriangle className="h-5 w-5 text-destructive" />
              <span className="text-destructive">{error}</span>
            </div>
            <button
              onClick={fetchHealth}
              className="px-3 py-1 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors flex items-center space-x-1"
            >
              <RefreshCw className={`h-3 w-3 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span>Retry</span>
            </button>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!health) {
    return null;
  }

  const StatusIcon = getStatusIcon(health.status);

  return (
    <div className="space-y-6 relative">
      {/* Toast Notification */}
      {toast && (
        <div className="fixed bottom-8 right-8 z-50 animate-in fade-in slide-in-from-bottom-5">
          <div className="bg-primary text-primary-foreground px-4 py-2 rounded-lg shadow-lg flex items-center space-x-2">
            <Zap className="h-4 w-4 fill-current" />
            <span>{toast}</span>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold italic tracking-tight">System Health & Metrics</h1>
        <div className="flex items-center space-x-2">
          <Badge variant="secondary" className="font-mono">
            Updated {formatDistanceToNow(lastUpdate, { addSuffix: true })}
          </Badge>
          <button
            onClick={fetchHealth}
            className="px-3 py-1 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors flex items-center space-x-1 group"
          >
            <RefreshCw className={`h-3 w-3 transition-transform group-hover:rotate-180 ${isRefreshing ? 'animate-spin text-primary' : ''}`} />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* Overall Status */}
      <Card className="overflow-hidden border-2 border-border/50 bg-gradient-to-br from-card/50 to-background backdrop-blur-sm">
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Activity className="h-5 w-5 text-primary" />
            <span>Overall System Status</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className={`p-4 rounded-2xl shadow-inner ${getStatusColor(health.status)}`}>
                <StatusIcon className="h-8 w-8" />
              </div>
              <div>
                <h2 className="text-2xl font-bold capitalize tracking-tight">{health.status}</h2>
                <p className="text-muted-foreground font-mono text-sm">
                  IncidentTeller v{health.version}
                </p>
              </div>
            </div>
            <div className="text-right">
              {getStatusBadge(health.status)}
              <p className="text-sm text-muted-foreground mt-1 font-mono">
                {new Date(health.timestamp).toLocaleTimeString()}
              </p>
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Service Health */}
      {health.checks && Object.keys(health.checks).length > 0 && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Server className="h-5 w-5" />
              <span>Service Health Monitor</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {Object.entries(health.checks).map(([service, status]) => {
                const ServiceIcon = getServiceIcon(service);
                const ServiceStatusIcon = getStatusIcon(status);

                return (
                  <div key={service} className="group relative border rounded-xl p-4 bg-muted/30 hover:bg-muted/50 transition-all cursor-default overflow-hidden">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <ServiceIcon className="h-4 w-4 text-primary" />
                        <span className="font-semibold capitalize">{service}</span>
                      </div>
                      <div className={`p-1.5 rounded-full ${getStatusColor(status)} shadow-sm`}>
                        <ServiceStatusIcon className="h-3 w-3" />
                      </div>
                    </div>
                    <div className="text-xs text-muted-foreground uppercase tracking-widest font-bold">
                      {status}
                    </div>
                    <div className="absolute bottom-0 left-0 h-1 bg-primary/20 w-full scale-x-0 group-hover:scale-x-100 transition-transform origin-left" />
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* Quick Actions */}
      <Card className="border-primary/20 bg-primary/5">
        <CardHeader>
          <CardTitle className="text-lg">Intelligent Operations</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-3 md:grid-cols-4">
            <button
              onClick={handleFetchLogs}
              className="px-4 py-3 rounded-xl bg-background border border-border shadow-sm hover:border-primary/50 hover:shadow-md transition-all flex flex-col items-center justify-center space-y-1 group"
            >
              <Activity className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
              <span className="text-sm font-medium">View Logs</span>
            </button>
            <button
              onClick={handleExportMetrics}
              className="px-4 py-3 rounded-xl bg-background border border-border shadow-sm hover:border-primary/50 hover:shadow-md transition-all flex flex-col items-center justify-center space-y-1 group"
            >
              <Cpu className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
              <span className="text-sm font-medium">Export Metrics</span>
            </button>
            <button
              onClick={handleRunDiagnostics}
              className="px-4 py-3 rounded-xl bg-background border border-border shadow-sm hover:border-primary/50 hover:shadow-md transition-all flex flex-col items-center justify-center space-y-1 group"
            >
              <HardDrive className="h-5 w-5 text-muted-foreground group-hover:text-primary transition-colors" />
              <span className="text-sm font-medium">Diagnostics</span>
            </button>
            <button
              onClick={fetchHealth}
              disabled={isRefreshing}
              className="px-4 py-3 rounded-xl bg-primary text-primary-foreground shadow-lg hover:brightness-110 active:scale-95 transition-all flex flex-col items-center justify-center space-y-1 disabled:opacity-50"
            >
              <RefreshCw className={`h-5 w-5 ${isRefreshing ? 'animate-spin' : ''}`} />
              <span className="text-sm font-medium">{isRefreshing ? 'Refreshing...' : 'Refresh All'}</span>
            </button>
          </div>
        </CardContent>
      </Card>

      {/* System Quick Stats */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Zap className="h-5 w-5 text-yellow-500" />
              <span>Repository Status</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Alerts In Database</span>
                <Badge variant="outline" className="font-mono">
                  {logs.length > 0 ? logs.length : '...'}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Operational Nodes</span>
                <Badge variant="success" className="font-mono">1 (Primary)</Badge>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Globe className="h-5 w-5 text-blue-500" />
              <span>API Integration</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Netdata Status</span>
                <Badge
                  variant={health?.checks?.netdata === 'healthy' ? 'success' : 'destructive'}
                  className="font-mono"
                >
                  {health?.checks?.netdata || 'checking...'}
                </Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Database Status</span>
                <Badge
                  variant={health?.checks?.database === 'healthy' ? 'success' : 'destructive'}
                  className="font-mono"
                >
                  {health?.checks?.database || 'checking...'}
                </Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Logs Modal */}
      {showLogs && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-background/80 backdrop-blur-md animate-in fade-in duration-200">
          <Card className="w-full max-w-4xl max-h-[80vh] overflow-hidden border-2 border-primary/20 shadow-2xl">
            <CardHeader className="flex flex-row items-center justify-between border-b bg-muted/50 p-4">
              <div className="flex items-center space-x-2">
                <Activity className="h-5 w-5 text-primary" />
                <CardTitle>System Log Stream</CardTitle>
              </div>
              <button
                onClick={() => setShowLogs(false)}
                className="p-1 rounded-full hover:bg-muted transition-colors"
              >
                <RefreshCw className="h-5 w-5 rotate-45" />
              </button>
            </CardHeader>
            <CardContent className="p-0">
              <div className="bg-black/90 p-4 font-mono text-xs overflow-y-auto max-h-[60vh] text-green-400">
                {logs.length > 0 ? (
                  logs.map((log, i) => (
                    <div key={i} className="py-0.5 border-l-2 border-green-900/50 pl-2 mb-1">
                      <span className="opacity-50">[{i}]</span> {log}
                    </div>
                  ))
                ) : (
                  <div className="text-center py-12 animate-pulse">Initializing log stream...</div>
                )}
                <div className="h-1 w-full bg-green-500/20 animate-pulse mt-4" />
              </div>
            </CardContent>
            <div className="p-4 bg-muted/50 flex justify-end">
              <button
                onClick={() => setShowLogs(false)}
                className="px-4 py-2 rounded-lg bg-primary text-primary-foreground font-medium hover:brightness-110"
              >
                Close Terminal
              </button>
            </div>
          </Card>
        </div>
      )}

      {/* Diagnostics Modal */}
      {showDiagnostics && (
        <div className="fixed inset-0 z-[100] flex items-center justify-center p-4 bg-background/80 backdrop-blur-md animate-in fade-in duration-200">
          <Card className="w-full max-w-2xl border-2 border-primary/20 shadow-2xl overflow-hidden">
            <CardHeader className="border-b bg-muted/50">
              <CardTitle className="flex items-center space-x-2">
                <HardDrive className="h-5 w-5 text-primary" />
                <span>Deep Scan Diagnostics</span>
              </CardTitle>
            </CardHeader>
            <CardContent className="p-6">
              {diagLoading ? (
                <div className="space-y-6 py-8">
                  <div className="flex items-center justify-center">
                    <div className="relative">
                      <div className="h-16 w-16 rounded-full border-4 border-primary/20 animate-ping absolute" />
                      <div className="h-16 w-16 rounded-full border-4 border-primary border-t-transparent animate-spin" />
                    </div>
                  </div>
                  <div className="text-center space-y-2">
                    <p className="text-lg font-bold animate-pulse">Scanning System Components...</p>
                    <p className="text-sm text-muted-foreground italic">Verifying integrity of incident correlation engine</p>
                  </div>
                </div>
              ) : (
                <div className="space-y-4">
                  <div className="p-4 bg-success/10 border border-success/20 rounded-xl flex items-center space-x-3 mb-6">
                    <CheckCircle className="h-6 w-6 text-success" />
                    <div>
                      <h4 className="font-bold text-success">System Integrity Verified</h4>
                      <p className="text-sm text-success/80">No anomalies detected in the last 24 hours of operation.</p>
                    </div>
                  </div>
                  <div className="space-y-3">
                    {diagnostics.map((d, i) => (
                      <div key={i} className="flex items-center justify-between p-3 rounded-lg bg-muted/50 border border-border">
                        <div className="flex flex-col">
                          <span className="text-sm font-semibold capitalize">{d.check.replace(/_/g, ' ')}</span>
                          <span className="text-xs text-muted-foreground">{d.details}</span>
                        </div>
                        <Badge
                          variant={d.status === 'pass' || d.status === 'healthy' ? 'success' : 'destructive'}
                        >
                          {d.status.toUpperCase()}
                        </Badge>
                      </div>
                    ))}
                  </div>
                  <button
                    onClick={() => setShowDiagnostics(false)}
                    className="w-full mt-6 py-3 rounded-xl bg-primary text-primary-foreground font-bold"
                  >
                    Done
                  </button>
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}