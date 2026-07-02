import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Heart, Search } from 'lucide-react';
import { useFavoritesStore } from '@/stores/favoritesStore';
import { usePageTitle } from '@/hooks';
import {
  PageGrid, Chamber, CornerBrace, TrustSeal,
  SealedButton, FrameButton, AnnotationTag,
} from '@/components/containment';

import './styles.css';

export default function FavoritesPage() {
  usePageTitle('Favorites');
  const navigate = useNavigate();
  const [loading, setLoading] = useState(true);
  const { favorites, total, fetchFavorites, isLoading, error } = useFavoritesStore();

  useEffect(() => {
    const load = async () => { setLoading(true); await fetchFavorites(1, 50); setLoading(false); };
    load();
  }, [fetchFavorites]);

  if (loading || isLoading) {
    return (
      <div className="fav-page">
        <PageGrid />
        <Chamber className="fav-hero">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="fav-hero__header">
            <div className="fav-hero__title-row">
              <TrustSeal size="lg" />
              <h1 className="fav-hero__title">Your Favorites</h1>
            </div>
          </div>
        </Chamber>
        <div className="fav-grid">
          {[...Array(6)].map((_, i) => <div key={i} className="fav-skeleton" />)}
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="fav-page">
        <PageGrid />
        <Chamber className="fav-hero">
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <div className="fav-hero__header">
            <div className="fav-hero__title-row">
              <TrustSeal size="lg" />
              <h1 className="fav-hero__title">Your Favorites</h1>
            </div>
          </div>
        </Chamber>
        <Chamber className="fav-empty">
          <Heart className="fav-empty__icon" />
          <h2 className="fav-empty__title">Something went wrong</h2>
          <p className="fav-empty__desc">{error}</p>
          <SealedButton onClick={() => fetchFavorites(1, 50)}>Try Again</SealedButton>
        </Chamber>
      </div>
    );
  }

  if (favorites.length === 0) {
    return (
      <div className="fav-page">
        <PageGrid />
        <Chamber className="fav-hero" ribs>
          <CornerBrace position="tl" />
          <CornerBrace position="br" />
          <AnnotationTag primary="MODULE FAV-01" secondary="Favorites" position="top-right" />
          <div className="fav-hero__header">
            <div className="fav-hero__title-row">
              <TrustSeal size="lg" />
              <h1 className="fav-hero__title">Your Favorites</h1>
            </div>
            <p className="fav-hero__subtitle">Functions you favorite will appear here for quick access</p>
          </div>
        </Chamber>
        <Chamber className="fav-empty">
          <Heart className="fav-empty__icon" />
          <h2 className="fav-empty__title">No favorites yet</h2>
          <p className="fav-empty__desc">Start exploring and add functions to your favorites by clicking the heart icon.</p>
          <div className="fav-empty__actions">
            <FrameButton onClick={() => navigate('/marketplace?type=functions')} iconLeft={<Search className="fav-icon-sm" />}>
              Discover Functions
            </FrameButton>
            <FrameButton onClick={() => navigate('/functions/new')}>Create Your Own</FrameButton>
          </div>
        </Chamber>
      </div>
    );
  }

  return (
    <div className="fav-page">
      <PageGrid />
      <Chamber className="fav-hero" ribs>
        <CornerBrace position="tl" />
        <CornerBrace position="br" />
        <AnnotationTag primary="MODULE FAV-01" secondary="Favorites" position="top-right" />
        <div className="fav-hero__header">
          <div className="fav-hero__title-row">
            <TrustSeal size="lg" />
            <h1 className="fav-hero__title">Your Favorites</h1>
          </div>
          <p className="fav-hero__subtitle">
            {total} {total === 1 ? 'function' : 'functions'} in your favorites
          </p>
        </div>
      </Chamber>

      <div className="fav-grid">
        {favorites.map((fav) => (
          <button key={fav.function_id} className="fav-card" onClick={() => navigate(`/functions/${fav.function_id}`)}>
            <div className="fav-card__icon-wrap">
              <Heart className="fav-card__icon" />
            </div>
            <div className="fav-card__info">
              <h3 className="fav-card__title">Function</h3>
              <p className="fav-card__meta">Added {new Date(fav.created_at).toLocaleDateString()}</p>
            </div>
          </button>
        ))}
      </div>
    </div>
  );
}
