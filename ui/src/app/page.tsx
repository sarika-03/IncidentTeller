'use client';

import React from 'react';
import Link from 'next/link';
import { RefreshCw } from 'lucide-react';
import { api } from '@/lib/api';
import { IncidentSummaryResponse } from '@/types';

export default function DashboardPage() {
  const [summary, setSummary] = React.useState<IncidentSummaryResponse | null>(null);
  const [loading, setLoading] = React.useState(true);
  const [error, setError] = React.useState<string | null>(null);

  const fetchData = async () => {
    try {
      const data = await api.getIncidentSummary();
      setSummary(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch dashboard data');
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    setLoading(true);
    fetchData();
  };

  React.useEffect(() => {
    fetchData();

    // Set up SSE for real-time updates
    const eventSource = api.subscribeToIncidents(() => {
      fetchData();
    });

    // Fallback polling
    const interval = setInterval(fetchData, 30000);

    return () => {
      eventSource.close();
      clearInterval(interval);
    };
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 p-4 md:p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8 flex items-center justify-between">
          <div>
            <h1 className="text-3xl font-bold text-gray-900">IncidentTeller Dashboard</h1>
            <p className="text-gray-600 mt-1">Real-time incident monitoring and analysis</p>
          </div>
          <button
            onClick={handleRefresh}
            disabled={loading}
            className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 transition-colors"
          >
            <RefreshCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
            {loading ? 'Loading...' : 'Refresh'}
          </button>
        </div>

        {/* Error Message */}
        {error && (
          <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-red-800 font-medium">Error</p>
            <p className="text-red-700 text-sm">{error}</p>
          </div>
        )}

        {/* Loading State */}
        {loading && !summary && (
          <div className="text-center py-12">
            <div className="inline-block">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
            <p className="text-gray-500 mt-4">Loading dashboard...</p>
          </div>
        )}

        {/* Dashboard Content */}
        {summary && (
          <>
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
              <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow">
                <p className="text-gray-600 text-sm font-medium mb-2">Active Incidents</p>
                <p className="text-4xl font-bold text-red-600">{typeof summary.activeIncidents === 'number' && !isNaN(summary.activeIncidents) ? summary.activeIncidents : 0}</p>
                <p className="text-xs text-gray-500 mt-2">Currently active</p>
              </div>

              <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow">
                <p className="text-gray-600 text-sm font-medium mb-2">Resolved Incidents</p>
                <p className="text-4xl font-bold text-green-600">{typeof summary.resolvedIncidents === 'number' && !isNaN(summary.resolvedIncidents) ? summary.resolvedIncidents : 0}</p>
                <p className="text-xs text-gray-500 mt-2">Successfully resolved</p>
              </div>

              <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow">
                <p className="text-gray-600 text-sm font-medium mb-2">AI Confidence</p>
                <p className="text-4xl font-bold text-blue-600">{Math.round((summary.averageConfidence || 0) * 100)}%</p>
                <p className="text-xs text-gray-500 mt-2">Analysis accuracy</p>
              </div>

              <div className="bg-white p-6 rounded-lg border border-gray-200 shadow-sm hover:shadow-md transition-shadow">
                <p className="text-gray-600 text-sm font-medium mb-2">Risk Level</p>
                <p className={`text-4xl font-bold ${
                  summary.riskLevel === 'high' ? 'text-red-600' :
                  summary.riskLevel === 'medium' ? 'text-yellow-600' :
                  'text-green-600'
                }`}>
                  {(summary.riskLevel || 'Unknown').toUpperCase()}
                </p>
                <p className="text-xs text-gray-500 mt-2">Current level</p>
              </div>
            </div>

            {/* Quick Actions */}
            <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6 mb-8">
              <h2 className="text-lg font-bold text-gray-900 mb-4">Quick Navigation</h2>
              <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-5 gap-3">
                <Link href="/incidents">
                  <button className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm">
                    View Incidents
                  </button>
                </Link>
                <Link href="/timeline">
                  <button className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm">
                    Timeline
                  </button>
                </Link>
                <Link href="/health">
                  <button className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm">
                    System Health
                  </button>
                </Link>
                <Link href="/analysis">
                  <button className="w-full px-4 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors font-medium text-sm">
                    AI Analysis
                  </button>
                </Link>
                <button
                  onClick={handleRefresh}
                  className="w-full px-4 py-3 bg-gray-200 text-gray-900 rounded-lg hover:bg-gray-300 transition-colors font-medium text-sm"
                >
                  Refresh
                </button>
              </div>
            </div>

            {/* System Information */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
              <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6">
                <h2 className="text-lg font-bold text-gray-900 mb-4">System Overview</h2>
                <div className="space-y-3">
                  <div className="flex justify-between items-center border-b border-gray-100 pb-3">
                    <span className="text-gray-600">Total Incidents</span>
                    <span className="font-bold text-gray-900 text-lg">{(typeof summary.activeIncidents === 'number' && !isNaN(summary.activeIncidents) ? summary.activeIncidents : 0) + (typeof summary.resolvedIncidents === 'number' && !isNaN(summary.resolvedIncidents) ? summary.resolvedIncidents : 0)}</span>
                  </div>
                  <div className="flex justify-between items-center border-b border-gray-100 pb-3">
                    <span className="text-gray-600">Resolution Rate</span>
                    <span className="font-bold text-gray-900 text-lg">
                      {(() => {
                        const active = typeof summary.activeIncidents === 'number' && !isNaN(summary.activeIncidents) ? summary.activeIncidents : 0;
                        const resolved = typeof summary.resolvedIncidents === 'number' && !isNaN(summary.resolvedIncidents) ? summary.resolvedIncidents : 0;
                        const total = active + resolved;
                        return total > 0 ? Math.round((resolved / total) * 100) : 0;
                      })()}%
                    </span>
                  </div>
                  <div className="flex justify-between items-center">
                    <span className="text-gray-600">Status</span>
                    <span className="px-3 py-1 bg-green-100 text-green-800 rounded-full text-sm font-medium">
                      Operational
                    </span>
                  </div>
                </div>
              </div>

              <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6">
                <h2 className="text-lg font-bold text-gray-900 mb-4">Latest Activity</h2>
                <div className="space-y-3">
                  {summary.lastIncidentTime ? (
                    <>
                      <div className="flex justify-between items-center border-b border-gray-100 pb-3">
                        <span className="text-gray-600">Last Incident</span>
                        <span className="font-bold text-gray-900 text-sm">
                          {new Date(summary.lastIncidentTime).toLocaleString()}
                        </span>
                      </div>
                      <div className="flex justify-between items-center border-b border-gray-100 pb-3">
                        <span className="text-gray-600">Time Since Last</span>
                        <span className="font-bold text-gray-900 text-sm">
                          {Math.floor((Date.now() - new Date(summary.lastIncidentTime).getTime()) / 60000)} minutes ago
                        </span>
                      </div>
                    </>
                  ) : (
                    <p className="text-gray-600 text-sm">No incident data available</p>
                  )}
                  <div className="flex justify-between items-center">
                    <span className="text-gray-600">Status</span>
                    <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                      summary.activeIncidents > 0 
                        ? 'bg-red-100 text-red-800' 
                        : 'bg-green-100 text-green-800'
                    }`}>
                      {summary.activeIncidents > 0 ? 'Incidents Active' : 'All Clear'}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
