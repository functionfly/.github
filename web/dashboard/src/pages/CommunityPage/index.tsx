import { useState } from 'react';
import { usePageTitle } from '@/hooks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Card, CardContent } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar';
import { Separator } from '@/components/ui/separator';
import {
  MessageSquare,
  Share2,
  Bookmark,
  MoreHorizontal,
  ArrowUp,
  ArrowDown,
  Image,
  Link2,
  Send,
  TrendingUp,
  Users,
  Flame,
  Star,
  Clock,
  Eye,
  ChevronRight,
} from 'lucide-react';
import './CommunityPage.css';

interface Post {
  id: string;
  author: {
    name: string;
    username: string;
    avatar: string;
    badge?: string;
  };
  community: string;
  communityIcon: string;
  timestamp: string;
  title: string;
  content: string;
  image?: string;
  upvotes: number;
  comments: number;
  shares: number;
  isUpvoted?: boolean;
  isDownvoted?: boolean;
  isSaved?: boolean;
  viewCount: string;
}

interface TrendingTopic {
  name: string;
  posts: number;
  trend: 'hot' | 'new' | 'top';
}

const mockPosts: Post[] = [
  {
    id: '1',
    author: {
      name: 'Sarah Chen',
      username: 'sarahchen',
      avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=100&h=100&fit=crop',
      badge: 'Developer',
    },
    community: 'FunctionFly',
    communityIcon: '🚀',
    timestamp: '2 hours ago',
    title: 'Just deployed my first AI agent using FunctionFly - here\'s what I learned',
    content: 'After weeks of experimentation, I finally got my autonomous agent running in production. The key insight was breaking down complex tasks into smaller sub-agents that communicate via State Fabric. Here are my top 3 takeaways...',
    image: 'https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800&h=400&fit=crop',
    upvotes: 847,
    comments: 156,
    shares: 42,
    isUpvoted: true,
    viewCount: '12.4K',
  },
  {
    id: '2',
    author: {
      name: 'Marcus Rodriguez',
      username: 'marcusdev',
      avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=100&h=100&fit=crop',
      badge: 'Early Adopter',
    },
    community: 'Showcase',
    communityIcon: '✨',
    timestamp: '4 hours ago',
    title: 'Built a real-time collaboration tool using WebSockets and FunctionFly agents',
    content: 'Created a multiplayer whiteboard where AI agents can draw alongside humans. Used the agent communication API to sync state across all connected clients. The latency is under 50ms!',
    upvotes: 623,
    comments: 89,
    shares: 28,
    viewCount: '8.9K',
  },
  {
    id: '3',
    author: {
      name: 'Alex Kim',
      username: 'alexkim',
      avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&h=100&fit=crop',
    },
    community: 'Help',
    communityIcon: '❓',
    timestamp: '6 hours ago',
    title: 'Best practices for agent memory management?',
    content: 'Looking for advice on how to handle long-term memory for agents. Should I use vector databases or something simpler? What have others found works well at scale?',
    upvotes: 234,
    comments: 67,
    shares: 5,
    viewCount: '3.2K',
  },
  {
    id: '4',
    author: {
      name: 'Emma Watson',
      username: 'emmacode',
      avatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=100&h=100&fit=crop',
      badge: 'Contributor',
    },
    community: 'Discussion',
    communityIcon: '💬',
    timestamp: '8 hours ago',
    title: 'The future of agent-to-agent communication standards',
    content: 'As more platforms emerge, I think we need to establish common protocols for agents to talk to each other. Just like HTTP standardized web communication, we need something similar for AI agents. Thoughts?',
    upvotes: 1567,
    comments: 312,
    shares: 156,
    isSaved: true,
    viewCount: '45.2K',
  },
  {
    id: '5',
    author: {
      name: 'David Park',
      username: 'davidpark',
      avatar: 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=100&h=100&fit=crop',
    },
    community: 'News',
    communityIcon: '📰',
    timestamp: '12 hours ago',
    title: 'FunctionFly announces new enterprise features for 2026',
    content: 'Just read the blog post - the new SLA guarantees and dedicated support tiers look promising for teams at scale. Also excited about the improved monitoring dashboard.',
    upvotes: 892,
    comments: 134,
    shares: 67,
    viewCount: '28.7K',
  },
];

const trendingTopics: TrendingTopic[] = [
  { name: 'agent-orchestration', posts: 234, trend: 'hot' },
  { name: 'web3-integration', posts: 156, trend: 'new' },
  { name: 'production-deployments', posts: 89, trend: 'top' },
  { name: 'memory-management', posts: 67, trend: 'hot' },
  { name: 'multi-agent-systems', posts: 45, trend: 'new' },
];

const communities = [
  { name: 'FunctionFly', icon: '🚀', members: '12.4K' },
  { name: 'Help', icon: '❓', members: '8.2K' },
  { name: 'Showcase', icon: '✨', members: '6.7K' },
  { name: 'Discussion', icon: '💬', members: '5.1K' },
  { name: 'News', icon: '📰', members: '3.8K' },
];

export function CommunityPage() {
  usePageTitle('Community');
  const [posts, setPosts] = useState<Post[]>(mockPosts);
  const [newPostContent, setNewPostContent] = useState('');
  const [newPostTitle, setNewPostTitle] = useState('');
  const [activeFeed, setActiveFeed] = useState<'hot' | 'new' | 'top'>('hot');
  const [showCompose, setShowCompose] = useState(false);

  const handleVote = (postId: string, type: 'up' | 'down') => {
    setPosts((prev) =>
      prev.map((post) => {
        if (post.id !== postId) return post;
        
        let upvoteDelta = 0;
        let downvoteDelta = 0;
        let newIsUpvoted = post.isUpvoted;
        let newIsDownvoted = post.isDownvoted;
        
        if (type === 'up') {
          if (post.isUpvoted) {
            newIsUpvoted = false;
            upvoteDelta = -1;
          } else {
            newIsUpvoted = true;
            upvoteDelta = 1;
            if (post.isDownvoted) {
              newIsDownvoted = false;
              downvoteDelta = -1;
            }
          }
        } else {
          if (post.isDownvoted) {
            newIsDownvoted = false;
            downvoteDelta = -1;
          } else {
            newIsDownvoted = true;
            downvoteDelta = 1;
            if (post.isUpvoted) {
              newIsUpvoted = false;
              upvoteDelta = -1;
            }
          }
        }
        
        return {
          ...post,
          upvotes: post.upvotes + upvoteDelta + downvoteDelta,
          isUpvoted: newIsUpvoted,
          isDownvoted: newIsDownvoted,
        };
      })
    );
  };

  const handleSave = (postId: string) => {
    setPosts((prev) =>
      prev.map((post) =>
        post.id === postId ? { ...post, isSaved: !post.isSaved } : post
      )
    );
  };

  const handleSubmitPost = () => {
    if (!newPostTitle.trim() || !newPostContent.trim()) return;
    
    const newPost: Post = {
      id: Date.now().toString(),
      author: {
        name: 'You',
        username: 'you',
        avatar: 'https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&h=100&fit=crop',
      },
      community: 'FunctionFly',
      communityIcon: '🚀',
      timestamp: 'just now',
      title: newPostTitle,
      content: newPostContent,
      upvotes: 1,
      comments: 0,
      shares: 0,
      isUpvoted: true,
      viewCount: '0',
    };
    
    setPosts((prev) => [newPost, ...prev]);
    setNewPostTitle('');
    setNewPostContent('');
    setShowCompose(false);
  };

  const getTrendIcon = (trend: TrendingTopic['trend']) => {
    switch (trend) {
      case 'hot':
        return <Flame className="w-3.5 h-3.5 text-orange-500" />;
      case 'new':
        return <Clock className="w-3.5 h-3.5 text-blue-500" />;
      case 'top':
        return <Star className="w-3.5 h-3.5 text-yellow-500" />;
    }
  };

  return (
    <div className="community-page">
      <div className="community-container">
        {/* Main Feed */}
        <div className="community-main">
          {/* Compose Area */}
          <Card className="community-compose-card">
            <CardContent className="p-4">
              {!showCompose ? (
                <button 
                  className="community-compose-trigger"
                  onClick={() => setShowCompose(true)}
                >
                  <Avatar className="w-10 h-10">
                    <AvatarImage src="https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=100&h=100&fit=crop" />
                    <AvatarFallback>Y</AvatarFallback>
                  </Avatar>
                  <span className="text-muted-foreground">Start a discussion...</span>
                </button>
              ) : (
                <div className="community-compose-form">
                  <Input
                    placeholder="Post title"
                    value={newPostTitle}
                    onChange={(e) => setNewPostTitle(e.target.value)}
                    className="mb-3"
                  />
                  <Textarea
                    placeholder="What's on your mind?"
                    value={newPostContent}
                    onChange={(e) => setNewPostContent(e.target.value)}
                    rows={4}
                    className="mb-3"
                  />
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      <Button variant="ghost" size="sm">
                        <Image className="w-4 h-4 mr-1" />
                        Image
                      </Button>
                      <Button variant="ghost" size="sm">
                        <Link2 className="w-4 h-4 mr-1" />
                        Link
                      </Button>
                    </div>
                    <div className="flex items-center gap-2">
                      <Button 
                        variant="ghost" 
                        size="sm"
                        onClick={() => setShowCompose(false)}
                      >
                        Cancel
                      </Button>
                      <Button 
                        size="sm"
                        onClick={handleSubmitPost}
                        disabled={!newPostTitle.trim() || !newPostContent.trim()}
                      >
                        <Send className="w-4 h-4 mr-1" />
                        Post
                      </Button>
                    </div>
                  </div>
                </div>
              )}
            </CardContent>
          </Card>

          {/* Feed Tabs */}
          <div className="community-feed-tabs">
            <button
              className={`community-feed-tab ${activeFeed === 'hot' ? 'active' : ''}`}
              onClick={() => setActiveFeed('hot')}
            >
              <Flame className="w-4 h-4" />
              Hot
            </button>
            <button
              className={`community-feed-tab ${activeFeed === 'new' ? 'active' : ''}`}
              onClick={() => setActiveFeed('new')}
            >
              <Clock className="w-4 h-4" />
              New
            </button>
            <button
              className={`community-feed-tab ${activeFeed === 'top' ? 'active' : ''}`}
              onClick={() => setActiveFeed('top')}
            >
              <TrendingUp className="w-4 h-4" />
              Top
            </button>
          </div>

          {/* Posts List */}
          <div className="community-posts">
            {posts.map((post) => (
              <Card key={post.id} className="community-post-card">
                <CardContent className="p-0">
                  <div className="community-post-layout">
                    {/* Vote Column */}
                    <div className="community-post-votes">
                      <button
                        className={`community-vote-btn upvote ${post.isUpvoted ? 'active' : ''}`}
                        onClick={() => handleVote(post.id, 'up')}
                        aria-label="Upvote"
                      >
                        <ArrowUp className="w-5 h-5" />
                      </button>
                      <span className="community-vote-count">{post.upvotes}</span>
                      <button
                        className={`community-vote-btn downvote ${post.isDownvoted ? 'active' : ''}`}
                        onClick={() => handleVote(post.id, 'down')}
                        aria-label="Downvote"
                      >
                        <ArrowDown className="w-5 h-5" />
                      </button>
                    </div>

                    {/* Content Column */}
                    <div className="community-post-content">
                      {/* Post Header */}
                      <div className="community-post-header">
                        <Avatar className="w-6 h-6">
                          <AvatarImage src={post.author.avatar} />
                          <AvatarFallback>{post.author.name[0]}</AvatarFallback>
                        </Avatar>
                        <span className="community-post-author">{post.author.name}</span>
                        {post.author.badge && (
                          <Badge variant="secondary" className="community-author-badge">
                            {post.author.badge}
                          </Badge>
                        )}
                        <span className="text-muted-foreground">•</span>
                        <span className="community-post-community">
                          <span>{post.communityIcon}</span>
                          {post.community}
                        </span>
                        <span className="text-muted-foreground">•</span>
                        <span className="text-muted-foreground text-sm">{post.timestamp}</span>
                      </div>

                      {/* Post Title */}
                      <h3 className="community-post-title">{post.title}</h3>

                      {/* Post Body */}
                      <p className="community-post-body">{post.content}</p>

                      {/* Post Image */}
                      {post.image && (
                        <div className="community-post-image-container">
                          <img src={post.image} alt="" className="community-post-image" />
                        </div>
                      )}

                      {/* Post Actions */}
                      <div className="community-post-actions">
                        <button className="community-action-btn">
                          <MessageSquare className="w-4 h-4" />
                          <span>{post.comments} Comments</span>
                        </button>
                        <button className="community-action-btn">
                          <Share2 className="w-4 h-4" />
                          <span>{post.shares} Share</span>
                        </button>
                        <button 
                          className={`community-action-btn ${post.isSaved ? 'saved' : ''}`}
                          onClick={() => handleSave(post.id)}
                        >
                          <Bookmark className="w-4 h-4" />
                          <span>{post.isSaved ? 'Saved' : 'Save'}</span>
                        </button>
                        <button className="community-action-btn">
                          <MoreHorizontal className="w-4 h-4" />
                        </button>
                      </div>

                      {/* Quick Comments Preview */}
                      {post.comments > 0 && (
                        <div className="community-comments-preview">
                          <div className="community-comment">
                            <Avatar className="w-5 h-5">
                              <AvatarImage src="https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=50&h=50&fit=crop" />
                              <AvatarFallback>A</AvatarFallback>
                            </Avatar>
                            <div className="community-comment-content">
                              <span className="community-comment-author">alice_dev</span>
                              <span className="community-comment-text">This is exactly what I was looking for! Thanks for sharing...</span>
                            </div>
                          </div>
                          <button className="community-view-comments">
                            View all {post.comments} comments
                            <ChevronRight className="w-4 h-4" />
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>

        {/* Sidebar */}
        <div className="community-sidebar">
          {/* Community Info Card */}
          <Card className="community-info-card">
            <CardContent className="p-4">
              <div className="community-info-header">
                <div className="community-info-icon">🚀</div>
                <div>
                  <h3 className="community-info-title">FunctionFly Community</h3>
                  <p className="text-sm text-muted-foreground">The official community</p>
                </div>
              </div>
              <div className="community-info-stats">
                <div className="community-info-stat">
                  <Users className="w-4 h-4" />
                  <span>12.4K members</span>
                </div>
                <div className="community-info-stat">
                  <Eye className="w-4 h-4" />
                  <span>Online: 234</span>
                </div>
              </div>
              <Button className="w-full mt-4">Join Community</Button>
            </CardContent>
          </Card>

          {/* Trending Topics */}
          <Card className="community-trending-card">
            <CardContent className="p-4">
              <h3 className="community-sidebar-title flex items-center gap-2">
                <TrendingUp className="w-4 h-4" />
                Trending Topics
              </h3>
              <div className="community-trending-list">
                {trendingTopics.map((topic, index) => (
                  <div key={topic.name} className="community-trending-item">
                    <span className="community-trending-rank">{index + 1}</span>
                    <div className="community-trending-content">
                      <span className="community-trending-name">#{topic.name}</span>
                      <div className="community-trending-meta">
                        {getTrendIcon(topic.trend)}
                        <span>{topic.posts} posts</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Quick Communities */}
          <Card className="community-communities-card">
            <CardContent className="p-4">
              <h3 className="community-sidebar-title flex items-center gap-2">
                <Users className="w-4 h-4" />
                Communities
              </h3>
              <div className="community-communities-list">
                {communities.map((community) => (
                  <button key={community.name} className="community-communities-item">
                    <span className="community-communities-icon">{community.icon}</span>
                    <span className="community-communities-name">{community.name}</span>
                    <span className="text-xs text-muted-foreground">{community.members}</span>
                  </button>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Rules Card */}
          <Card className="community-rules-card">
            <CardContent className="p-4">
              <h3 className="community-sidebar-title flex items-center gap-2">
                <Star className="w-4 h-4" />
                Community Rules
              </h3>
              <ol className="community-rules-list">
                <li>Be respectful and constructive</li>
                <li>No spam or self-promotion</li>
                <li>Use appropriate post flairs</li>
                <li>Search before posting</li>
                <li>Share knowledge freely</li>
              </ol>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}

export default CommunityPage;
