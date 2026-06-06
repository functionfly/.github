import { useEffect } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  ArrowUpRight,
  CheckCircle2,
  CircleDot,
  Clock,
  Code2,
  FileText,
  ListChecks,
  ShieldCheck,
  Target,
  Timer,
  ToggleLeft,
  Zap,
} from 'lucide-react';
import { PageLayout } from '@/components/layout/PageLayout';
import { PageHeader } from '@/components/layout/PageHeader';
import { SectionHeader } from '@/components/ui/section-header';
import { Button } from '@/components/ui/button';
import { trackEvent, trackPageView } from '@/lib/analytics';

type QuestionFormat = {
  id: string;
  name: string;
  blurb: string;
  Icon: React.ComponentType<{ className?: string }>;
  example: { stem: string; options: string[] };
  tip: string;
};

const QUESTION_FORMATS: QuestionFormat[] = [
  {
    id: 'multiple_choice',
    name: 'Single answer',
    blurb: 'One correct option out of four. Most common format.',
    Icon: CircleDot,
    example: {
      stem: 'Which FunctionFly runtime is responsible for executing user code?',
      options: ['SAR', 'Composer', 'Vault', 'DNS Resolver'],
    },
    tip: 'Look for absolute language like "always" / "never" — it is often the distractor.',
  },
  {
    id: 'multi_select',
    name: 'Multi-select',
    blurb: 'Choose all options that apply. Points awarded per correct pick.',
    Icon: ListChecks,
    example: {
      stem: 'Which of the following are valid credential states? (Select all that apply)',
      options: ['Active', 'Expiring', 'Revoked', 'Encrypted'],
    },
    tip: 'Partial credit applies — but a wrong pick can cancel a correct one. Read carefully.',
  },
  {
    id: 'true_false',
    name: 'True / False',
    blurb: 'Short factual statements, no options to choose from.',
    Icon: ToggleLeft,
    example: {
      stem: 'Credentials issued by FunctionFly are cryptographically signed and publicly verifiable.',
      options: ['True', 'False'],
    },
    tip: 'If a statement is mostly true but has one false detail, the answer is False.',
  },
];

const TIMELINE = [
  {
    step: '01',
    title: 'Start the exam',
    detail: 'Begin when ready. Timer starts immediately — pausing is not allowed.',
    Icon: Timer,
  },
  {
    step: '02',
    title: 'Work through questions',
    detail: 'Move freely between questions. Answers autosave to the server.',
    Icon: ListChecks,
  },
  {
    step: '03',
    title: 'Complete practical challenges',
    detail: 'Live code or configuration tasks in a sandboxed environment.',
    Icon: Code2,
  },
  {
    step: '04',
    title: 'Submit & get graded',
    detail: 'Review your answers, then submit. Results are usually instant.',
    Icon: CheckCircle2,
  },
];

const RULES = [
  {
    title: 'Time limit is firm',
    detail: 'Exams auto-submit when the clock runs out. Unanswered questions count as zero.',
    Icon: Clock,
  },
  {
    title: 'One attempt per purchase',
    detail: 'Abandoning an exam forfeits the attempt. Failed exams can be retaken after a cooldown.',
    Icon: Target,
  },
  {
    title: 'Passing threshold',
    detail: 'You need the tier pass threshold (typically 70%) across the combined score.',
    Icon: ShieldCheck,
  },
  {
    title: 'No external aids',
    detail: 'Reference material is not provided. Practical challenges are open-book within the sandbox.',
    Icon: FileText,
  },
];

export function ExamFormatPage() {
  const navigate = useNavigate();

  useEffect(() => {
    trackPageView('/certification/exams');
  }, []);

  return (
    <PageLayout>
      <PageHeader
        title="Exam format"
        subtitle="Everything you need to know before you start a certification exam."
      />

      {/* ── Section 01 · At a glance ───────────────────────────────────── */}
      <section aria-labelledby="glance-heading" className="mb-12 space-y-5">
        <SectionHeader
          eyebrow="01 · At a glance"
          title="What to expect"
          description="Each tier follows the same structure. Differences are in length, difficulty, and the mix of practical challenges."
          id="glance-heading"
        />
        <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <GlanceCard
            label="Time"
            value="30–90 min"
            sub="Depends on tier"
            Icon={Timer}
          />
          <GlanceCard
            label="Questions"
            value="20–60"
            sub="Mixed formats"
            Icon={ListChecks}
          />
          <GlanceCard
            label="Practical"
            value="0–5"
            sub="Sandboxed challenges"
            Icon={Code2}
          />
        </div>
      </section>

      {/* ── Section 02 · Question formats ──────────────────────────────── */}
      <section aria-labelledby="formats-heading" className="mb-12 space-y-5">
        <SectionHeader
          eyebrow="02 · Question formats"
          title="How questions are asked"
          description="All questions are auto-graded. There is no partial manual review."
          id="formats-heading"
        />
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
          {QUESTION_FORMATS.map((q, i) => (
            <FormatCard key={q.id} format={q} index={i} />
          ))}
        </div>
      </section>

      {/* ── Section 03 · The flow ───────────────────────────────────────── */}
      <section aria-labelledby="flow-heading" className="mb-12 space-y-5">
        <SectionHeader
          eyebrow="03 · The flow"
          title="From start to results"
          description="A linear progression. You can revisit earlier questions until you submit."
          id="flow-heading"
        />
        <ol className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {TIMELINE.map((s, i) => (
            <motion.li
              key={s.step}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.25, delay: i * 0.05 }}
              className="relative rounded-lg border border-white/[0.06] bg-bg-secondary/40 p-4"
            >
              <div className="mb-3 flex items-center justify-between">
                <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-text-muted">
                  {s.step}
                </p>
                <s.Icon className="h-3.5 w-3.5 text-brand-400" aria-hidden />
              </div>
              <p className="text-sm font-medium text-text-primary">{s.title}</p>
              <p className="mt-1 text-xs leading-relaxed text-text-secondary">
                {s.detail}
              </p>
            </motion.li>
          ))}
        </ol>
      </section>

      {/* ── Section 04 · Rules ─────────────────────────────────────────── */}
      <section aria-labelledby="rules-heading" className="mb-12 space-y-5">
        <SectionHeader
          eyebrow="04 · Rules"
          title="What you need to know"
          description="No surprises on exam day. Read these once and you're set."
          id="rules-heading"
        />
        <ul className="grid grid-cols-1 gap-3 md:grid-cols-2">
          {RULES.map((r) => (
            <li
              key={r.title}
              className="flex items-start gap-3 rounded-lg border border-white/[0.06] bg-bg-secondary/40 p-4"
            >
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-white/[0.08] bg-white/[0.03]">
                <r.Icon className="h-3.5 w-3.5 text-text-muted" aria-hidden />
              </div>
              <div>
                <p className="text-sm font-medium text-text-primary">{r.title}</p>
                <p className="mt-1 text-xs leading-relaxed text-text-secondary">
                  {r.detail}
                </p>
              </div>
            </li>
          ))}
        </ul>
      </section>

      {/* ── Section 05 · Scoring ───────────────────────────────────────── */}
      <section aria-labelledby="scoring-heading" className="mb-12 space-y-5">
        <SectionHeader
          eyebrow="05 · Scoring"
          title="How you're graded"
          description="A single combined score drives the pass/fail decision."
          id="scoring-heading"
        />
        <div className="grid grid-cols-1 gap-3 md:grid-cols-2">
          <ScoreCard
            label="Knowledge"
            weight="60–80%"
            detail="Multiple-choice, multi-select, and true/false questions across the domain."
            Icon={ListChecks}
          />
          <ScoreCard
            label="Practical"
            weight="20–40%"
            detail="Sandboxed challenges scored by an automated rubric against expected outputs."
            Icon={Code2}
          />
        </div>
        <div className="flex items-start gap-3 rounded-lg border border-brand-500/20 bg-brand-500/[0.04] p-4">
          <Zap className="mt-0.5 h-4 w-4 shrink-0 text-brand-400" aria-hidden />
          <p className="text-sm text-text-secondary">
            Pass thresholds are per tier and shown on the certification catalog.
            Hitting the threshold issues a signed credential and a public
            verification URL.
          </p>
        </div>
      </section>

      {/* ── CTA ────────────────────────────────────────────────────────── */}
      <section className="rounded-xl border border-dashed border-white/[0.08] bg-white/[0.012] p-6 sm:p-8">
        <div className="flex flex-col items-start gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-text-muted">
              Ready to begin?
            </p>
            <h3 className="mt-1 font-display text-lg font-medium tracking-tight text-text-primary">
              Pick a tier and start your first attempt.
            </h3>
            <p className="mt-1 text-sm text-text-secondary">
              You can abandon before the timer starts without penalty.
            </p>
          </div>
          <div className="flex flex-wrap gap-2">
            <Button
              onClick={() => {
                trackEvent('exam_format_browse_tiers');
                navigate('/certification');
              }}
            >
              Browse tiers
              <ArrowUpRight className="h-3.5 w-3.5" />
            </Button>
            <Button variant="ghost" asChild>
              <Link to="/credentials">
                <FileText className="h-3.5 w-3.5" />
                Back to credentials
              </Link>
            </Button>
          </div>
        </div>
      </section>
    </PageLayout>
  );
}

// ── Sub-components ────────────────────────────────────────────────────────────

function GlanceCard({
  label,
  value,
  sub,
  Icon,
}: {
  label: string;
  value: string;
  sub: string;
  Icon: React.ComponentType<{ className?: string }>;
}) {
  return (
    <div className="rounded-lg border border-white/[0.06] bg-bg-secondary/40 p-4">
      <div className="mb-3 flex items-center justify-between">
        <p className="font-mono text-[10px] uppercase tracking-[0.18em] text-text-muted">
          {label}
        </p>
        <Icon className="h-3.5 w-3.5 text-text-muted" aria-hidden />
      </div>
      <p className="font-display text-2xl font-medium tracking-tight text-text-primary">
        {value}
      </p>
      <p className="mt-1 text-xs text-text-secondary">{sub}</p>
    </div>
  );
}

function FormatCard({ format, index }: { format: QuestionFormat; index: number }) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 8 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.25, delay: index * 0.05 }}
      className="flex flex-col rounded-lg border border-white/[0.06] bg-bg-secondary/40"
    >
      <div className="flex items-center gap-3 border-b border-white/[0.06] p-4">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-white/[0.08] bg-white/[0.03]">
          <format.Icon className="h-3.5 w-3.5 text-brand-400" aria-hidden />
        </div>
        <div className="min-w-0">
          <p className="text-sm font-medium text-text-primary">{format.name}</p>
          <p className="text-xs text-text-secondary">{format.blurb}</p>
        </div>
      </div>
      <div className="flex-1 space-y-3 p-4">
        <p className="text-sm leading-relaxed text-text-primary">
          {format.example.stem}
        </p>
        <ul className="space-y-1.5" role="list">
          {format.example.options.map((opt, i) => (
            <li
              key={i}
              className="flex items-center gap-2 rounded-md border border-white/[0.06] bg-white/[0.012] px-3 py-2 text-xs text-text-secondary"
            >
              <span className="font-mono text-[10px] tabular-nums text-text-muted">
                {String.fromCharCode(65 + i)}
              </span>
              <span>{opt}</span>
            </li>
          ))}
        </ul>
        <p className="border-t border-white/[0.06] pt-3 text-[11px] leading-relaxed text-text-muted">
          <span className="font-mono uppercase tracking-wider text-text-secondary">
            Tip ·{' '}
          </span>
          {format.tip}
        </p>
      </div>
    </motion.div>
  );
}

function ScoreCard({
  label,
  weight,
  detail,
  Icon,
}: {
  label: string;
  weight: string;
  detail: string;
  Icon: React.ComponentType<{ className?: string }>;
}) {
  return (
    <div className="flex items-start gap-3 rounded-lg border border-white/[0.06] bg-bg-secondary/40 p-4">
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md border border-white/[0.08] bg-white/[0.03]">
        <Icon className="h-4 w-4 text-text-muted" aria-hidden />
      </div>
      <div className="min-w-0 flex-1">
        <div className="flex items-center justify-between gap-2">
          <p className="text-sm font-medium text-text-primary">{label}</p>
          <p className="font-mono text-[11px] uppercase tracking-wider text-brand-400">
            {weight}
          </p>
        </div>
        <p className="mt-1 text-xs leading-relaxed text-text-secondary">
          {detail}
        </p>
      </div>
    </div>
  );
}
