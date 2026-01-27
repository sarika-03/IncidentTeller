'use client';

import React from 'react';
import { useParams, useRouter } from 'next/navigation';
import Link from 'next/link';
import {
  ArrowLeft,
  AlertTriangle,
  Brain,
  Clock,
  Activity,
  Target,
  CheckCircle,
  XCircle,
  AlertCircle
} from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { api } from '@/lib/api';
import { IncidentDetailResponse } from '@/types';
import { formatDistanceToNow } from 'date-fns';

const getTimelineIcon = (type: string) => {
  switch (type) {
    case 'TRIGGERED':
      return XCircle;
    case 'RESOLVED':
      return CheckCircle;
    case 'UPDATE':
      return AlertCircle;
    default:
      return Activity;
  }
};

const getSeverityColor = (severity: string) => {
  switch (severity) {
    case 'critical':
      return 'text-destructive bg-destructive/10';
    case 'warning':
      return 'text-warning bg-warning/10';
    case 'success':
      return 'text-success bg-success/10';
    default:
      return 'text-muted-foreground bg-muted';
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
      return <Badge variant="destructive">Critical Risk</Badge>;
    case 'high':
      return <Badge variant="destructive">High Risk</Badge>;
    case 'medium':
      return <Badge variant="warning">Medium Risk</Badge>;
    case 'low':
      return <Badge variant="success">Low Risk</Badge>;
    default:
      return <Badge variant="secondary">Unknown</Badge>;
  }
};

interface TimelineEventProps {
  event: any;
  isFirst: boolean;
}

const TimelineEvent: React.FC<TimelineEventProps> = ({ event, isFirst }) => {
  const Icon = getTimelineIcon(event.type);
  const severityColor = getSeverityColor(event.severity);

  return (
    <div className="flex items-start space-x-4">
      <div className="flex flex-col items-center">
        <div className={`p-2 rounded-full ${severityColor}`}>
          <Icon className="h-4 w-4" />
        </div>
        {!isFirst && <div className="w-0.5 h-8 bg-border mt-2" />}
      </div>
      <div className="flex-1 pb-8">
        <div className="flex items-center space-x-2 mb-1">
          <Badge variant="outline">{event.type}</Badge>
          <span className="text-sm text-muted-foreground">
            {event.timestamp && !isNaN(new Date(event.timestamp).getTime()) ? formatDistanceToNow(new Date(event.timestamp), { addSuffix: true }) : 'Unknown time'}
          </span>
          {event.durationSinceStart && (
            <span className="text-sm text-muted-foreground">
              ({event.durationSinceStart})
            </span>
          )}
        </div>
        <p className="text-sm">{event.message}</p>
        <div className="flex items-center space-x-2 mt-1">
          <Badge variant="secondary">{event.resourceType}</Badge>
        </div>
      </div>
    </div>
  );
};

export default function IncidentDetailPage() {
  const params = useParams();
  const router = useRouter();
  const id = params.id as string;

  const [incident, setIncident] = React.useState<IncidentDetailResponse | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  const fetchIncident = async () => {
    try {
      const data = await api.getIncident(id);
      setIncident(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch incident details');
      if (err instanceof Error && err.message.includes('404')) {
        setTimeout(() => router.push('/incidents'), 2000);
      }
    } finally {
      setLoading(false);
    }
  };

  React.useEffect(() => {
    if (id) {
      fetchIncident();
    }
  }, [id]);

  React.useEffect(() => {
    // Set up polling for real-time updates
    const interval = setInterval(fetchIncident, 10000);
    return () => clearInterval(interval);
  }, [id]);

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

  if (!incident) {
    return null;
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center space-x-4">
        <Link href="/incidents">
          <button className="p-2 rounded-md bg-secondary hover:bg-secondary/80 transition-colors">
            <ArrowLeft className="h-4 w-4" />
          </button>
        </Link>
        <div className="flex-1">
          <h1 className="text-3xl font-bold">{incident.title || `${incident.rootCause || 'Unknown'} Incident`}</h1>
          <p className="text-muted-foreground">
            Started {incident.startedAt && !isNaN(new Date(incident.startedAt).getTime()) ? formatDistanceToNow(new Date(incident.startedAt), { addSuffix: true }) : 'Unknown time'}
          </p>
        </div>
        <div className="flex flex-col items-end space-y-2">
          {getStatusBadge(incident.status)}
          {getRiskBadge(incident.riskLevel)}
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid gap-6 lg:grid-cols-3">
        {/* Left Column - Incident Info & Timeline */}
        <div className="lg:col-span-2 space-y-6">
          {/* Incident Overview */}
          <Card>
            <CardHeader>
              <CardTitle>Incident Overview</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div>
                  <span className="text-sm text-muted-foreground">Duration</span>
                  <p className="font-medium">{incident.duration || 'N/A'}</p>
                </div>
                <div>
                  <span className="text-sm text-muted-foreground">Total Events</span>
                  <p className="font-medium">{typeof incident.totalEvents === 'number' && !isNaN(incident.totalEvents) ? incident.totalEvents : 0}</p>
                </div>
                <div>
                  <span className="text-sm text-muted-foreground">Risk Level</span>
                  <p className="font-medium">{incident.riskLevel}</p>
                </div>
                <div>
                  <span className="text-sm text-muted-foreground">Status</span>
                  <p className="font-medium">
                    {incident.resolvedAt ? 'Resolved' : 'Active'}
                  </p>
                </div>
              </div>
            </CardContent>
          </Card>

          {/* Event Timeline */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center space-x-2">
                <Clock className="h-5 w-5" />
                <span>Event Timeline</span>
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {incident.eventTimeline && incident.eventTimeline.length > 0 ? (
                  incident.eventTimeline.map((event, index) => (
                    <TimelineEvent
                      key={index}
                      event={event}
                      isFirst={index === 0}
                    />
                  ))
                ) : (
                  <p className="text-muted-foreground text-sm">No timeline events available</p>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Right Column - AI Analysis */}
        <div className="space-y-6">
          {/* Root Cause Analysis */}
          {incident.rootCause && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Brain className="h-5 w-5" />
                  <span>AI Root Cause Analysis</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <span className="text-sm text-muted-foreground">Primary Cause</span>
                    <p className="font-medium">{incident.rootCause.resourceType} - {incident.rootCause.chart}</p>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Host</span>
                    <p className="font-medium">{incident.rootCause.host}</p>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Confidence</span>
                    <div className="flex items-center space-x-2">
                      <div className="flex-1 bg-secondary rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full"
                          style={{ width: `${(incident.rootCause.confidence || 0) * 100}%` }}
                        ></div>
                      </div>
                      <span className="text-sm font-medium">
                        {Math.round((incident.rootCause.confidence || 0) * 100)}%
                      </span>
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Pattern Type</span>
                    <Badge variant="secondary">{incident.rootCause.patternType}</Badge>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">AI Reasoning</span>
                    <p className="text-sm">{incident.rootCause.reasoning}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}

          {/* Blast Radius Analysis */}
          {incident.blastRadius && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center space-x-2">
                  <Target className="h-5 w-5" />
                  <span>Blast Radius Analysis</span>
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <span className="text-sm text-muted-foreground">Impact Score</span>
                    <div className="flex items-center space-x-2">
                      <div className="flex-1 bg-secondary rounded-full h-2">
                        <div
                          className={`h-2 rounded-full ${incident.blastRadius.impactScore > 0.7 ? 'bg-destructive' :
                            incident.blastRadius.impactScore > 0.4 ? 'bg-warning' : 'bg-success'
                            }`}
                          style={{ width: `${(incident.blastRadius.impactScore || 0) * 100}%` }}
                        ></div>
                      </div>
                      <span className="text-sm font-medium">
                        {Math.round((incident.blastRadius.impactScore || 0) * 100)}%
                      </span>
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Risk Level</span>
                    <Badge variant={
                      incident.blastRadius.riskLevel === 'critical' ? 'destructive' :
                        incident.blastRadius.riskLevel === 'high' ? 'destructive' :
                          incident.blastRadius.riskLevel === 'medium' ? 'warning' : 'success'
                    }>
                      {incident.blastRadius.riskLevel}
                    </Badge>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Cascade Probability</span>
                    <div className="flex items-center space-x-2">
                      <div className="flex-1 bg-secondary rounded-full h-2">
                        <div
                          className="bg-primary h-2 rounded-full"
                          style={{ width: `${(incident.blastRadius.cascadeProbability || 0) * 100}%` }}
                        ></div>
                      </div>
                      <span className="text-sm font-medium">
                        {Math.round((incident.blastRadius.cascadeProbability || 0) * 100)}%
                      </span>
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Predicted Duration</span>
                    <p className="font-medium">{incident.blastRadius.durationPredicted || 'N/A'}</p>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Affected Services</span>
                    <div className="flex flex-wrap gap-1">
                      {incident.blastRadius.affectedServices && incident.blastRadius.affectedServices.length > 0 ? (
                        incident.blastRadius.affectedServices.map((service, index) => (
                          <Badge key={index} variant="outline">{service}</Badge>
                        ))
                      ) : (
                        <span className="text-xs text-muted-foreground">No affected services identified</span>
                      )}
                    </div>
                  </div>
                  <div>
                    <span className="text-sm text-muted-foreground">Business Impact</span>
                    <p className="text-sm">{incident.blastRadius.businessImpact}</p>
                  </div>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  );
}