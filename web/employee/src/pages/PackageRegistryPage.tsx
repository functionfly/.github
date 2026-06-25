import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { packageRegistryApi, type Package, type PackageVersion } from '@/api/package_registry';
import { Package as PackageIcon, Plus, Search, Download, ExternalLink, Tag, Box } from 'lucide-react';

const registryColors: Record<string, string> = {
  npm: 'bg-red-500/20 text-red-400',
  cargo: 'bg-orange-500/20 text-orange-400',
  pypi: 'bg-blue-500/20 text-blue-400',
  docker: 'bg-cyan-500/20 text-cyan-400',
  maven: 'bg-yellow-500/20 text-yellow-400',
};

export function PackageRegistryPage() {
  const queryClient = useQueryClient();
  const [search, setSearch] = useState('');
  const [typeFilter, setTypeFilter] = useState('');
  const [selectedPkg, setSelectedPkg] = useState<string | null>(null);
  const [showPublish, setShowPublish] = useState(false);
  const [form, setForm] = useState({
    name: '',
    scope: '',
    description: '',
    registry_type: 'npm',
    repository_url: '',
  });

  const { data, isLoading } = useQuery({
    queryKey: ['packages', typeFilter],
    queryFn: () => packageRegistryApi.list(typeFilter ? { registry_type: typeFilter } : undefined),
  });

  const { data: detailData } = useQuery({
    queryKey: ['package', selectedPkg],
    queryFn: () => packageRegistryApi.get(selectedPkg!),
    enabled: !!selectedPkg,
  });

  const { data: versionsData } = useQuery({
    queryKey: ['package-versions', selectedPkg],
    queryFn: () => packageRegistryApi.listVersions(selectedPkg!),
    enabled: !!selectedPkg,
  });

  const publishMutation = useMutation({
    mutationFn: (data: Partial<Package>) => packageRegistryApi.publish(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['packages'] });
      setShowPublish(false);
      setForm({ name: '', scope: '', description: '', registry_type: 'npm', repository_url: '' });
    },
  });

  const packages = (data?.data?.packages || []).filter((p) =>
    !search ||
    p.name.toLowerCase().includes(search.toLowerCase()) ||
    (p.description && p.description.toLowerCase().includes(search.toLowerCase())) ||
    (p.scope && p.scope.toLowerCase().includes(search.toLowerCase()))
  );

  const selectedPkgData = detailData?.data?.pkg;
  const versions: PackageVersion[] = versionsData?.data?.versions || [];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <PackageIcon className="h-6 w-6 text-emerald-400" />
          <h1 className="text-2xl font-bold">Package Registry</h1>
        </div>
        <button
          onClick={() => setShowPublish(true)}
          className="flex items-center gap-2 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700"
        >
          <Plus className="h-4 w-4" />
          Publish Package
        </button>
      </div>

      <div className="flex items-center gap-3">
        <Search className="h-4 w-4 text-gray-400" />
        <input
          type="text"
          placeholder="Search packages..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-64 rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
        />
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value)}
          className="rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
        >
          <option value="">All Types</option>
          <option value="npm">npm</option>
          <option value="cargo">cargo</option>
          <option value="pypi">pypi</option>
          <option value="docker">docker</option>
          <option value="maven">maven</option>
        </select>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : packages.length === 0 ? (
        <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
          <PackageIcon className="mb-4 h-12 w-12 text-gray-600" />
          <p className="text-gray-400">{search ? 'No packages match your search' : 'No packages published'}</p>
        </div>
      ) : (
        <div className="space-y-3">
          {packages.map((pkg) => (
            <button
              key={pkg.id}
              onClick={() => setSelectedPkg(pkg.id === selectedPkg ? null : pkg.id)}
              className={`w-full rounded-xl border p-4 text-left transition-colors ${
                pkg.id === selectedPkg
                  ? 'border-emerald-600 bg-gray-800'
                  : 'border-gray-800 bg-gray-900 hover:bg-gray-800'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-emerald-500/10">
                    <Box className="h-5 w-5 text-emerald-400" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <h3 className="font-medium text-gray-100">
                        {pkg.scope ? `${pkg.scope}/` : ''}{pkg.name}
                      </h3>
                      <span className={`rounded-full px-2 py-0.5 text-xs ${registryColors[pkg.registry_type] || 'bg-gray-500/20 text-gray-400'}`}>
                        {pkg.registry_type}
                      </span>
                      {pkg.is_internal && (
                        <span className="rounded-full bg-purple-500/20 px-2 py-0.5 text-xs text-purple-400">internal</span>
                      )}
                    </div>
                    {pkg.description && <p className="mt-0.5 text-sm text-gray-500">{pkg.description}</p>}
                    <div className="mt-1 flex items-center gap-3 text-xs text-gray-500">
                      {pkg.latest_version && (
                        <span className="flex items-center gap-1"><Tag className="h-3 w-3" />v{pkg.latest_version}</span>
                      )}
                      <span className="flex items-center gap-1"><Download className="h-3 w-3" />{pkg.total_downloads.toLocaleString()} downloads</span>
                      {pkg.published_at && <span>{new Date(pkg.published_at).toLocaleDateString()}</span>}
                    </div>
                  </div>
                </div>
              </div>

              {pkg.id === selectedPkg && selectedPkgData && (
                <div className="mt-4 border-t border-gray-800 pt-4">
                  <div className="grid grid-cols-2 gap-4 sm:grid-cols-4">
                    <div className="rounded-lg bg-gray-800 p-3">
                      <span className="text-xs text-gray-500">Registry</span>
                      <p className="mt-1 text-sm font-medium text-gray-100">{selectedPkgData.registry_type}</p>
                    </div>
                    <div className="rounded-lg bg-gray-800 p-3">
                      <span className="text-xs text-gray-500">Latest</span>
                      <p className="mt-1 text-sm font-medium text-gray-100">{selectedPkgData.latest_version || '-'}</p>
                    </div>
                    <div className="rounded-lg bg-gray-800 p-3">
                      <span className="text-xs text-gray-500">Downloads</span>
                      <p className="mt-1 text-sm font-medium text-gray-100">{selectedPkgData.total_downloads.toLocaleString()}</p>
                    </div>
                    <div className="rounded-lg bg-gray-800 p-3">
                      <span className="text-xs text-gray-500">Type</span>
                      <p className="mt-1 text-sm font-medium text-gray-100">{selectedPkgData.is_internal ? 'Internal' : 'Public'}</p>
                    </div>
                  </div>

                  {selectedPkgData.repository_url && (
                    <div className="mt-3">
                      <a
                        href={selectedPkgData.repository_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="flex items-center gap-1 text-sm text-emerald-400 hover:text-emerald-300"
                      >
                        <ExternalLink className="h-3 w-3" />
                        Repository
                      </a>
                    </div>
                  )}

                  <div className="mt-4">
                    <h4 className="mb-2 text-sm font-medium text-gray-300">Versions</h4>
                    {versions.length === 0 ? (
                      <p className="text-sm text-gray-500">No versions available</p>
                    ) : (
                      <div className="space-y-1">
                        {versions.map((v) => (
                          <div key={v.id} className="flex items-center justify-between rounded-lg bg-gray-800 px-3 py-2">
                            <div className="flex items-center gap-2">
                              <Tag className="h-3 w-3 text-gray-500" />
                              <span className="text-sm font-medium text-gray-100">v{v.version}</span>
                              {v.description && <span className="text-xs text-gray-500">{v.description}</span>}
                            </div>
                            <div className="flex items-center gap-3 text-xs text-gray-500">
                              <span className="flex items-center gap-1"><Download className="h-3 w-3" />{v.downloads.toLocaleString()}</span>
                              <span>{new Date(v.published_at).toLocaleDateString()}</span>
                            </div>
                          </div>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              )}
            </button>
          ))}
        </div>
      )}

      {showPublish && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl bg-gray-900 p-6">
            <h2 className="mb-4 text-lg font-semibold">Publish Package</h2>
            <input
              type="text"
              placeholder="Package name"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              autoFocus
            />
            <input
              type="text"
              placeholder="Scope (optional, e.g. @myorg)"
              value={form.scope}
              onChange={(e) => setForm({ ...form, scope: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <textarea
              placeholder="Description"
              value={form.description}
              onChange={(e) => setForm({ ...form, description: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
              rows={2}
            />
            <select
              value={form.registry_type}
              onChange={(e) => setForm({ ...form, registry_type: e.target.value })}
              className="mb-3 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100"
            >
              <option value="npm">npm</option>
              <option value="cargo">cargo</option>
              <option value="pypi">pypi</option>
              <option value="docker">docker</option>
              <option value="maven">maven</option>
            </select>
            <input
              type="text"
              placeholder="Repository URL (optional)"
              value={form.repository_url}
              onChange={(e) => setForm({ ...form, repository_url: e.target.value })}
              className="mb-4 w-full rounded-lg border border-gray-700 bg-gray-800 px-3 py-2 text-sm text-gray-100 placeholder-gray-500"
            />
            <div className="flex justify-end gap-3">
              <button onClick={() => setShowPublish(false)} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-gray-200">Cancel</button>
              <button
                onClick={() => publishMutation.mutate({
                  name: form.name,
                  scope: form.scope || undefined,
                  description: form.description || undefined,
                  registry_type: form.registry_type,
                  repository_url: form.repository_url || undefined,
                })}
                disabled={!form.name.trim()}
                className="rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50"
              >
                Publish
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
