import { useState, useRef, useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import { orgImportApi, type OrgChartImport } from '@/api/org_import';
import { Upload, FileSpreadsheet, AlertTriangle, CheckCircle, Download, Loader2 } from 'lucide-react';

export function OrgImportPage() {
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [importResult, setImportResult] = useState<OrgChartImport | null>(null);
  const [pollingId, setPollingId] = useState<string | null>(null);

  const uploadMutation = useMutation({
    mutationFn: (file: File) => orgImportApi.upload(file),
    onSuccess: (data) => {
      const imp = data.data.import;
      setImportResult(imp);
      if (imp.status === 'processing') {
        setPollingId(imp.id);
      }
    },
  });

  useEffect(() => {
    if (!pollingId) return;
    const interval = setInterval(async () => {
      try {
        const res = await orgImportApi.getStatus(pollingId);
        const imp = res.data.import;
        setImportResult(imp);
        if (imp.status !== 'processing') {
          setPollingId(null);
        }
      } catch {
        // ignore polling errors
      }
    }, 2000);
    return () => clearInterval(interval);
  }, [pollingId]);

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (file) {
      uploadMutation.mutate(file);
    }
  }

  function handleDrop(e: React.DragEvent) {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file) {
      uploadMutation.mutate(file);
    }
  }

  const imp = importResult;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Upload className="h-6 w-6 text-indigo-400" />
          <h1 className="text-2xl font-bold">Org Chart Import</h1>
        </div>
        <a
          href="/templates/orgchart-import-template.csv"
          download
          className="flex items-center gap-2 rounded-lg bg-gray-800 px-4 py-2 text-sm text-gray-300 hover:bg-gray-700"
        >
          <Download className="h-4 w-4" />
          Download Template
        </a>
      </div>

      <div
        onDragOver={(e) => e.preventDefault()}
        onDrop={handleDrop}
        className="flex flex-col items-center justify-center rounded-xl border-2 border-dashed border-gray-700 bg-gray-900 p-12 transition-colors hover:border-indigo-500"
      >
        <FileSpreadsheet className="mb-4 h-12 w-12 text-gray-600" />
        <p className="mb-2 text-gray-300">Drop your CSV or Excel file here</p>
        <p className="mb-4 text-sm text-gray-500">or click to browse</p>
        <input
          ref={fileInputRef}
          type="file"
          accept=".csv,.xlsx,.xls"
          onChange={handleFileChange}
          className="hidden"
        />
        <button
          onClick={() => fileInputRef.current?.click()}
          disabled={uploadMutation.isPending}
          className="flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
        >
          {uploadMutation.isPending ? (
            <Loader2 className="h-4 w-4 animate-spin" />
          ) : (
            <Upload className="h-4 w-4" />
          )}
          {uploadMutation.isPending ? 'Uploading...' : 'Select File'}
        </button>
      </div>

      {uploadMutation.isError && (
        <div className="rounded-xl border border-red-800 bg-red-900/20 p-4">
          <div className="flex items-center gap-2 text-red-400">
            <AlertTriangle className="h-4 w-4" />
            <span className="text-sm font-medium">Upload failed</span>
          </div>
          <p className="mt-1 text-sm text-red-300">{(uploadMutation.error as Error)?.message || 'Unknown error'}</p>
        </div>
      )}

      {imp && (
        <div className="rounded-xl border border-gray-800 bg-gray-900 p-6">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-lg font-semibold text-gray-100">Import Status</h2>
            <span className={`rounded-full px-3 py-1 text-xs font-medium ${
              imp.status === 'completed' ? 'bg-green-500/20 text-green-400' :
              imp.status === 'processing' ? 'bg-blue-500/20 text-blue-400' :
              imp.status === 'failed' ? 'bg-red-500/20 text-red-400' :
              'bg-gray-500/20 text-gray-400'
            }`}>
              {imp.status}
            </span>
          </div>

          <div className="mb-4 grid grid-cols-4 gap-4">
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">File</span>
              <p className="mt-1 text-sm font-medium text-gray-100">{imp.file_name}</p>
            </div>
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">Total Rows</span>
              <p className="mt-1 text-sm font-medium text-gray-100">{imp.total_rows}</p>
            </div>
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">Processed</span>
              <p className="mt-1 text-sm font-medium text-green-400">{imp.processed_rows}</p>
            </div>
            <div className="rounded-lg bg-gray-800 p-3">
              <span className="text-xs text-gray-500">Errors</span>
              <p className="mt-1 text-sm font-medium text-red-400">{imp.error_rows}</p>
            </div>
          </div>

          {imp.status === 'processing' && imp.total_rows > 0 && (
            <div className="mb-4">
              <div className="mb-1 flex justify-between text-xs text-gray-500">
                <span>Progress</span>
                <span>{Math.round((imp.processed_rows / imp.total_rows) * 100)}%</span>
              </div>
              <div className="h-2 w-full overflow-hidden rounded-full bg-gray-700">
                <div
                  className="h-full rounded-full bg-blue-500 transition-all"
                  style={{ width: `${(imp.processed_rows / imp.total_rows) * 100}%` }}
                />
              </div>
            </div>
          )}

          {imp.status === 'completed' && imp.error_rows === 0 && (
            <div className="flex items-center gap-2 rounded-lg bg-green-500/10 p-3 text-green-400">
              <CheckCircle className="h-4 w-4" />
              <span className="text-sm">All rows imported successfully</span>
            </div>
          )}

          {imp.errors.length > 0 && (
            <div>
              <h3 className="mb-2 text-sm font-medium text-gray-300">Errors</h3>
              <div className="max-h-60 overflow-y-auto space-y-1">
                {imp.errors.map((err, i) => (
                  <div key={i} className="flex items-start gap-2 rounded-lg bg-red-500/5 px-3 py-2">
                    <AlertTriangle className="mt-0.5 h-3 w-3 flex-shrink-0 text-red-400" />
                    <span className="text-xs text-red-300">
                      <span className="font-medium">Row {err.row}:</span> {err.message}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
