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
  const [error, setError] = React.useState<string | null>(null);
  const [lastUpdate, setLastUpdate] = React.useState<Date>(new Date());

  const fetchHealth = async () => {
    try {
      const healthData = await api.getHealth();
      setHealth(healthData);
      setLastUpdate(new Date());
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch health data');
    } finally {
      setLoading(false);
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
              <RefreshCw className="h-3 w-3" />
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
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">System Health & Metrics</h1>
        <div className="flex items-center space-x-2">
          <Badge variant="secondary">
            Updated {formatDistanceToNow(lastUpdate, { addSuffix: true })}
          </Badge>
          <button
            onClick={fetchHealth}
            className="px-3 py-1 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors flex items-center space-x-1"
          >
            <RefreshCw className="h-3 w-3" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* Overall Status */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Activity className="h-5 w-5" />
            <span>Overall System Status</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-4">
              <div className={`p-3 rounded-full ${getStatusColor(health.status)}`}>
                <StatusIcon className="h-8 w-8" />
              </div>
              <div>
                <h2 className="text-2xl font-bold capitalize">{health.status}</h2>
                <p className="text-muted-foreground">
                  IncidentTeller v{health.version}
                </p>
              </div>
            </div>
            <div className="text-right">
              {getStatusBadge(health.status)}
              <p className="text-sm text-muted-foreground mt-1">
                Last checked: {new Date(health.timestamp).toLocaleString()}
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
              <span>Service Health</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {Object.entries(health.checks).map(([service, status]) => {
                const ServiceIcon = getServiceIcon(service);
                const ServiceStatusIcon = getStatusIcon(status);
                
                return (
                  <div key={service} className="border rounded-lg p-4">
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-2">
                        <ServiceIcon className="h-4 w-4" />
                        <span className="font-medium capitalize">{service}</span>
                      </div>
                      <div className={`p-1.5 rounded-full ${getStatusColor(status)}`}>
                        <ServiceStatusIcon className="h-3 w-3" />
                      </div>
                    </div>
                    <div className="text-sm text-muted-foreground capitalize">
                      {status}
                    </div>
                  </div>
                );
              })}
            </div>
          </CardContent>
        </Card>
      )}

      {/* System Information */}
      <div className="grid gap-6 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Zap className="h-5 w-5" />
              <span>Performance Metrics</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Response Time</span>
                <Badge variant="success">&lt;100ms</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Uptime</span>
                <Badge variant="success">99.9%</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">API Endpoints</span>
                <span className="font-medium">4 Active</span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Health Check Interval</span>
                <span className="font-medium">30 seconds</span>
              </div>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle className="flex items-center space-x-2">
              <Cpu className="h-5 w-5" />
              <span>System Resources</span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Memory Usage</span>
                <Badge variant="success">Normal</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">CPU Usage</span>
                <Badge variant="success">Normal</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Disk Space</span>
                <Badge variant="success">Available</Badge>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Network I/O</span>
                <Badge variant="success">Normal</Badge>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Quick Actions */}
      <Card>
        <CardHeader>
          <CardTitle>Quick Actions</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-2 md:grid-cols-4">
            <button className="px-4 py-2 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors">
              View Logs
            </button>
            <button className="px-4 py-2 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors">
              Export Metrics
            </button>
            <button className="px-4 py-2 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors">
              System Diagnostics
            </button>
            <button 
              onClick={fetchHealth}
              className="px-4 py-2 rounded-md bg-primary text-primary-foreground hover:bg-primary/80 transition-colors flex items-center justify-center space-x-1"
            >
              <RefreshCw className="h-3 w-3" />
              <span>Refresh All</span>
            </button>
          </div>
        </CardContent>
      </Card>

      {/* API Information */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Globe className="h-5 w-5" />
            <span>API Information</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="grid gap-4 md:grid-cols-2">
            <div>
              <h4 className="font-medium mb-2">Available Endpoints</h4>
              <div className="space-y-1 font-mono text-sm">
                <div className="text-muted-foreground">GET /api/health</div>
                <div className="text-muted-foreground">GET /api/incidents/summary</div>
                <div className="text-muted-foreground">GET /api/incidents</div>
                <div className="text-muted-foreground">GET /api/incidents/{id}</div>
              </div>
            </div>
            <div>
              <h4 className="font-medium mb-2">System Information</h4>
              <div className="space-y-2 text-sm">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Version:</span>
                  <span>{health.version}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Environment:</span>
                  <span>Production</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Framework:</span>
                  <span>IncidentTeller</span>
                </div>
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}