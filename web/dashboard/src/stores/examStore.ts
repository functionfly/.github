import { create } from 'zustand';

interface ExamState {
  // Current exam
  examId: string | null;
  answers: Record<string, string>;
  currentQuestionIndex: number;

  // Actions
  setExamId: (id: string | null) => void;
  setAnswer: (questionId: string, answer: string) => void;
  setAnswers: (answers: Record<string, string>) => void;
  setCurrentQuestionIndex: (index: number) => void;
  nextQuestion: () => void;
  prevQuestion: () => void;
  reset: () => void;
}

export const useExamStore = create<ExamState>((set) => ({
  examId: null,
  answers: {},
  currentQuestionIndex: 0,

  setExamId: (id) => set({ examId: id }),

  setAnswer: (questionId, answer) =>
    set((state) => ({
      answers: { ...state.answers, [questionId]: answer },
    })),

  setAnswers: (answers) => set({ answers }),

  setCurrentQuestionIndex: (index) => set({ currentQuestionIndex: index }),

  nextQuestion: () =>
    set((state) => ({
      currentQuestionIndex: state.currentQuestionIndex + 1,
    })),

  prevQuestion: () =>
    set((state) => ({
      currentQuestionIndex: Math.max(0, state.currentQuestionIndex - 1),
    })),

  reset: () =>
    set({
      examId: null,
      answers: {},
      currentQuestionIndex: 0,
    }),
}));
