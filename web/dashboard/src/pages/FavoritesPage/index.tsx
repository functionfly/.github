import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Heart, Search } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { useFavoritesStore } from '@/stores/favoritesStore';

export default function FavoritesPage() {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);

  const { favorites, total, fetchFavorites, isLoading, error } = useFavoritesStore();

  useEffect(() => {
    const load = async () => {
      setLoading(true);
      await fetchFavorites(1, 50);
      setLoading(false);
    };
    load();
  }, [fetchFavorites]);

  const handleCreateFunction = () => {
    navigate('/functions/new');
  };

  const handleGoToDiscovery = () => {
    navigate('/functions/discovery/hot');
  };

  if (loading || isLoading) {
    return (
      <div className="favorites-page">
        <div className="favorites-page-header">
          <div className="favorites-page-title">
            <Heart className="h-6 w-6 text-red-500" />
            <h1>Your Favorites</h1>
          </div>
        </div>
        <div className="favorites-loading">
          {[...Array(6)].map((_, i) => (
            <div key={i} className="favorites-skeleton" />
          ))}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="favorites-page">
        <div className="favorites-page-header">
          <div className="favorites-page-title">
            <Heart className="h-6 w-6 text-red-500" />
            <h1>Your Favorites</h1>
          </div>
        </div>
        <div className="favorites-empty-state">
          <Heart className="h-16 w-16 text-gray-400" />
          <h2>Something went wrong</h2>
          <p>{error}</p>
          <Button className="btn-primary" onClick={() => fetchFavorites(1, 50)}>Try Again</Button>
        </div>
      </div>
    );
  }

  if (favorites.length === 0) {
    return (
      <div className="favorites-page">
        <div className="favorites-page-header">
          <div className="favorites-page-title">
            <Heart className="h-6 w-6 text-red-500" />
            <h1>Your Favorites</h1>
          </div>
          <p className="favorites-page-subtitle">
            Functions you favorite will appear here for quick access
          </p>
        </div>
        <div className="favorites-empty-state">
          <div className="favorites-empty-icon">
            <Heart className="h-16 w-16 text-gray-400" />
          </div>
          <h2 className="favorites-empty-title">No favorites yet</h2>
          <p className="favorites-empty-description">
            Start exploring and add functions to your favorites by clicking the heart icon.
          </p>
          <div className="favorites-empty-actions">
            <Button variant="secondary" onClick={handleGoToDiscovery}>
              <Search className="h-4 w-4 mr-2" />
              Discover Functions
            </Button>
            <Button variant="secondary" onClick={handleCreateFunction}>
              Create Your Own
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="favorites-page">
      <div className="favorites-page-header">
        <div className="favorites-page-title">
          <Heart className="h-6 w-6 text-red-500 fill-current" />
          <h1>Your Favorites</h1>
        </div>
        <p className="favorites-page-subtitle">
          {total} {total === 1 ? 'function' : 'functions'} in your favorites
        </p>
      </div>

      <div className="favorites-grid">
        {favorites.map((fav) => (
          <div
            key={fav.function_id}
            className="favorite-card"
            onClick={() => navigate(`/functions/${fav.function_id}`)}
          >
            <div className="favorite-card-icon">
              <Heart className="h-5 w-5 text-red-500 fill-current" />
            </div>
            <div className="favorite-card-info">
              <h3 className="favorite-card-title">Function</h3>
              <p className="favorite-card-meta">
                Added {new Date(fav.created_at).toLocaleDateString()}
              </p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
