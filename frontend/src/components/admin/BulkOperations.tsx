import React, { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Progress } from '@/components/ui/progress';
import { 
  Upload, 
  Download, 
  FileText, 
  CheckCircle, 
  AlertCircle,
  Info
} from 'lucide-react';

interface BulkOperationResult {
  success: boolean;
  message: string;
  processed: number;
  failed: number;
  errors?: string[];
}

const BulkOperations: React.FC = () => {
  const [importFile, setImportFile] = useState<File | null>(null);
  const [importProgress, setImportProgress] = useState(0);
  const [importResult, setImportResult] = useState<BulkOperationResult | null>(null);
  const [exportFormat, setExportFormat] = useState<'csv' | 'json'>('csv');
  const [exportResult, setExportResult] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (file) {
      setImportFile(file);
      setImportResult(null);
      setError(null);
    }
  };

  const handleImport = async () => {
    if (!importFile) return;

    try {
      setLoading(true);
      setImportProgress(0);
      setError(null);

      const formData = new FormData();
      formData.append('file', importFile);

      // Simulate progress
      const progressInterval = setInterval(() => {
        setImportProgress(prev => {
          if (prev >= 90) {
            clearInterval(progressInterval);
            return prev;
          }
          return prev + 10;
        });
      }, 200);

      const response = await fetch('/api/v1/admin/products/import', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        },
        body: formData
      });

      clearInterval(progressInterval);
      setImportProgress(100);

      if (!response.ok) {
        throw new Error('Failed to import products');
      }

      const result = await response.json();
      setImportResult(result);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to import products');
    } finally {
      setLoading(false);
    }
  };

  const handleExport = async () => {
    try {
      setLoading(true);
      setError(null);

      const response = await fetch(`/api/v1/admin/products/export?format=${exportFormat}`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('adminToken')}`
        }
      });

      if (!response.ok) {
        throw new Error('Failed to export products');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `products.${exportFormat}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      setExportResult('Export completed successfully');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to export products');
    } finally {
      setLoading(false);
    }
  };

  const downloadTemplate = () => {
    const csvContent = `name,description,price,category_id,sku,status,tags,metadata
"Sample Product","A sample product description",29.99,"category-uuid","SKU001","active","tag1,tag2","{\"weight\":\"1kg\",\"dimensions\":\"10x10x10\"}"`;
    
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'product_template.csv';
    document.body.appendChild(a);
    a.click();
    window.URL.revokeObjectURL(url);
    document.body.removeChild(a);
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold">Bulk Operations</h1>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Import Section */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Upload className="h-5 w-5" />
              Import Products
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="import-file">Select CSV File</Label>
              <Input
                id="import-file"
                type="file"
                accept=".csv"
                onChange={handleFileSelect}
                className="mt-1"
              />
              {importFile && (
                <p className="text-sm text-muted-foreground mt-1">
                  Selected: {importFile.name} ({(importFile.size / 1024).toFixed(1)} KB)
                </p>
              )}
            </div>

            <div className="flex gap-2">
              <Button onClick={downloadTemplate} variant="outline">
                <FileText className="h-4 w-4 mr-2" />
                Download Template
              </Button>
              <Button 
                onClick={handleImport} 
                disabled={!importFile || loading}
                className="flex-1"
              >
                <Upload className="h-4 w-4 mr-2" />
                Import Products
              </Button>
            </div>

            {importProgress > 0 && (
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Importing...</span>
                  <span>{importProgress}%</span>
                </div>
                <Progress value={importProgress} className="w-full" />
              </div>
            )}

            {importResult && (
              <div className={`p-3 rounded-lg ${
                importResult.success 
                  ? 'bg-green-50 border border-green-200 text-green-700'
                  : 'bg-red-50 border border-red-200 text-red-700'
              }`}>
                <div className="flex items-center gap-2 mb-2">
                  {importResult.success ? (
                    <CheckCircle className="h-4 w-4" />
                  ) : (
                    <AlertCircle className="h-4 w-4" />
                  )}
                  <span className="font-medium">{importResult.message}</span>
                </div>
                <p className="text-sm">
                  Processed: {importResult.processed} | Failed: {importResult.failed}
                </p>
                {importResult.errors && importResult.errors.length > 0 && (
                  <div className="mt-2">
                    <p className="text-sm font-medium">Errors:</p>
                    <ul className="text-sm list-disc list-inside">
                      {importResult.errors.map((error, index) => (
                        <li key={index}>{error}</li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            )}
          </CardContent>
        </Card>

        {/* Export Section */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Download className="h-5 w-5" />
              Export Products
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <Label htmlFor="export-format">Export Format</Label>
              <select
                id="export-format"
                value={exportFormat}
                onChange={(e) => setExportFormat(e.target.value as 'csv' | 'json')}
                className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="csv">CSV</option>
                <option value="json">JSON</option>
              </select>
            </div>

            <Button 
              onClick={handleExport} 
              disabled={loading}
              className="w-full"
            >
              <Download className="h-4 w-4 mr-2" />
              Export Products
            </Button>

            {exportResult && (
              <div className="bg-green-50 border border-green-200 text-green-700 p-3 rounded-lg">
                <div className="flex items-center gap-2">
                  <CheckCircle className="h-4 w-4" />
                  <span className="font-medium">{exportResult}</span>
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Instructions */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            Instructions
          </CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <h3 className="font-medium mb-2">Import Format</h3>
              <p className="text-sm text-muted-foreground mb-2">
                Use CSV format with the following columns:
              </p>
              <div className="bg-gray-50 p-3 rounded text-sm font-mono">
                name, description, price, category_id, sku, status, tags, metadata
              </div>
            </div>

            <div>
              <h3 className="font-medium mb-2">Required Fields</h3>
              <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
                <li><strong>name</strong> - Product name (required)</li>
                <li><strong>price</strong> - Product price as decimal (required)</li>
                <li><strong>category_id</strong> - Valid category UUID (required)</li>
                <li><strong>sku</strong> - Unique product SKU (required)</li>
              </ul>
            </div>

            <div>
              <h3 className="font-medium mb-2">Optional Fields</h3>
              <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
                <li><strong>description</strong> - Product description</li>
                <li><strong>status</strong> - active, inactive, or draft (default: active)</li>
                <li><strong>tags</strong> - Comma-separated tags</li>
                <li><strong>metadata</strong> - JSON object as string</li>
              </ul>
            </div>

            <div>
              <h3 className="font-medium mb-2">Tips</h3>
              <ul className="text-sm text-muted-foreground list-disc list-inside space-y-1">
                <li>Download the template to see the correct format</li>
                <li>Ensure category_id values exist in the system</li>
                <li>SKU values must be unique</li>
                <li>Metadata should be valid JSON if provided</li>
                <li>Large files may take several minutes to process</li>
              </ul>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default BulkOperations;
