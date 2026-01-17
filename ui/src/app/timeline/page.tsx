import React from 'react';
import { 
  Activity, 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  Pause,
  Play,
  RefreshCw
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';
import { IncidentDetailResponse, IncidentListResponse } from '@/types';
import { formatDistanceToNow } from 'date-fns';

const getTimelineIcon = (type: string) => {
  switch (type) {
    case 'TRIGGERED':
      return AlertTriangle;
    case 'RESOLVED':
      return CheckCircle;
    case 'UPDATE':
      return Activity;
    default:
      return Activity;
  }
};

const getSeverityColor = (severity: string) => {
  switch (severity) {
    case 'critical':
      return 'text-destructive bg-destructive/10 border-destructive/20';
    case 'warning':
      return 'text-warning bg-warning/10 border-warning/20';
    case 'success':
      return 'text-success bg-success/10 border-success/20';
    default:
      return 'text-muted-foreground bg-muted border-border';
  }
};

interface TimelineEventProps {
  event: any;
  showIncidentTitle?: boolean;
}

const TimelineEvent: React.FC<TimelineEventProps> = ({ event, showIncidentTitle = false }) => {
  const Icon = getTimelineIcon(event.type);
  const severityColor = getSeverityColor(event.severity);

  return (
    <div className={`border-l-2 ${severityColor.split(' ')[2]} pl-4 pb-4 relative`}>
      <div className="absolute -left-2 top-0">
        <div className={`p-1.5 rounded-full ${severityColor.split(' ').slice(0, 2).join(' ')}`}>
          <Icon className="h-3 w-3" />
        </div>
      </div>
      <div className="space-y-1">
        <div className="flex items-center space-x-2">
          <span className="text-xs text-muted-foreground font-mono">
            {new Date(event.timestamp).toLocaleTimeString()}
          </span>
          <Badge variant="outline" className="text-xs">
            {event.type}
          </Badge>
          <Badge variant="secondary" className="text-xs">
            {event.resourceType}
          </Badge>
        </div>
        {showIncidentTitle && (
          <p className="text-sm font-medium text-primary">
            {event.incidentTitle}
          </p>
        )}
        <p className="text-sm">{event.message}</p>
        <p className="text-xs text-muted-foreground">
          {formatDistanceToNow(new Date(event.timestamp), { addSuffix: true })}
        </p>
      </div>
    </div>
  );
};

export default function LiveTimelinePage() {
  const [incidents, setIncidents] = React.useState<IncidentListResponse | null>(null);
  const [detailedIncidents, setDetailedIncidents] = React.useState<Record<string, IncidentDetailResponse>>({});
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);
  const [isPaused, setIsPaused] = React.useState(false);
  const [autoRefresh, setAutoRefresh] = React.useState(true);
  const [refreshInterval, setRefreshInterval] = React.useState(5000);

  const fetchAllIncidents = async () => {
    try {
      // Fetch incident list
      const incidentsData = await api.getIncidents(1, 50); // Get more incidents for timeline
      setIncidents(incidentsData);

      // Fetch detailed data for each incident to get timeline events
      const detailedData: Record<string, IncidentDetailResponse> = {};
      for (const incident of incidentsData.incidents.slice(0, 10)) { // Limit to first 10 for performance
        try {
          const detail = await api.getIncident(incident.id);
          detailedData[incident.id] = detail;
        } catch (err) {
          console.error(`Failed to fetch details for incident ${incident.id}:`, err);
        }
      }
      setDetailedIncidents(detailedData);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch timeline data');
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    fetchAllIncidents();
  }, []);

  React.useEffect(() => {
    if (!isPaused && autoRefresh) {
      const interval = setInterval(fetchAllIncidents, refreshInterval);
      return () => clearInterval(interval);
    }
  }, [isPaused, autoRefresh, refreshInterval]);

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
              onClick={fetchAllIncidents}
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

  // Combine all timeline events from all incidents
  const allTimelineEvents: any[] = [];
  Object.entries(detailedIncidents).forEach(([incidentId, incident]) => {
    incident.eventTimeline.forEach((event, index) => {
      allTimelineEvents.push({
        ...event,
        incidentId,
        incidentTitle: incident.title,
        isFirstInIncident: index === 0,
      });
    });
  });

  // Sort by timestamp (newest first)
  allTimelineEvents.sort((a, b) => new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime());

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <h1 className="text-3xl font-bold">Live Timeline</h1>
          <Badge variant={autoRefresh ? "success" : "secondary"} className="animate-pulse">
            {autoRefresh ? 'Live' : 'Paused'}
          </Badge>
        </div>
        
        <div className="flex items-center space-x-2">
          <select
            value={refreshInterval}
            onChange={(e) => setRefreshInterval(Number(e.target.value))}
            className="px-3 py-1 rounded-md bg-secondary text-secondary-foreground border border-border"
          >
            <option value={5000}>5s</option>
            <option value={10000}>10s</option>
            <option value={30000}>30s</option>
            <option value={60000}>1m</option>
          </select>
          
          <button
            onClick={() => setAutoRefresh(!autoRefresh)}
            className={`px-3 py-1 rounded-md transition-colors flex items-center space-x-1 ${
              autoRefresh 
                ? 'bg-success text-success-foreground hover:bg-success/80' 
                : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
          >
            {autoRefresh ? <Pause className="h-3 w-3" /> : <Play className="h-3 w-3" />}
            <span>{autoRefresh ? 'Pause' : 'Resume'}</span>
          </button>
          
          <button
            onClick={fetchAllIncidents}
            className="px-3 py-1 rounded-md bg-secondary text-secondary-foreground hover:bg-secondary/80 transition-colors flex items-center space-x-1"
          >
            <RefreshCw className="h-3 w-3" />
            <span>Refresh</span>
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Total Incidents</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{incidents?.total || 0}</div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Timeline Events</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{allTimelineEvents.length}</div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Active Incidents</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-destructive">
              {incidents?.incidents.filter(i => !i.resolvedAt).length || 0}
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Last Update</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-sm font-mono">
              {allTimelineEvents.length > 0 
                ? formatDistanceToNow(new Date(allTimelineEvents[0].timestamp), { addSuffix: true })
                : 'No events'
              }
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Timeline */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center space-x-2">
            <Clock className="h-5 w-5" />
            <span>Real-time Events</span>
          </CardTitle>
        </CardHeader>
        <CardContent>
          {allTimelineEvents.length === 0 ? (
            <div className="text-center py-12">
              <Activity className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold mb-2">No Timeline Events</h3>
              <p className="text-muted-foreground">No incident events have been recorded yet.</p>
            </div>
          ) : (
            <div className="max-h-96 overflow-y-auto space-y-0">
              {allTimelineEvents.map((event, index) => (
                <TimelineEvent 
                  key={`${event.incidentId}-${index}`} 
                  event={event} 
                  showIncidentTitle={event.isFirstInIncident}
                />
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}