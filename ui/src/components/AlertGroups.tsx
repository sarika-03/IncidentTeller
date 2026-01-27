import React from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { AlertCircle, GitBranch, Clock, Layers } from 'lucide-react';

interface Alert {
  id: string;
  name: string;
  host: string;
  status: string;
  occurred_at: string;
  value: number;
  resource_type: string;
}

interface AlertGroup {
  id: string;
  alert_count: number;
  primary_host: string;
  affected_hosts: string[];
  resource_types: string[];
  start_time: string;
  end_time: string;
  duration: string;
  is_cascading: boolean;
  group_type: string;
  alerts: Alert[];
}

interface AlertGroupsComponentProps {
  groups: AlertGroup[];
  loading: boolean;
  error: string | null;
}

const getGroupTypeColor = (type: string): string => {
  switch (type) {
    case 'cascading':
      return 'bg-red-100 text-red-800';
    case 'multi_host':
      return 'bg-orange-100 text-orange-800';
    case 'single_host':
      return 'bg-blue-100 text-blue-800';
    default:
      return 'bg-gray-100 text-gray-800';
  }
};

const getStatusColor = (status: string): string => {
  switch (status) {
    case 'CRITICAL':
      return 'text-red-600 bg-red-50';
    case 'WARNING':
      return 'text-orange-600 bg-orange-50';
    case 'CLEAR':
      return 'text-green-600 bg-green-50';
    default:
      return 'text-gray-600 bg-gray-50';
  }
};

export const AlertGroupsComponent: React.FC<AlertGroupsComponentProps> = ({
  groups,
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

  if (groups.length === 0) {
    return (
      <Card>
        <CardContent className="p-6 text-center text-muted-foreground">
          No alert groups found
        </CardContent>
      </Card>
    );
  }

  return (
    <div className="space-y-4">
      {groups.map((group) => (
        <Card key={group.id}>
          <CardHeader>
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center space-x-3 mb-2">
                  {group.is_cascading && (
                    <GitBranch className="h-5 w-5 text-red-500" />
                  )}
                  <CardTitle className="text-lg">
                    {group.primary_host}
                  </CardTitle>
                  <Badge className={getGroupTypeColor(group.group_type)}>
                    {group.group_type.replace('_', ' ')}
                  </Badge>
                  {group.is_cascading && (
                    <Badge variant="destructive">Cascading</Badge>
                  )}
                </div>
                <p className="text-sm text-muted-foreground">
                  {group.alert_count} alerts
                </p>
              </div>
              <div className="text-right">
                <div className="flex items-center space-x-2 text-sm text-muted-foreground">
                  <Clock className="h-4 w-4" />
                  <span>{group.duration || 'N/A'}</span>
                </div>
              </div>
            </div>
          </CardHeader>

          <CardContent className="space-y-4">
            {/* Affected Hosts */}
            {group.affected_hosts.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold mb-2 flex items-center space-x-2">
                  <AlertCircle className="h-4 w-4" />
                  <span>Affected Hosts</span>
                </h4>
                <div className="flex flex-wrap gap-2">
                  {group.affected_hosts.map((host, idx) => (
                    <Badge key={`${group.id}-host-${idx}`} variant="secondary">
                      {host}
                    </Badge>
                  ))}
                </div>
              </div>
            )}

            {/* Resource Types */}
            {group.resource_types.length > 0 && (
              <div>
                <h4 className="text-sm font-semibold mb-2 flex items-center space-x-2">
                  <Layers className="h-4 w-4" />
                  <span>Resource Types</span>
                </h4>
                <div className="flex flex-wrap gap-2">
                  {group.resource_types.map((type, idx) => (
                    <Badge key={`${group.id}-type-${idx}`} variant="outline">
                      {type}
                    </Badge>
                  ))}
                </div>
              </div>
            )}

            {/* Alert List */}
            <div>
              <h4 className="text-sm font-semibold mb-3">Alerts</h4>
              <div className="space-y-2 max-h-64 overflow-y-auto">
                {group.alerts.map((alert, idx) => (
                  <div
                    key={`${group.id}-alert-${idx}`}
                    className="flex items-start justify-between p-2 rounded border border-border text-sm"
                  >
                    <div className="flex-1">
                      <p className="font-medium">{alert.name}</p>
                      <p className="text-xs text-muted-foreground">
                        {alert.host} • {alert.resource_type}
                      </p>
                    </div>
                    <div className="flex items-center space-x-2">
                      <Badge className={getStatusColor(alert.status)}>
                        {alert.status}
                      </Badge>
                      <span className="text-xs text-muted-foreground">
                        {new Date(alert.occurred_at).toLocaleTimeString()}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Timeline Info */}
            <div className="pt-2 border-t flex justify-between text-xs text-muted-foreground">
              <span>Start: {new Date(group.start_time).toLocaleTimeString()}</span>
              <span>End: {new Date(group.end_time).toLocaleTimeString()}</span>
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  );
};
