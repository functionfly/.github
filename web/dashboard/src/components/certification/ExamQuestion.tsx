import { cn } from '@/lib/utils';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import type { CertQuestionPublic } from '@/api/certification';

interface ExamQuestionProps {
  question: CertQuestionPublic;
  selectedAnswer?: string;
  onAnswer: (answer: string) => void;
  questionNumber: number;
  totalQuestions: number;
}

const difficultyColors: Record<string, string> = {
  easy: 'bg-emerald-500/10 text-emerald-500 border-emerald-500/20',
  medium: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
  hard: 'bg-red-500/10 text-red-500 border-red-500/20',
};

export function ExamQuestion({
  question,
  selectedAnswer,
  onAnswer,
  questionNumber,
  totalQuestions,
}: ExamQuestionProps) {
  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-text-muted">
            Question {questionNumber} of {totalQuestions}
          </span>
          <Badge variant="outline" className="text-xs">
            {question.category}
          </Badge>
          <span className={cn(
            'inline-flex items-center rounded-full border px-2 py-0.5 text-xs font-medium',
            difficultyColors[question.difficulty] || difficultyColors.medium
          )}>
            {question.difficulty}
          </span>
        </div>
        <span className="text-sm text-text-muted">{question.points} pt{question.points !== 1 ? 's' : ''}</span>
      </div>

      {/* Question text */}
      <h3 className="text-lg font-medium text-text-primary">{question.question_text}</h3>

      {/* Options */}
      <RadioGroup value={selectedAnswer} onValueChange={onAnswer} className="space-y-3">
        {question.options.map((option) => (
          <div
            key={option.id}
            className={cn(
              'flex items-center space-x-3 rounded-lg border p-4 transition-all cursor-pointer',
              selectedAnswer === option.id
                ? 'border-brand-500 bg-brand-500/5'
                : 'border-theme bg-card hover:border-brand-500/50 hover:bg-brand-500/5'
            )}
            onClick={() => onAnswer(option.id)}
          >
            <RadioGroupItem value={option.id} id={`${question.id}-${option.id}`} />
            <Label
              htmlFor={`${question.id}-${option.id}`}
              className="flex-1 cursor-pointer text-sm text-text-primary"
            >
              {option.text}
            </Label>
          </div>
        ))}
      </RadioGroup>
    </div>
  );
}
