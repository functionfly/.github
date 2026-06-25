import { adminApiClient } from '@/lib/api/adminClient';
import { API_ROUTES } from '@/lib/constants';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { Globe, MapPin, Search, ThumbsUp, ThumbsDown, X } from 'lucide-react';
import { useState } from 'react';

interface CityAliasEntry {
  alias: string;
  source: string;
  created: string;
}

interface CityReviewEntry {
  city_id: number;
  slug: string;
  name: string;
  state_code: string;
  state_name: string;
  country_code: string;
  country_name: string;
  latitude: number;
  longitude: number;
  population: number;
  metro_name?: string;
  created_at: string;
  alias_count: number;
}

interface CityReviewDetail {
  city_id: number;
  slug: string;
  name: string;
  state_code: string;
  state_name: string;
  country_code: string;
  country_name: string;
  latitude: number;
  longitude: number;
  population: number;
  review_status: string;
  reviewed_at?: string;
  reviewed_by?: string;
  review_notes?: string;
  metro_name?: string;
  created_at: string;
  aliases: CityAliasEntry[];
}

interface CityReviewListResponse {
  total: number;
  limit: number;
  offset: number;
  cities: CityReviewEntry[];
}

interface ReviewCityRequest {
  status: 'approved' | 'rejected';
  notes?: string;
}

function unwrap<T>(response: { data?: T } | T): T {
  return (response as { data?: T }).data !== undefined
    ? (response as { data: T }).data
    : (response as T);
}

export function AdminCityReviewsPage() {
  const queryClient = useQueryClient();
  const [selectedCityId, setSelectedCityId] = useState<number | null>(null);
  const [reviewNotes, setReviewNotes] = useState('');
  const [showAll, setShowAll] = useState(false);

  const { data: pendingCities, isLoading: pendingLoading } = useQuery<CityReviewListResponse>({
    queryKey: ['city-reviews-pending'],
    queryFn: async () => {
      const res = await adminApiClient.get<CityReviewListResponse>(API_ROUTES.ADMIN_CITIES_PENDING);
      return unwrap(res);
    },
  });

  const { data: allCities, isLoading: allLoading } = useQuery<any>({
    queryKey: ['city-reviews-all'],
    queryFn: async () => {
      const res = await adminApiClient.get<any>('/cities');
      return unwrap(res);
    },
    enabled: showAll,
  });

  const { data: cityDetail } = useQuery<CityReviewDetail>({
    queryKey: ['city-review-detail', selectedCityId],
    queryFn: async () => {
      const res = await adminApiClient.get<CityReviewDetail>(`/cities/${selectedCityId}/review`);
      return unwrap(res);
    },
    enabled: selectedCityId !== null,
  });

  const reviewMutation = useMutation({
    mutationFn: async ({ cityId, data }: { cityId: number; data: ReviewCityRequest }) => {
      console.log('Mutation called with cityId:', cityId, 'data:', data);
      return adminApiClient.post(`/cities/${cityId}/review`, data);
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['city-reviews-pending'] });
      setSelectedCityId(null);
      setReviewNotes('');
      alert('City reviewed successfully!');
    },
    onError: (error: any) => {
      console.error('Review mutation error:', error?.response?.data || error.message);
      alert(`Failed to review city: ${error?.response?.data?.error || error.message}`);
    },
  });

  const handleApprove = () => {
    console.log('Approve clicked, selectedCityId:', selectedCityId);
    if (selectedCityId === null) return;
    reviewMutation.mutate({ cityId: selectedCityId, data: { status: 'approved', notes: reviewNotes } });
  };

  const handleReject = () => {
    if (selectedCityId === null) return;
    reviewMutation.mutate({ cityId: selectedCityId, data: { status: 'rejected', notes: reviewNotes } });
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-gray-100">Cities</h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Manage city locations and review geocoded additions
          </p>
        </div>
        <div className="flex items-center gap-4">
          <button
            onClick={() => setShowAll(!showAll)}
            className={`px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
              showAll
                ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
            }`}
          >
            {showAll ? 'Show Pending Only' : 'Show All Cities'}
          </button>
          <div className="flex items-center gap-2 px-3 py-1.5 bg-amber-100 dark:bg-amber-900/30 rounded-full">
            <span className="text-amber-800 dark:text-amber-400 text-sm font-medium">
              {pendingCities?.total ?? 0} pending
            </span>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
          <div className="p-4 border-b border-gray-200 dark:border-gray-700">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <input
                type="text"
                placeholder="Search cities..."
                className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg dark:bg-gray-700 dark:text-gray-100"
              />
            </div>
          </div>

          {pendingLoading ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">Loading...</div>
          ) : showAll && allCities ? (
            allCities.cities.length === 0 ? (
              <div className="p-8 text-center text-gray-500 dark:text-gray-400">No cities found</div>
            ) : (
              <div className="divide-y divide-gray-200 dark:divide-gray-700 max-h-[600px] overflow-y-auto">
                {allCities.cities.map((city: any) => (
                  <button
                    key={city.city_id}
                    onClick={() => setSelectedCityId(city.city_id)}
                    className={`w-full p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors ${
                      selectedCityId === city.city_id ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                    }`}
                  >
                    <div className="flex items-start justify-between">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">
                            {city.name || city.slug}
                          </h3>
                          <span className={`text-xs px-2 py-0.5 rounded ${
                            city.review_status === 'pending' ? 'bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400' :
                            city.review_status === 'approved' ? 'bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400' :
                            'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400'
                          }`}>
                            {city.review_status}
                          </span>
                        </div>
                        <div className="flex items-center gap-1.5 mt-1 text-xs text-gray-500 dark:text-gray-400">
                          <Globe className="w-3 h-3" />
                          <span>{city.country_name}</span>
                        </div>
                      </div>
                    </div>
                  </button>
                ))}
              </div>
            )
          ) : pendingCities?.cities.length === 0 ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              No cities pending review
            </div>
          ) : (
            <div className="divide-y divide-gray-200 dark:divide-gray-700 max-h-[600px] overflow-y-auto">
              {pendingCities?.cities.map((city) => (
                <button
                  key={city.city_id}
                  onClick={() => setSelectedCityId(city.city_id)}
                  className={`w-full p-4 text-left hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors ${
                    selectedCityId === city.city_id ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                  }`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <h3 className="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">
                          {city.name}
                        </h3>
                        <span className="text-xs px-2 py-0.5 bg-amber-100 dark:bg-amber-900/30 text-amber-700 dark:text-amber-400 rounded">
                          pending
                        </span>
                      </div>
                      <div className="flex items-center gap-1.5 mt-1 text-xs text-gray-500 dark:text-gray-400">
                        <Globe className="w-3 h-3" />
                        <span>{city.country_name}</span>
                        {city.state_code && (
                          <>
                            <span>·</span>
                            <span>{city.state_code}</span>
                          </>
                        )}
                      </div>
                      <div className="flex items-center gap-1.5 mt-1 text-xs text-gray-500 dark:text-gray-400">
                        <MapPin className="w-3 h-3" />
                        <span>
                          {city.latitude.toFixed(4)}, {city.longitude.toFixed(4)}
                        </span>
                      </div>
                    </div>
                    <div className="text-xs text-gray-400 dark:text-gray-500">
                      {new Date(city.created_at).toLocaleDateString()}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
          {selectedCityId === null ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">
              Select a city to review
            </div>
          ) : !cityDetail ? (
            <div className="p-8 text-center text-gray-500 dark:text-gray-400">Loading...</div>
          ) : (
            <div className="p-6 space-y-6">
              <div className="flex items-start justify-between">
                <div>
                  <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">{cityDetail.name}</h2>
                  <p className="text-sm text-gray-500 dark:text-gray-400">{cityDetail.slug}</p>
                </div>
                <button
                  onClick={() => setSelectedCityId(null)}
                  className="p-1 hover:bg-gray-100 dark:hover:bg-gray-700 rounded"
                >
                  <X className="w-5 h-5 text-gray-400" />
                </button>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-1">
                  <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Country
                  </label>
                  <p className="text-sm text-gray-900 dark:text-gray-100">
                    {cityDetail.country_name} ({cityDetail.country_code})
                  </p>
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    State
                  </label>
                  <p className="text-sm text-gray-900 dark:text-gray-100">
                    {cityDetail.state_name || cityDetail.state_code || 'N/A'}
                  </p>
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Coordinates
                  </label>
                  <p className="text-sm text-gray-900 dark:text-gray-100">
                    {cityDetail.latitude.toFixed(6)}, {cityDetail.longitude.toFixed(6)}
                  </p>
                </div>
                <div className="space-y-1">
                  <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                    Population
                  </label>
                  <p className="text-sm text-gray-900 dark:text-gray-100">
                    {cityDetail.population > 0 ? cityDetail.population.toLocaleString() : 'Unknown'}
                  </p>
                </div>
                {cityDetail.metro_name && (
                  <div className="space-y-1 col-span-2">
                    <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                      Metro Area
                    </label>
                    <p className="text-sm text-gray-900 dark:text-gray-100">{cityDetail.metro_name}</p>
                  </div>
                )}
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                  Aliases ({cityDetail.aliases.length})
                </label>
                <div className="flex flex-wrap gap-2">
                  {cityDetail.aliases.map((alias, i) => (
                    <span
                      key={i}
                      className="text-xs px-2 py-1 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded"
                      title={`Source: ${alias.source}`}
                    >
                      {alias.alias}
                    </span>
                  ))}
                </div>
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">
                  Review Notes
                </label>
                <textarea
                  value={reviewNotes}
                  onChange={(e) => setReviewNotes(e.target.value)}
                  rows={3}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg dark:bg-gray-700 dark:text-gray-100 text-sm"
                  placeholder="Add notes about this review..."
                />
              </div>

              <div className="flex gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
                <button
                  onClick={handleApprove}
                  disabled={reviewMutation.isPending}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50"
                >
                  <ThumbsUp className="w-4 h-4" />
                  Approve
                </button>
                <button
                  onClick={handleReject}
                  disabled={reviewMutation.isPending}
                  className="flex-1 flex items-center justify-center gap-2 px-4 py-2.5 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50"
                >
                  <ThumbsDown className="w-4 h-4" />
                  Reject
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default AdminCityReviewsPage;
