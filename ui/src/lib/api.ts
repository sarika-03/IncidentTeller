import type {
  IncidentSummaryResponse,
  IncidentListResponse,
  IncidentDetailResponse,
  HealthResponse,
  TimelineResponse,
  ErrorResponse,
} from '@/types';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

class ApiError extends Error {
  constructor(public message: string, public status: number) {
    super(message);
    this.name = 'ApiError';
  }
}

async function apiRequest<T>(endpoint: string, options?: RequestInit): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  const defaultHeaders = {
    'Content-Type': 'application/json',
  };

  const config: RequestInit = {
    ...options,
    headers: {
      ...defaultHeaders,
      ...options?.headers,
    },
  };

  try {
    const response = await fetch(url, config);

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new ApiError(
        errorData.message || `HTTP error! status: ${response.status}`,
        response.status
      );
    }

    return await response.json();
  } catch (error) {
    if (error instanceof ApiError) {
      throw error;
    }
    throw new ApiError('Network error or server unavailable', 0);
  }
}

export const api = {
  // Dashboard
  async getIncidentSummary() {
    return apiRequest<IncidentSummaryResponse>('/incidents/summary');
  },

  // Incident List
  async getIncidents(page: number = 1, pageSize: number = 20) {
    return apiRequest<IncidentListResponse>(`/incidents?page=${page}&page_size=${pageSize}`);
  },

  // Incident Details
  async getIncident(id: string) {
    return apiRequest<IncidentDetailResponse>(`/incidents/${id}`);
  },

  // Health
  async getHealth() {
    return apiRequest<HealthResponse>('/health');
  },

  // Timeline
  async getTimeline(incidentId: string) {
    return apiRequest<TimelineResponse>(`/timeline/${incidentId}`);
  },

  // Enhanced Timeline with cascade detection
  async getEnhancedTimeline(incidentId: string) {
    return apiRequest<any>(`/timeline-enhanced/${incidentId}`);
  },

  // AI Analysis
  async getAIAnalysis() {
    return apiRequest<any>('/analyze', { method: 'POST' });
  },

  // Alert Groups
  async getAlertGroups() {
    return apiRequest<any>('/alert-groups');
  },

  // Logs
  async getLogs() {
    return apiRequest<{ logs: string[]; count: number }>('/logs');
  },

  // Diagnostics
  async getDiagnostics() {
    return apiRequest<{ status: string; diagnostics: any[]; timestamp: string }>('/diagnostics');
  },

  // Metrics Export
  async exportMetrics() {
    const response = await fetch(`${API_BASE_URL}/metrics/export`);
    if (!response.ok) throw new Error('Failed to export metrics');
    return response.blob();
  },

  // SSE for real-time updates
  subscribeToIncidents(onIncident: (incident: any) => void) {
    const eventSource = new EventSource(`${API_BASE_URL}/events`);
    let isConnected = false;
    let hasErrored = false;
    
    eventSource.onopen = () => {
      isConnected = true;
      hasErrored = false;
      console.log('SSE connected');
    };

    eventSource.onmessage = (event) => {
      try {
        const incident = JSON.parse(event.data);
        onIncident(incident);
      } catch (error) {
        // Silently ignore parse errors for malformed SSE data
      }
    };

    eventSource.onerror = (error) => {
      isConnected = false;
      if (!hasErrored) {
        hasErrored = true;
        // Only log the first error to avoid spam
        if (eventSource.readyState === EventSource.CLOSED) {
          console.warn('SSE connection closed');
        } else if (eventSource.readyState === EventSource.CONNECTING) {
          // Connection is retrying, don't spam logs
        }
      }
    };

    return eventSource;
  },
};

export { ApiError };