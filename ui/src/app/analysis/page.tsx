'use client';

import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { AIAnalysisComponent } from '@/components/AIAnalysis';
import { AlertGroupsComponent } from '@/components/AlertGroups';
import { api } from '@/lib/api';
import { AlertTriangle, Brain, Layers } from 'lucide-react';

export default function AnalysisPage() {
  const [aiAnalysis, setAIAnalysis] = React.useState<any>(null);
  const [alertGroups, setAlertGroups] = React.useState<any[]>([]);
  const [aiLoading, setAILoading] = React.useState(false);
  const [groupsLoading, setGroupsLoading] = React.useState(false);
  const [aiError, setAIError] = React.useState<string | null>(null);
  const [groupsError, setGroupsError] = React.useState<string | null>(null);
  const [activeTab, setActiveTab] = React.useState<'analysis' | 'groups'>('analysis');

  const fetchAIAnalysis = async () => {
    setAILoading(true);
    setAIError(null);
    try {
      const data = await api.getAIAnalysis();
      setAIAnalysis(data);
    } catch (error) {
      setAIError(error instanceof Error ? error.message : 'Failed to fetch AI analysis');
    } finally {
      setAILoading(false);
    }
  };

  const fetchAlertGroups = async () => {
    setGroupsLoading(true);
    setGroupsError(null);
    try {
      const data = await api.getAlertGroups();
      setAlertGroups(data.groups || []);
    } catch (error) {
      setGroupsError(error instanceof Error ? error.message : 'Failed to fetch alert groups');
    } finally {
      setGroupsLoading(false);
    }
  };

  React.useEffect(() => {
    fetchAIAnalysis();
    fetchAlertGroups();
  }, []);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold">AI Analysis</h1>
        <p className="text-muted-foreground">
          AI-powered incident analysis with alert grouping and recommendations
        </p>
      </div>

      {/* Tabs */}
      <div className="flex gap-2 border-b">
        <button
          onClick={() => setActiveTab('analysis')}
          className={`px-4 py-2 font-medium transition-colors flex items-center space-x-2 ${
            activeTab === 'analysis'
              ? 'border-b-2 border-primary text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <Brain className="h-4 w-4" />
          <span>AI Analysis</span>
        </button>
        <button
          onClick={() => setActiveTab('groups')}
          className={`px-4 py-2 font-medium transition-colors flex items-center space-x-2 ${
            activeTab === 'groups'
              ? 'border-b-2 border-primary text-primary'
              : 'text-muted-foreground hover:text-foreground'
          }`}
        >
          <Layers className="h-4 w-4" />
          <span>Alert Groups</span>
        </button>
      </div>

      {/* Content */}
      <div>
        {activeTab === 'analysis' ? (
          <div className="space-y-4">
            <div className="flex justify-end">
              <button
                onClick={fetchAIAnalysis}
                disabled={aiLoading}
                className="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 transition-colors flex items-center gap-2"
              >
                {aiLoading ? 'Analyzing...' : 'Regenerate Analysis'}
              </button>
            </div>
            <AIAnalysisComponent
              analysis={aiAnalysis}
              loading={aiLoading}
              error={aiError}
            />
          </div>
        ) : (
          <div className="space-y-4">
            <div className="flex justify-end">
              <button
                onClick={fetchAlertGroups}
                disabled={groupsLoading}
                className="px-4 py-2 rounded-lg bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50 transition-colors flex items-center gap-2"
              >
                {groupsLoading ? 'Loading...' : 'Refresh Groups'}
              </button>
            </div>
            <AlertGroupsComponent
              groups={alertGroups}
              loading={groupsLoading}
              error={groupsError}
            />
          </div>
        )}
      </div>
    </div>
  );
}
