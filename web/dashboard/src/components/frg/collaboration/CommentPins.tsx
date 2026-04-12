/**
 * CommentPins Component
 * Show comment pins on the canvas with expandable threads
 */

import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  MessageCircle,
  X,
  Send,
  MoreHorizontal,
  Check,
  Trash2,
} from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';

interface Comment {
  id: string;
  author: string;
  content: string;
  timestamp: string;
  resolved: boolean;
}

interface CommentPin {
  id: string;
  x: number;
  y: number;
  comments: Comment[];
  nodeId?: string;
}

interface CommentPinsProps {
  pins?: CommentPin[];
  onAddComment?: (pinId: string, content: string) => void;
  onResolvePin?: (pinId: string) => void;
  onDeletePin?: (pinId: string) => void;
  className?: string;
}

function CommentThread({
  pin,
  isOpen,
  onClose,
  onAddComment,
  onResolve,
  onDelete,
}: {
  pin: CommentPin;
  isOpen: boolean;
  onClose: () => void;
  onAddComment: (content: string) => void;
  onResolve: () => void;
  onDelete: () => void;
}) {
  const [newComment, setNewComment] = useState('');
  const unresolvedCount = pin.comments.filter((c) => !c.resolved).length;

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (newComment.trim()) {
      onAddComment(newComment);
      setNewComment('');
    }
  };

  return (
    <AnimatePresence>
      {isOpen && (
        <motion.div
          initial={{ opacity: 0, scale: 0.9, y: 10 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.9, y: 10 }}
          className={cn(
            "absolute z-50",
            "left-8 top-0"
          )}
        >
          <Card className="w-72 border-[var(--border-subtle)] bg-[var(--bg-secondary)] shadow-xl">
            <CardHeader className="p-3 pb-2 flex flex-row items-center justify-between space-y-0">
              <div className="flex items-center gap-2">
                <CardTitle className="text-sm font-semibold">
                  {unresolvedCount} comment{unresolvedCount !== 1 ? 's' : ''}
                </CardTitle>
                {pin.nodeId && (
                  <Badge variant="secondary" className="text-[10px]">
                    Node {pin.nodeId.slice(0, 6)}
                  </Badge>
                )}
              </div>
              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={onResolve}
                  title="Mark as resolved"
                >
                  <Check className="w-3 h-3" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={onDelete}
                  title="Delete thread"
                >
                  <Trash2 className="w-3 h-3" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={onClose}
                >
                  <X className="w-3 h-3" />
                </Button>
              </div>
            </CardHeader>
            <CardContent className="p-3 pt-0 space-y-3">
              {/* Comments List */}
              <div className="space-y-2 max-h-40 overflow-auto">
                {pin.comments.map((comment) => (
                  <div
                    key={comment.id}
                    className={cn(
                      "p-2 rounded-lg text-sm",
                      comment.resolved
                        ? "bg-[var(--bg-tertiary)]/50 opacity-60"
                        : "bg-[var(--bg-tertiary)]"
                    )}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-medium text-xs">{comment.author}</span>
                      <span className="text-[10px] text-[var(--text-muted)]">
                        {new Date(comment.timestamp).toLocaleDateString()}
                      </span>
                    </div>
                    <p className="text-[var(--text-secondary)] text-xs">{comment.content}</p>
                  </div>
                ))}
              </div>

              {/* Add Comment */}
              <form onSubmit={handleSubmit} className="flex items-center gap-2">
                <Input
                  placeholder="Add a comment..."
                  value={newComment}
                  onChange={(e) => setNewComment(e.target.value)}
                  className="h-8 text-xs"
                />
                <Button type="submit" size="sm" className="h-8 px-2">
                  <Send className="w-3 h-3" />
                </Button>
              </form>
            </CardContent>
          </Card>
        </motion.div>
      )}
    </AnimatePresence>
  );
}

function PinMarker({
  pin,
  isSelected,
  onClick,
}: {
  pin: CommentPin;
  isSelected: boolean;
  onClick: () => void;
}) {
  const unresolvedCount = pin.comments.filter((c) => !c.resolved).length;

  return (
    <motion.button
      initial={{ scale: 0 }}
      animate={{ scale: 1 }}
      whileHover={{ scale: 1.1 }}
      whileTap={{ scale: 0.95 }}
      onClick={onClick}
      className={cn(
        "absolute flex items-center justify-center",
        "w-6 h-6 rounded-full shadow-lg",
        "transition-colors duration-200",
        isSelected
          ? "bg-brand-500 text-white ring-2 ring-brand-500/30"
          : "bg-[var(--bg-secondary)] text-[var(--text-primary)] hover:bg-brand-500 hover:text-white border border-[var(--border-subtle)]"
      )}
      style={{
        left: pin.x,
        top: pin.y,
      }}
    >
      <MessageCircle className="w-3 h-3" />
      {unresolvedCount > 0 && (
        <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-[10px] rounded-full flex items-center justify-center">
          {unresolvedCount}
        </span>
      )}
    </motion.button>
  );
}

// Mock data for demonstration
const mockPins: CommentPin[] = [
  {
    id: '1',
    x: 150,
    y: 200,
    nodeId: 'node-1',
    comments: [
      {
        id: 'c1',
        author: 'Alex',
        content: 'Should we add error handling here?',
        timestamp: '2026-04-10T10:00:00Z',
        resolved: false,
      },
      {
        id: 'c2',
        author: 'Sam',
        content: 'Good point, let me check the requirements',
        timestamp: '2026-04-10T10:05:00Z',
        resolved: false,
      },
    ],
  },
  {
    id: '2',
    x: 400,
    y: 350,
    comments: [
      {
        id: 'c3',
        author: 'Jordan',
        content: 'This connection looks incorrect',
        timestamp: '2026-04-10T09:30:00Z',
        resolved: true,
      },
    ],
  },
  {
    id: '3',
    x: 650,
    y: 150,
    nodeId: 'node-3',
    comments: [
      {
        id: 'c4',
        author: 'Alex',
        content: 'Need to update the schema for this input',
        timestamp: '2026-04-10T11:00:00Z',
        resolved: false,
      },
    ],
  },
];

export function CommentPins({
  pins = mockPins,
  onAddComment,
  onResolvePin,
  onDeletePin,
  className,
}: CommentPinsProps) {
  const [selectedPinId, setSelectedPinId] = useState<string | null>(null);

  const handleAddComment = (pinId: string, content: string) => {
    onAddComment?.(pinId, content);
  };

  const handleResolve = (pinId: string) => {
    onResolvePin?.(pinId);
    setSelectedPinId(null);
  };

  const handleDelete = (pinId: string) => {
    onDeletePin?.(pinId);
    setSelectedPinId(null);
  };

  return (
    <div className={cn("absolute inset-0 pointer-events-none", className)}>
      {pins.map((pin) => (
        <div
          key={pin.id}
          className="absolute pointer-events-auto"
          style={{ left: 0, top: 0 }}
        >
          <PinMarker
            pin={pin}
            isSelected={selectedPinId === pin.id}
            onClick={() =>
              setSelectedPinId(selectedPinId === pin.id ? null : pin.id)
            }
          />
          <CommentThread
            pin={pin}
            isOpen={selectedPinId === pin.id}
            onClose={() => setSelectedPinId(null)}
            onAddComment={(content) => handleAddComment(pin.id, content)}
            onResolve={() => handleResolve(pin.id)}
            onDelete={() => handleDelete(pin.id)}
          />
        </div>
      ))}
    </div>
  );
}

export default CommentPins;
