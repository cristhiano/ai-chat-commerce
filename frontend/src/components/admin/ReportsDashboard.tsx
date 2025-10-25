import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { 
  BarChart3, 
  TrendingUp, 
  TrendingDown,
  Package,
  AlertTriangle,
  DollarSign,
  RefreshCw
} from 'lucide-react';

interface InventoryReport {
  totalProducts: number;
  totalQuantity: number;
  lowStockItems: number;
  outOfStockItems: number;
  overstockItems: number;
  reservedQuantity: number;
  availableQuantity: number;
}

interface SalesReport {
  totalOrders: number;
  totalRevenue: number;
  averageOrderValue: number;
  topSellingProducts: Array<{
    productId: string;
    productName: string;
    quantitySold: number;
    revenue: number;
  }>;
}

interface AlertSummary {
  totalAlerts: number;
  unreadAlerts: number;
  lowStockAlerts: number;
  outOfStockAlerts: number;
  overstockAlerts: number;
}

const ReportsDashboard: React.FC = () => {
  const [inventoryReport, setInventoryReport] = useState<InventoryReport | null>(null);
  const [salesReport, setSalesReport] = useState<SalesReport | null>(null);
  const [alertSummary, setAlertSummary] = useState<AlertSummary | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [timeRange, setTimeRange] = useState<string>('30d');

  useEffect(() => {
    fetchReports();
  }, [timeRange]);

  const fetchReports = async () => {
    try {
      setLoading(true);
      setError(null);

      const [inventoryResponse, salesResponse, alertsResponse] = await Promise.all([
        fetch('/api/v1/admin/inventory/report', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
          }
        }),
        fetch(`/api/v1/admin/sales/report?time_range=${timeRange}`, {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
          }
        }),
        fetch('/api/v1/admin/inventory/alerts/summary', {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
          }
        })
      ]);

      if (!inventoryResponse.ok) {
        throw new Error('Failed to fetch inventory report');
      }

      const inventoryData = await inventoryResponse.json();
      setInventoryReport(inventoryData);

      if (salesResponse.ok) {
        const salesData = await salesResponse.json();
        setSalesReport(salesData);
      }

      if (alertsResponse.ok) {
        const alertsData = await alertsResponse.json();
        setAlertSummary(alertsData);
      }

    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load reports');
    } finally {
      setLoading(false);
    }
  };

  const getTimeRangeLabel = (range: string) => {
    switch (range) {
      case '7d': return 'Last 7 days';
      case '30d': return 'Last 30 days';
      case '90d': return 'Last 90 days';
      case '1y': return 'Last year';
      default: return 'Last 30 days';
    }
  };

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
        <h1 className="text-3xl font-bold">Reports & Analytics</h1>
        <div className="flex gap-2">
          <Select value={timeRange} onValueChange={setTimeRange}>
            <SelectTrigger className="w-48">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="7d">Last 7 days</SelectItem>
              <SelectItem value="30d">Last 30 days</SelectItem>
              <SelectItem value="90d">Last 90 days</SelectItem>
              <SelectItem value="1y">Last year</SelectItem>
            </SelectContent>
          </Select>
          <Button onClick={fetchReports} variant="outline">
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

      {/* Overview Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Products</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {inventoryReport?.totalProducts || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Available Stock</CardTitle>
            <TrendingUp className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">
              {inventoryReport?.availableQuantity || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Low Stock Items</CardTitle>
            <AlertTriangle className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-yellow-600">
              {inventoryReport?.lowStockItems || 0}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Out of Stock</CardTitle>
            <TrendingDown className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold text-red-600">
              {inventoryReport?.outOfStockItems || 0}
            </div>
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Inventory Report */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <BarChart3 className="h-5 w-5" />
              Inventory Overview
            </CardTitle>
          </CardHeader>
          <CardContent>
            {inventoryReport ? (
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Total Quantity</span>
                  <span className="font-bold">{inventoryReport.totalQuantity}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Available</span>
                  <span className="font-bold text-green-600">{inventoryReport.availableQuantity}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Reserved</span>
                  <span className="font-bold text-blue-600">{inventoryReport.reservedQuantity}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Low Stock</span>
                  <span className="font-bold text-yellow-600">{inventoryReport.lowStockItems}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Out of Stock</span>
                  <span className="font-bold text-red-600">{inventoryReport.outOfStockItems}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Overstock</span>
                  <span className="font-bold text-purple-600">{inventoryReport.overstockItems}</span>
                </div>
              </div>
            ) : (
              <p className="text-muted-foreground text-center py-4">
                No inventory data available
              </p>
            )}
          </CardContent>
        </Card>

        {/* Sales Report */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <DollarSign className="h-5 w-5" />
              Sales Overview ({getTimeRangeLabel(timeRange)})
            </CardTitle>
          </CardHeader>
          <CardContent>
            {salesReport ? (
              <div className="space-y-4">
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Total Orders</span>
                  <span className="font-bold">{salesReport.totalOrders}</span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Total Revenue</span>
                  <span className="font-bold text-green-600">
                    ${salesReport.totalRevenue.toFixed(2)}
                  </span>
                </div>
                <div className="flex justify-between items-center">
                  <span className="text-sm font-medium">Average Order Value</span>
                  <span className="font-bold">
                    ${salesReport.averageOrderValue.toFixed(2)}
                  </span>
                </div>
                {salesReport.topSellingProducts.length > 0 && (
                  <div className="mt-4">
                    <h4 className="text-sm font-medium mb-2">Top Selling Products</h4>
                    <div className="space-y-2">
                      {salesReport.topSellingProducts.slice(0, 3).map((product) => (
                        <div key={product.productId} className="flex justify-between items-center text-sm">
                          <span className="truncate">{product.productName}</span>
                          <span className="font-medium">{product.quantitySold} sold</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            ) : (
              <p className="text-muted-foreground text-center py-4">
                No sales data available for {getTimeRangeLabel(timeRange)}
              </p>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Alert Summary */}
      {alertSummary && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5" />
              Alert Summary
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="text-center">
                <div className="text-2xl font-bold">{alertSummary.totalAlerts}</div>
                <div className="text-sm text-muted-foreground">Total Alerts</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-yellow-600">{alertSummary.unreadAlerts}</div>
                <div className="text-sm text-muted-foreground">Unread</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-yellow-600">{alertSummary.lowStockAlerts}</div>
                <div className="text-sm text-muted-foreground">Low Stock</div>
              </div>
              <div className="text-center">
                <div className="text-2xl font-bold text-red-600">{alertSummary.outOfStockAlerts}</div>
                <div className="text-sm text-muted-foreground">Out of Stock</div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};

export default ReportsDashboard;
