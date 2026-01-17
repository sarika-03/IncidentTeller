import React from 'react';
import Link from 'next/link';
import { AlertTriangle, CheckCircle, Clock, Cpu, HardDrive, Network, Activity } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';
import { IncidentListResponse, IncidentListItemResponse } from '@/types';
import { formatDistanceToNow } from 'date-fns';

const getResourceIcon = (resourceType: string) => {
  switch (resourceType.toLowerCase()) {
    case 'cpu':
      return Cpu;
    case 'memory':
      return Activity;
    case 'disk':
      return HardDrive;
    case 'network':
      return Network;
    default:
      return AlertTriangle;
  }
};

const getStatusBadge = (status: string) => {
  switch (status) {
    case 'CRITICAL':
      return <Badge variant="destructive">Critical</Badge>;
    case 'WARNING':
      return <Badge variant="warning">Warning</Badge>;
    case 'CLEAR':
      return <Badge variant="success">Resolved</Badge>;
    default:
      return <Badge variant="secondary">Unknown</Badge>;
  }
};

const getRiskBadge = (riskLevel: string) => {
  switch (riskLevel) {
    case 'critical':
      return <Badge variant="destructive">Critical</Badge>;
    case 'high':
      return <Badge variant="destructive">High</Badge>;
    case 'medium':
      return <Badge variant="warning">Medium</Badge>;
    case 'low':
      return <Badge variant="success">Low</Badge>;
    default:
      return <Badge variant="secondary">Unknown</Badge>;
  }
};

interface IncidentCardProps {
  incident: IncidentListItemResponse;
}

const IncidentCard: React.FC<IncidentCardProps> = ({ incident }) => {
  const ResourceIcon = getResourceIcon(incident.rootCause);

  return (
    <Card className="hover:shadow-md transition-shadow cursor-pointer">
      <Link href={`/incidents/${incident.id}`}>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <div className="p-2 bg-muted rounded-md">
                <ResourceIcon className="h-4 w-4" />
              </div>
              <div>
                <CardTitle className="text-lg">{incident.title}</CardTitle>
                <p className="text-sm text-muted-foreground">{incident.rootCause} • {incident.host || 'Unknown'}</p>
              </div>
            </div>
            <div className="flex flex-col items-end space-y-2">
              {getStatusBadge(incident.status)}
              {getRiskBadge(incident.riskLevel)}
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
            <div>
              <span className="text-muted-foreground">Started</span>
              <p className="font-medium">{formatDistanceToNow(new Date(incident.startedAt), { addSuffix: true })}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Duration</span>
              <p className="font-medium">{incident.duration}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Events</span>
              <p className="font-medium">{incident.totalEvents}</p>
            </div>
            <div>
              <span className="text-muted-foreground">Status</span>
              <p className="font-medium">
                {incident.resolvedAt ? (
                  <span className="text-success">Resolved</span>
                ) : (
                  <span className="text-destructive">Active</span>
                )}
              </p>
            </div>
          </div>
        </CardContent>
      </Link>
    </Card>
  );
};

export default function IncidentsPage() {
  const [incidents, setIncidents] = React.useState<IncidentListResponse | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [page, setPage] = React.useState(1);

  const fetchIncidents = async (pageNum: number = 1) => {
    try {
      const data = await api.getIncidents(pageNum, 20);
      setIncidents(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch incidents');
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchIncidents(page);
  }, [page]);

  React.useEffect(() => {
    // Set up polling for real-time updates
    const interval = setInterval(() => {
      fetchIncidents(page);
    }, 15000);
    
    return () => clearInterval(interval);
  }, [page]);

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
          <div className="flex items-center space-x-2">
            <AlertTriangle className="h-5 w-5 text-destructive" />
            <span className="text-destructive">{error}</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!incidents || incidents.incidents.length === 0) {
    return (
      <div className="space-y-6">
        <h1 className="text-3xl font-bold">Incidents</h1>
        <Card>
          <CardContent className="p-12 text-center">
            <CheckCircle className="h-16 w-16 text-success mx-auto mb-4" />
            <h3 className="text-xl font-semibold mb-2">No Incidents Found</h3>
            <p className="text-muted-foreground">All systems are operating normally. No incidents have been reported.</p>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Incidents</h1>
        <div className="flex items-center space-x-4">
          <Badge variant="secondary">
            {incidents.total} Total Incidents
          </Badge>
          <Badge variant="secondary">
            Page {page} of {Math.ceil(incidents.total / incidents.pageSize)}
          </Badge>
        </div>
      </div>

      <div className="grid gap-4">
        {incidents.incidents.map((incident) => (
          <IncidentCard key={incident.id} incident={incident} />
        ))}
      </div>

      {incidents.total > incidents.pageSize && (
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            Showing {((page - 1) * incidents.pageSize) + 1} to {Math.min(page * incidents.pageSize, incidents.total)} of {incidents.total} incidents
          </div>
          <div className="flex space-x-2">
            <button
              onClick={() => setPage(p => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <button
              onClick={() => setPage(p => p + 1)}
              disabled={page * incidents.pageSize >= incidents.total}
              className="px-3 py-1 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}