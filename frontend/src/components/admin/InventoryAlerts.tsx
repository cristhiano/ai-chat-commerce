import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { 
  AlertTriangle, 
  CheckCircle, 
  Filter,
  RefreshCw
} from 'lucide-react';

interface InventoryAlert {
  id: string;
  productId: string;
  productName: string;
  variantId?: string;
  variantName?: string;
  currentQuantity: number;
  threshold: number;
  location: string;
  alertType: 'low_stock' | 'out_of_stock' | 'overstock';
  createdAt: string;
  isRead: boolean;
}

const InventoryAlerts: React.FC = () => {
  const [alerts, setAlerts] = useState<InventoryAlert[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<string>('all');
  const [readFilter, setReadFilter] = useState<string>('unread');

  useEffect(() => {
    fetchAlerts();
  }, [readFilter]);

  const fetchAlerts = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (readFilter === 'read') {
        params.append('is_read', 'true');
      } else if (readFilter === 'unread') {
        params.append('is_read', 'false');
      }

      const response = await fetch(`/api/v1/admin/inventory/alerts?${params}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch alerts');
      }

      const data = await response.json();
      setAlerts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load alerts');
    } finally {
      setLoading(false);
    }
  };

  const markAsRead = async (alertId: string) => {
    try {
      const response = await fetch(`/api/v1/admin/inventory/alerts/${alertId}/read`, {
        method: 'PATCH',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to mark alert as read');
      }

      setAlerts(alerts.map(alert => 
        alert.id === alertId ? { ...alert, isRead: true } : alert
      ));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark alert as read');
    }
  };

  const markAllAsRead = async () => {
    try {
      const unreadAlerts = alerts.filter(alert => !alert.isRead);
      await Promise.all(unreadAlerts.map(alert => markAsRead(alert.id)));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to mark all alerts as read');
    }
  };

  const getAlertIcon = (alertType: string) => {
    switch (alertType) {
      case 'low_stock':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case 'out_of_stock':
        return <AlertTriangle className="h-4 w-4 text-red-500" />;
      case 'overstock':
        return <AlertTriangle className="h-4 w-4 text-blue-500" />;
      default:
        return <AlertTriangle className="h-4 w-4 text-gray-500" />;
    }
  };

  const getAlertBadgeColor = (alertType: string) => {
    switch (alertType) {
      case 'low_stock':
        return 'bg-yellow-100 text-yellow-800';
      case 'out_of_stock':
        return 'bg-red-100 text-red-800';
      case 'overstock':
        return 'bg-blue-100 text-blue-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getPriority = (alertType: string) => {
    switch (alertType) {
      case 'out_of_stock':
        return 1;
      case 'low_stock':
        return 2;
      case 'overstock':
        return 3;
      default:
        return 4;
    }
  };

  const filteredAlerts = alerts
    .filter(alert => {
      if (filter === 'all') return true;
      return alert.alertType === filter;
    })
    .sort((a, b) => {
      // Sort by priority (out of stock first), then by date
      const priorityA = getPriority(a.alertType);
      const priorityB = getPriority(b.alertType);
      if (priorityA !== priorityB) {
        return priorityA - priorityB;
      }
      return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
    });

  const unreadCount = alerts.filter(alert => !alert.isRead).length;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold">Inventory Alerts</h1>
          {unreadCount > 0 && (
            <p className="text-muted-foreground mt-1">
              {unreadCount} unread alert{unreadCount !== 1 ? 's' : ''}
            </p>
          )}
        </div>
        <div className="flex gap-2">
          {unreadCount > 0 && (
            <Button onClick={markAllAsRead} variant="outline">
              <CheckCircle className="h-4 w-4 mr-2" />
              Mark All as Read
            </Button>
          )}
          <Button onClick={fetchAlerts} variant="outline">
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      {/* Filters */}
      <Card>
        <CardContent className="pt-6">
          <div className="flex gap-4">
            <div className="flex items-center gap-2">
              <Filter className="h-4 w-4" />
              <Select value={filter} onValueChange={setFilter}>
                <SelectTrigger className="w-48">
                  <SelectValue placeholder="Filter by type" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">All Types</SelectItem>
                  <SelectItem value="low_stock">Low Stock</SelectItem>
                  <SelectItem value="out_of_stock">Out of Stock</SelectItem>
                  <SelectItem value="overstock">Overstock</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <Select value={readFilter} onValueChange={setReadFilter}>
              <SelectTrigger className="w-48">
                <SelectValue placeholder="Filter by status" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="unread">Unread Only</SelectItem>
                <SelectItem value="read">Read Only</SelectItem>
                <SelectItem value="all">All Alerts</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </CardContent>
      </Card>

      {/* Alerts List */}
      <Card>
        <CardHeader>
          <CardTitle>Alerts ({filteredAlerts.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {filteredAlerts.length === 0 ? (
            <div className="text-center py-8">
              <AlertTriangle className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <p className="text-muted-foreground">
                {readFilter === 'unread' ? 'No unread alerts' : 'No alerts found'}
              </p>
            </div>
          ) : (
            <div className="space-y-3">
              {filteredAlerts.map((alert) => (
                <div
                  key={alert.id}
                  className={`flex items-center justify-between p-4 border rounded-lg ${
                    !alert.isRead ? 'bg-yellow-50 border-yellow-200' : ''
                  }`}
                >
                  <div className="flex items-center gap-3">
                    {getAlertIcon(alert.alertType)}
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <h3 className="font-medium">{alert.productName}</h3>
                        {alert.variantName && (
                          <span className="text-sm text-muted-foreground">
                            ({alert.variantName})
                          </span>
                        )}
                        <Badge className={getAlertBadgeColor(alert.alertType)}>
                          {alert.alertType.replace('_', ' ')}
                        </Badge>
                        {!alert.isRead && (
                          <Badge variant="outline" className="bg-yellow-100 text-yellow-800">
                            New
                          </Badge>
                        )}
                      </div>
                      <p className="text-sm text-muted-foreground">
                        Current: {alert.currentQuantity} | Threshold: {alert.threshold} | {alert.location}
                      </p>
                      <p className="text-xs text-muted-foreground">
                        {new Date(alert.createdAt).toLocaleString()}
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    {!alert.isRead && (
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => markAsRead(alert.id)}
                      >
                        <CheckCircle className="h-4 w-4 mr-1" />
                        Mark as Read
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
};

export default InventoryAlerts;
