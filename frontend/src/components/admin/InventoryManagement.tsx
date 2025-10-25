import React, { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { 
  Package, 
  Search, 
  Edit, 
  AlertTriangle,
  TrendingUp,
  TrendingDown
} from 'lucide-react';

interface InventoryItem {
  id: string;
  productId: string;
  productName: string;
  variantId?: string;
  variantName?: string;
  quantityAvailable: number;
  quantityReserved: number;
  warehouseLocation: string;
  lastUpdated: string;
}

interface InventoryUpdateRequest {
  productId: string;
  variantId?: string;
  quantity: number;
  location: string;
  operation: 'add' | 'subtract' | 'set';
}

const InventoryManagement: React.FC = () => {
  const [inventory, setInventory] = useState<InventoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [showUpdateForm, setShowUpdateForm] = useState(false);
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null);
  const [updateForm, setUpdateForm] = useState<InventoryUpdateRequest>({
    productId: '',
    variantId: undefined,
    quantity: 0,
    location: '',
    operation: 'set'
  });

  useEffect(() => {
    fetchInventory();
  }, []);

  const fetchInventory = async () => {
    try {
      setLoading(true);
      const response = await fetch('/api/v1/admin/inventory', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to fetch inventory');
      }

      const data = await response.json();
      setInventory(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load inventory');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateInventory = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const response = await fetch('/api/v1/admin/inventory/update', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        },
        body: JSON.stringify(updateForm)
      });

      if (!response.ok) {
        throw new Error('Failed to update inventory');
      }

      await fetchInventory();
      resetUpdateForm();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update inventory');
    }
  };

  const handleEdit = (item: InventoryItem) => {
    setEditingItem(item);
    setUpdateForm({
      productId: item.productId,
      variantId: item.variantId,
      quantity: item.quantityAvailable,
      location: item.warehouseLocation,
      operation: 'set'
    });
    setShowUpdateForm(true);
  };

  const resetUpdateForm = () => {
    setUpdateForm({
      productId: '',
      variantId: undefined,
      quantity: 0,
      location: '',
      operation: 'set'
    });
    setEditingItem(null);
    setShowUpdateForm(false);
  };

  const getStockStatus = (item: InventoryItem) => {
    const available = item.quantityAvailable - item.quantityReserved;
    if (available === 0) return { status: 'out_of_stock', color: 'bg-red-100 text-red-800' };
    if (available < 10) return { status: 'low_stock', color: 'bg-yellow-100 text-yellow-800' };
    if (available > 100) return { status: 'overstock', color: 'bg-blue-100 text-blue-800' };
    return { status: 'in_stock', color: 'bg-green-100 text-green-800' };
  };

  const getStockIcon = (item: InventoryItem) => {
    const available = item.quantityAvailable - item.quantityReserved;
    if (available === 0) return <AlertTriangle className="h-4 w-4 text-red-500" />;
    if (available < 10) return <TrendingDown className="h-4 w-4 text-yellow-500" />;
    if (available > 100) return <TrendingUp className="h-4 w-4 text-blue-500" />;
    return <Package className="h-4 w-4 text-green-500" />;
  };

  const filteredInventory = inventory.filter(item =>
    item.productName.toLowerCase().includes(searchTerm.toLowerCase()) ||
    item.warehouseLocation.toLowerCase().includes(searchTerm.toLowerCase())
  );

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
        <h1 className="text-3xl font-bold">Inventory Management</h1>
        <Button onClick={() => setShowUpdateForm(true)}>
          <Edit className="h-4 w-4 mr-2" />
          Update Inventory
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      {/* Search */}
      <Card>
        <CardContent className="pt-6">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
            <Input
              placeholder="Search inventory..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
        </CardContent>
      </Card>

      {/* Update Form */}
      {showUpdateForm && (
        <Card>
          <CardHeader>
            <CardTitle>
              {editingItem ? 'Update Inventory Item' : 'Add Inventory Update'}
            </CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleUpdateInventory} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <Label htmlFor="productId">Product ID</Label>
                  <Input
                    id="productId"
                    value={updateForm.productId}
                    onChange={(e) => setUpdateForm({ ...updateForm, productId: e.target.value })}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="variantId">Variant ID (optional)</Label>
                  <Input
                    id="variantId"
                    value={updateForm.variantId || ''}
                    onChange={(e) => setUpdateForm({ ...updateForm, variantId: e.target.value || undefined })}
                  />
                </div>
                <div>
                  <Label htmlFor="quantity">Quantity</Label>
                  <Input
                    id="quantity"
                    type="number"
                    value={updateForm.quantity}
                    onChange={(e) => setUpdateForm({ ...updateForm, quantity: parseInt(e.target.value) })}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="location">Warehouse Location</Label>
                  <Input
                    id="location"
                    value={updateForm.location}
                    onChange={(e) => setUpdateForm({ ...updateForm, location: e.target.value })}
                    required
                  />
                </div>
                <div>
                  <Label htmlFor="operation">Operation</Label>
                  <Select value={updateForm.operation} onValueChange={(value: any) => setUpdateForm({ ...updateForm, operation: value })}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="set">Set Quantity</SelectItem>
                      <SelectItem value="add">Add to Quantity</SelectItem>
                      <SelectItem value="subtract">Subtract from Quantity</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
              <div className="flex gap-2">
                <Button type="submit">
                  {editingItem ? 'Update Inventory' : 'Apply Update'}
                </Button>
                <Button type="button" variant="outline" onClick={resetUpdateForm}>
                  Cancel
                </Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {/* Inventory List */}
      <Card>
        <CardHeader>
          <CardTitle>Inventory Levels ({filteredInventory.length})</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {filteredInventory.map((item) => {
              const stockStatus = getStockStatus(item);
              const available = item.quantityAvailable - item.quantityReserved;
              
              return (
                <div
                  key={item.id}
                  className="flex items-center justify-between p-4 border rounded-lg"
                >
                  <div className="flex items-center gap-3">
                    {getStockIcon(item)}
                    <div>
                      <h3 className="font-medium">{item.productName}</h3>
                      {item.variantName && (
                        <p className="text-sm text-muted-foreground">
                          Variant: {item.variantName}
                        </p>
                      )}
                      <p className="text-sm text-muted-foreground">
                        Location: {item.warehouseLocation}
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="flex items-center gap-2 mb-1">
                      <Badge className={stockStatus.color}>
                        {stockStatus.status.replace('_', ' ')}
                      </Badge>
                    </div>
                    <p className="text-sm">
                      Available: <span className="font-medium">{available}</span>
                    </p>
                    <p className="text-sm text-muted-foreground">
                      Total: {item.quantityAvailable} | Reserved: {item.quantityReserved}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      size="sm"
                      variant="outline"
                      onClick={() => handleEdit(item)}
                    >
                      <Edit className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              );
            })}
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default InventoryManagement;
