'use client';

import React from 'react';
import { AlertTriangle, RefreshCw, ChevronDown, AlertCircle } from 'lucide-react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ReactNode;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false, error: null };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: React.ErrorInfo) {
    console.error('Error caught by boundary:', error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return (
        this.props.fallback || (
          <Card className="border-destructive">
            <CardContent className="p-6">
              <div className="flex items-start space-x-3">
                <AlertTriangle className="h-5 w-5 text-destructive flex-shrink-0 mt-0.5" />
                <div>
                  <h3 className="font-semibold text-destructive mb-1">Something went wrong</h3>
                  <p className="text-sm text-muted-foreground mb-2">
                    {this.state.error?.message || 'An unexpected error occurred'}
                  </p>
                  <button
                    onClick={() => this.setState({ hasError: false, error: null })}
                    className="text-sm text-primary hover:underline"
                  >
                    Try again
                  </button>
                </div>
              </div>
            </CardContent>
          </Card>
        )
      );
    }

    return this.props.children;
  }
}

interface LoadingStateProps {
  isLoading: boolean;
  error?: string | null;
  children: React.ReactNode;
  onRetry?: () => void;
}

export const LoadingState: React.FC<LoadingStateProps> = ({
  isLoading,
  error,
  children,
  onRetry,
}) => {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-6">
        <div className="flex flex-col items-center space-y-2">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <p className="text-sm text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <Card className="border-destructive">
        <CardContent className="p-6">
          <div className="flex items-start justify-between">
            <div className="flex items-start space-x-3 flex-1">
              <AlertCircle className="h-5 w-5 text-destructive flex-shrink-0 mt-0.5" />
              <div>
                <h3 className="font-semibold text-destructive mb-1">Error</h3>
                <p className="text-sm text-muted-foreground">{error}</p>
              </div>
            </div>
            {onRetry && (
              <button
                onClick={onRetry}
                className="ml-2 p-2 rounded hover:bg-muted transition-colors"
                title="Retry"
              >
                <RefreshCw className="h-4 w-4" />
              </button>
            )}
          </div>
        </CardContent>
      </Card>
    );
  }

  return <>{children}</>;
};

interface CollapsibleProps {
  title: string;
  defaultOpen?: boolean;
  children: React.ReactNode;
  badge?: string;
}

export const Collapsible: React.FC<CollapsibleProps> = ({
  title,
  defaultOpen = false,
  children,
  badge,
}) => {
  const [isOpen, setIsOpen] = React.useState(defaultOpen);

  return (
    <Card>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full px-6 py-4 flex items-center justify-between hover:bg-muted/50 transition-colors"
      >
        <div className="flex items-center space-x-2">
          <ChevronDown
            className={`h-4 w-4 transition-transform ${isOpen ? 'rotate-180' : ''}`}
          />
          <h3 className="font-semibold">{title}</h3>
          {badge && <Badge variant="outline">{badge}</Badge>}
        </div>
      </button>
      {isOpen && (
        <div className="px-6 py-4 border-t">
          {children}
        </div>
      )}
    </Card>
  );
};

interface NotificationProps {
  type: 'info' | 'warning' | 'error' | 'success';
  message: string;
  onDismiss?: () => void;
}

export const Notification: React.FC<NotificationProps> = ({
  type,
  message,
  onDismiss,
}) => {
  const colors = {
    info: 'bg-blue-50 text-blue-800 border-blue-200',
    warning: 'bg-yellow-50 text-yellow-800 border-yellow-200',
    error: 'bg-red-50 text-red-800 border-red-200',
    success: 'bg-green-50 text-green-800 border-green-200',
  };

  const icons = {
    info: '📋',
    warning: '⚠️',
    error: '❌',
    success: '✅',
  };

  return (
    <div className={`border p-4 rounded-lg flex items-center justify-between ${colors[type]}`}>
      <div className="flex items-center space-x-2">
        <span>{icons[type]}</span>
        <p className="text-sm">{message}</p>
      </div>
      {onDismiss && (
        <button
          onClick={onDismiss}
          className="ml-2 font-bold cursor-pointer hover:opacity-70"
        >
          ×
        </button>
      )}
    </div>
  );
};

interface SkeletonProps {
  className?: string;
}

export const Skeleton: React.FC<SkeletonProps> = ({ className = '' }) => (
  <div className={`animate-pulse bg-muted rounded ${className}`} />
);

interface RetryableProps {
  onRetry: () => Promise<void>;
  maxRetries?: number;
  children: React.ReactNode;
}

export const Retryable: React.FC<RetryableProps> = ({
  onRetry,
  maxRetries = 3,
  children,
}) => {
  const [retrying, setRetrying] = React.useState(false);
  const [retryCount, setRetryCount] = React.useState(0);

  const handleRetry = async () => {
    if (retryCount >= maxRetries) return;
    
    setRetrying(true);
    try {
      await onRetry();
      setRetryCount(0);
    } catch (error) {
      setRetryCount(retryCount + 1);
    } finally {
      setRetrying(false);
    }
  };

  return (
    <div className="space-y-2">
      {children}
      {retryCount > 0 && retryCount < maxRetries && (
        <div className="text-xs text-muted-foreground">
          Retry attempt {retryCount} of {maxRetries}
        </div>
      )}
      {retryCount >= maxRetries && (
        <p className="text-xs text-destructive">Max retries reached</p>
      )}
      <button
        onClick={handleRetry}
        disabled={retrying || retryCount >= maxRetries}
        className="text-sm text-primary hover:underline disabled:opacity-50"
      >
        {retrying ? 'Retrying...' : 'Retry'}
      </button>
    </div>
  );
};
