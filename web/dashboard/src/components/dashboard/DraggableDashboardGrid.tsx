import { useState, useCallback } from 'react';
import {
  DndContext,
  closestCenter,
  KeyboardSensor,
  PointerSensor,
  TouchSensor,
  useSensor,
  useSensors,
  DragOverlay,
  defaultDropAnimationSideEffects,
  type DragStartEvent,
  type DragEndEvent,
  type DropAnimation,
} from '@dnd-kit/core';
import {
  arrayMove,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { GripVertical, RefreshCcw } from 'lucide-react';
import { motion } from 'framer-motion';

export interface DraggableSection {
  id: string;
  content: React.ReactNode;
}

interface SortableItemProps {
  id: string;
  children: React.ReactNode;
  isDragging?: boolean;
}

function SortableItem({ id, children, isDragging }: SortableItemProps) {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isCurrentlyDragging,
  } = useSortable({ id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    zIndex: isCurrentlyDragging ? 50 : undefined,
  };

  return (
    <motion.div
      ref={setNodeRef}
      style={style}
      className={`relative ${isCurrentlyDragging ? 'opacity-50' : ''}`}
      layout
    >
      <div className="group relative">
        {/* Drag Handle */}
        <div
          className={`absolute left-0 top-1/2 -translate-y-1/2 -translate-x-3 z-10 cursor-grab active:cursor-grabbing
            opacity-0 group-hover:opacity-100 transition-opacity duration-200
            p-1.5 rounded-md bg-bg-secondary/80 border border-border-subtle
            shadow-sm hover:bg-bg-tertiary hover:border-border-hover
            ${isCurrentlyDragging ? 'opacity-100 cursor-grabbing' : ''}`}
          {...attributes}
          {...listeners}
        >
          <GripVertical className="w-4 h-4 text-text-muted" />
        </div>
        <div className="pl-2">
          {children}
        </div>
      </div>
    </motion.div>
  );
}

export interface DraggableDashboardGridProps {
  sections: DraggableSection[];
  storageKey: string;
  onOrderChange?: (newOrder: string[]) => void;
}

const defaultDropAnimation: DropAnimation = {
  sideEffects: defaultDropAnimationSideEffects({
    styles: {
      active: {
        opacity: '0.5',
      },
    },
  }),
};

export function DraggableDashboardGrid({
  sections: initialSections,
  storageKey,
  onOrderChange,
}: DraggableDashboardGridProps) {
  // Compute valid sections (filter out null content) - must be done before any hooks
  const validSections = initialSections.filter(s => s.content !== null && s.content !== undefined);
  const validSectionIds = new Set(validSections.map(s => s.id));
  
  // Load saved order from localStorage or use default order
  const [sectionOrder, setSectionOrder] = useState<string[]>(() => {
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem(storageKey);
      if (saved) {
        try {
          const parsed = JSON.parse(saved) as string[];
          // Compute valid sections again inside the lazy initializer
          const validSecs = initialSections.filter(s => s.content !== null && s.content !== undefined);
          const validIds = new Set(validSecs.map(s => s.id));
          // Validate that all saved IDs are valid section IDs (ignoring any that may have been removed)
          const filtered = parsed.filter(id => validIds.has(id));
          // Add any new sections not in the saved order
          const missing = validSecs.filter(s => !parsed.includes(s.id)).map(s => s.id);
          if (filtered.length > 0) return [...filtered, ...missing];
        } catch {
          // Invalid saved data, use default
        }
      }
    }
    return validSections.map(s => s.id);
  });

  const [activeId, setActiveId] = useState<string | null>(null);
  
  const orderedSections = sectionOrder
    .filter(id => validSectionIds.has(id))
    .map(id => validSections.find(s => s.id === id))
    .filter((s): s is DraggableSection => s !== undefined);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    }),
    useSensor(TouchSensor, {
      activationConstraint: {
        delay: 250,
        tolerance: 5,
      },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const handleDragStart = useCallback((event: DragStartEvent) => {
    setActiveId(event.active.id as string);
  }, []);

  const handleDragEnd = useCallback((event: DragEndEvent) => {
    const { active, over } = event;

    if (over && active.id !== over.id) {
      setSectionOrder((items) => {
        const oldIndex = items.indexOf(active.id as string);
        const newIndex = items.indexOf(over.id as string);
        const newOrder = arrayMove(items, oldIndex, newIndex);
        
        // Save to localStorage
        localStorage.setItem(storageKey, JSON.stringify(newOrder));
        
        // Notify parent
        onOrderChange?.(newOrder);
        
        return newOrder;
      });
    }

    setActiveId(null);
  }, [storageKey, onOrderChange]);

  const handleResetLayout = useCallback(() => {
    const defaultOrder = initialSections.map(s => s.id);
    setSectionOrder(defaultOrder);
    localStorage.setItem(storageKey, JSON.stringify(defaultOrder));
    onOrderChange?.(defaultOrder);
  }, [initialSections, storageKey, onOrderChange]);

  const activeSection = activeId ? initialSections.find(s => s.id === activeId) : null;

  return (
    <div className="relative">
      {/* Reset Layout Button */}
      <div className="fixed bottom-6 right-6 z-40">
        <motion.button
          onClick={handleResetLayout}
          className="flex items-center gap-2 px-4 py-2.5 bg-bg-secondary/90 backdrop-blur-sm 
            border border-border-subtle rounded-full shadow-lg
            text-sm font-medium text-text-secondary hover:text-text-primary
            hover:bg-bg-tertiary hover:border-border-hover transition-all duration-200
            group"
          whileHover={{ scale: 1.02 }}
          whileTap={{ scale: 0.98 }}
          title="Reset layout to default order"
        >
          <RefreshCcw className="w-4 h-4 group-hover:-rotate-180 transition-transform duration-500" />
          <span>Reset Layout</span>
        </motion.button>
      </div>

      {/* Drag Instructions */}
      <div className="mb-4 text-center sm:text-left">
        <p className="text-sm text-text-muted flex items-center gap-2 justify-center sm:justify-start">
          <GripVertical className="w-4 h-4" />
          <span>Drag the handle on the left to rearrange sections</span>
        </p>
      </div>

      <DndContext
        sensors={sensors}
        collisionDetection={closestCenter}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <SortableContext items={sectionOrder} strategy={verticalListSortingStrategy}>
          <div className="space-y-6">
            {orderedSections.map((section) => (
              <SortableItem key={section.id} id={section.id}>
                {section.content}
              </SortableItem>
            ))}
          </div>
        </SortableContext>

        <DragOverlay dropAnimation={defaultDropAnimation}>
          {activeSection ? (
            <div className="opacity-80 scale-[1.02] shadow-2xl">
              <div className="absolute -left-3 top-1/2 -translate-y-1/2 p-1.5 rounded-md bg-brand-500 text-white">
                <GripVertical className="w-4 h-4" />
              </div>
              {activeSection.content}
            </div>
          ) : null}
        </DragOverlay>
      </DndContext>
    </div>
  );
}

export default DraggableDashboardGrid;
