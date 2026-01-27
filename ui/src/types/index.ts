export interface Alert {
  id: string;
  externalId: number;
  host: string;
  chart: string;
  family: string;
  name: string;
  status: 'UNDEFINED' | 'CLEAR' | 'WARNING' | 'CRITICAL';
  oldStatus: 'UNDEFINED' | 'CLEAR' | 'WARNING' | 'CRITICAL';
  value: number;
  occurredAt: string;
  description: string;
  resourceType: 'UNKNOWN' | 'CPU' | 'MEMORY' | 'DISK' | 'NETWORK' | 'PROCESS';
  labels: Record<string, string>;
}

export interface Incident {
  id: string;
  title: string;
  status: 'UNDEFINED' | 'CLEAR' | 'WARNING' | 'CRITICAL';
  startedAt: string;
  resolvedAt?: string;
  events: Alert[];
}

export interface RootCauseResponse {
  alertId?: string;
  resourceType: string;
  chart: string;
  host: string;
  confidence: number;
  patternType: string;
  reasoning: string;
  alternativeCauses: AlternativeCauseResponse[];
}

export interface AlternativeCauseResponse {
  alertId: string;
  resourceType: string;
  chart: string;
  host: string;
  confidence: number;
}

export interface BlastRadiusResponse {
  impactScore: number;
  affectedServices: string[];
  cascadeProbability: number;
  durationPredicted: string;
  businessImpact: string;
  riskLevel: string;
}

export interface TimelineEventResponse {
  timestamp: string;
  type: 'TRIGGERED' | 'UPDATE' | 'RESOLVED' | 'NOTE';
  message: string;
  severity: 'info' | 'warning' | 'critical' | 'success';
  durationSinceStart?: string;
  resourceType: string;
}

export interface TimelineResponse {
  incidentId: string;
  events: TimelineEventResponse[];
  total: number;
  duration: string;
}

export interface IncidentDetailResponse {
  id: string;
  title: string;
  status: string;
  startedAt: string;
  resolvedAt?: string;
  duration: string;
  rootCause?: RootCauseResponse;
  blastRadius?: BlastRadiusResponse;
  riskLevel: string;
  totalEvents: number;
  eventTimeline: TimelineEventResponse[];
}

export interface IncidentListItemResponse {
  id: string;
  title: string;
  status: string;
  startedAt: string;
  resolvedAt?: string;
  duration: string;
  rootCause: string;
  host?: string;
  totalEvents: number;
  riskLevel: string;
}

export interface IncidentListResponse {
  incidents: IncidentListItemResponse[];
  total: number;
  page: number;
  pageSize: number;
}

export interface IncidentSummaryResponse {
  activeIncidents: number;
  resolvedIncidents: number;
  averageConfidence: number;
  riskLevel: string;
  lastIncidentTime?: string;
}

export interface HealthResponse {
  status: string;
  version: string;
  timestamp: string;
  checks?: Record<string, string>;
}

export interface ErrorResponse {
  error: string;
  message?: string;
  code: number;
}