import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Brain, AlertCircle, CheckCircle, Zap, Clock } from 'lucide-react';

interface AIAnalysisData {
  summary: string;
  root_cause_text: string;
  impact_assessment: string;
  recommendations: {
    immediate: string[];
    short_term: string[];
    long_term: string[];
  };
  generated_at: string;
  alert_count: number;
  time_span: string;
}

interface AIAnalysisComponentProps {
  analysis: AIAnalysisData | null;
  loading: boolean;
  error: string | null;
}

export const AIAnalysisComponent: React.FC<AIAnalysisComponentProps> = ({
  analysis,
  loading,
  error,
}) => {
  if (loading) {
    return (
      <Card>
        <CardContent className="p-6 flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </CardContent>
      </Card>
    );
  }

  if (error) {
    return (
      <Card>
        <CardContent className="p-6">
          <div className="flex items-center space-x-2 text-destructive">
            <AlertCircle className="h-5 w-5" />
            <span>{error}</span>
          </div>
        </CardContent>
      </Card>
    );
  }

  if (!analysis) {
    return null;
  }

  return (
    <div className="space-y-4">
      {/* Summary */}
      <Card>
        <CardHeader>
          <div className="flex items-center space-x-2">
            <Brain className="h-5 w-5 text-blue-500" />
            <CardTitle>AI Summary</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm leading-relaxed">{analysis.summary}</p>
          <div className="mt-4 flex items-center space-x-4 text-xs text-muted-foreground">
            <span className="flex items-center space-x-1">
              <AlertCircle className="h-4 w-4" />
              <span>{analysis.alert_count} alerts</span>
            </span>
            <span className="flex items-center space-x-1">
              <Clock className="h-4 w-4" />
              <span>{analysis.time_span}</span>
            </span>
          </div>
        </CardContent>
      </Card>

      {/* Root Cause Analysis */}
      <Card>
        <CardHeader>
          <div className="flex items-center space-x-2">
            <AlertCircle className="h-5 w-5 text-orange-500" />
            <CardTitle>Root Cause Analysis</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm leading-relaxed whitespace-pre-wrap">{analysis.root_cause_text}</p>
        </CardContent>
      </Card>

      {/* Impact Assessment */}
      <Card>
        <CardHeader>
          <div className="flex items-center space-x-2">
            <Zap className="h-5 w-5 text-red-500" />
            <CardTitle>Impact Assessment</CardTitle>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm leading-relaxed">{analysis.impact_assessment}</p>
        </CardContent>
      </Card>

      {/* Recommendations */}
      <Card>
        <CardHeader>
          <div className="flex items-center space-x-2">
            <CheckCircle className="h-5 w-5 text-green-500" />
            <CardTitle>Recommended Actions</CardTitle>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {/* Immediate Actions */}
          {analysis.recommendations.immediate.length > 0 && (
            <div>
              <Badge className="mb-2 bg-red-500">Immediate - within 5 minutes</Badge>
              <ul className="space-y-2">
                {analysis.recommendations.immediate.map((action, idx) => (
                  <li key={idx} className="flex items-start space-x-2 text-sm">
                    <span className="text-red-500 font-bold mt-1">•</span>
                    <span>{action}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Short-term Actions */}
          {analysis.recommendations.short_term.length > 0 && (
            <div>
              <Badge className="mb-2 bg-orange-500">Short-term - within 8 hours</Badge>
              <ul className="space-y-2">
                {analysis.recommendations.short_term.map((action, idx) => (
                  <li key={idx} className="flex items-start space-x-2 text-sm">
                    <span className="text-orange-500 font-bold mt-1">•</span>
                    <span>{action}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Long-term Actions */}
          {analysis.recommendations.long_term.length > 0 && (
            <div>
              <Badge className="mb-2 bg-blue-500">Long-term (Prevention)</Badge>
              <ul className="space-y-2">
                {analysis.recommendations.long_term.map((action, idx) => (
                  <li key={idx} className="flex items-start space-x-2 text-sm">
                    <span className="text-blue-500 font-bold mt-1">•</span>
                    <span>{action}</span>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Generated Info */}
      <div className="text-xs text-muted-foreground">
        Analysis generated at {new Date(analysis.generated_at).toLocaleString()}
      </div>
    </div>
  );
};
