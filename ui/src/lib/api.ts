import type {
  IncidentSummaryResponse,
  IncidentListResponse,
  IncidentDetailResponse,
  HealthResponse,
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
};

export { ApiError };