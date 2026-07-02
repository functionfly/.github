import { usePageTitle } from '@/hooks';
import { Chamber, PageGrid } from '@/components/containment';
import { Loader2, Bell, ArrowLeft, CheckCircle2, MessageSquare } from 'lucide-react';
import { communityApi, formatRelativeTime, displayAuthor, type CommunityNotification } from '@/api/community';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { toast } from 'sonner';
import { useState } from 'react';

function NotificationItem({ notif }: { notif: CommunityNotification }) {
  const icon = notif.type === 'accepted' ? <CheckCircle2 size={16} /> : <MessageSquare size={16} />;
  const label = notif.type === 'accepted'
    ? 'accepted your answer'
    : 'replied to your post';
  const link = notif.post_slug ? `/community/${notif.post_slug}` : '#';

  return (
    <Link to={link} className={`sc-community-notif-item ${notif.is_read ? '' : 'sc-community-notif-unread'}`}>
      <div className="sc-community-notif-icon">{icon}</div>
      <div className="sc-community-notif-content">
        <p className="sc-community-notif-text">
          <strong>{displayAuthor(notif.actor)}</strong> {label}
          {notif.post_title && <> on <em>{notif.post_title}</em></>}
        </p>
        <span className="sc-community-notif-time">{formatRelativeTime(notif.created_at)}</span>
      </div>
    </Link>
  );
}

export function CommunityNotificationsPage() {
  usePageTitle('Notifications');
  const queryClient = useQueryClient();
  const [offset, setOffset] = useState(0);
  const limit = 20;

  const { data, isLoading } = useQuery({
    queryKey: ['community-notifications', offset],
    queryFn: () => communityApi.listNotifications(limit, offset),
  });

  const markReadMutation = useMutation({
    mutationFn: () => communityApi.markNotificationsRead(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['community-notifications'] });
      queryClient.invalidateQueries({ queryKey: ['community-unread-notifications'] });
      toast.success('All marked as read');
    },
  });

  const notifs = data?.notifications ?? [];
  const total = data?.total ?? 0;
  const hasMore = offset + limit < total;

  return (
    <div className="sc-community-page">
      <PageGrid />
      <div className="sc-thread-container">
        <Link to="/community" className="sc-thread-back">
          <ArrowLeft size={14} />
          <span>Back to community</span>
        </Link>

        <Chamber>
          <div style={{ padding: 'var(--space-5)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <div>
              <h1 style={{ fontFamily: 'var(--font-display)', fontSize: 22, fontWeight: 600, color: 'var(--text)', margin: 0 }}>
                <Bell size={20} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 8 }} />
                Notifications
              </h1>
              <p style={{ fontFamily: 'var(--font-body)', fontSize: 13, color: 'var(--text-dim)', marginTop: 'var(--space-2)' }}>
                {total} notification{total !== 1 ? 's' : ''}
              </p>
            </div>
            {total > 0 && (
              <button
                className="sc-community-sort-tab"
                onClick={() => markReadMutation.mutate()}
                disabled={markReadMutation.isPending}
              >
                Mark all read
              </button>
            )}
          </div>
        </Chamber>

        {isLoading ? (
          <Chamber nested className="sc-community-loading">
            <Loader2 size={24} className="sc-community-spinner" />
            <span className="sc-community-loading-text">Loading notifications...</span>
          </Chamber>
        ) : notifs.length === 0 ? (
          <Chamber nested className="sc-community-empty">
            <Bell size={40} className="sc-community-empty-icon" />
            <p className="sc-community-empty-title">No notifications</p>
            <p className="sc-community-empty-description">You'll be notified when someone replies to your posts.</p>
          </Chamber>
        ) : (
          <>
            <div className="sc-community-notif-list">
              {notifs.map((notif) => (
                <NotificationItem key={notif.id} notif={notif} />
              ))}
            </div>
            <div className="sc-community-pagination">
              {offset > 0 && (
                <button className="sc-community-sort-tab" onClick={() => setOffset(Math.max(0, offset - limit))}>
                  Previous
                </button>
              )}
              <span className="sc-community-pagination-info">
                {offset + 1}–{Math.min(offset + limit, total)} of {total}
              </span>
              {hasMore && (
                <button className="sc-community-sort-tab active" onClick={() => setOffset(offset + limit)}>
                  Load more
                </button>
              )}
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default CommunityNotificationsPage;
