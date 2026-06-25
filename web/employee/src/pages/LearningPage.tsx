import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { learningApi, type LearningCourse, type EmployeeLearning } from '@/api/learning';
import { GraduationCap, BookOpen, Clock, Award, Play, CheckCircle } from 'lucide-react';

const difficultyColors: Record<string, string> = {
  beginner: 'bg-green-500/20 text-green-400',
  intermediate: 'bg-blue-500/20 text-blue-400',
  advanced: 'bg-orange-500/20 text-orange-400',
  expert: 'bg-red-500/20 text-red-400',
};

export function LearningPage() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'catalog' | 'progress'>('catalog');
  const { data: coursesData } = useQuery({
    queryKey: ['learning', 'courses'],
    queryFn: () => learningApi.listCourses(),
  });
  const { data: progressData } = useQuery({
    queryKey: ['learning', 'progress'],
    queryFn: () => learningApi.getMyProgress(),
  });

  const enrollMutation = useMutation({
    mutationFn: (courseId: string) => learningApi.enroll(courseId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['learning'] });
    },
  });

  const courses = coursesData?.data?.courses || [];
  const progress = progressData?.data?.progress || [];

  const enrolledCourseIds = new Set(progress.map((p) => p.course_id));

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Learning Center</h1>

      <div className="flex gap-1 rounded-lg bg-gray-900 p-1">
        <button
          onClick={() => setTab('catalog')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${
            tab === 'catalog' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'
          }`}
        >
          Course Catalog
        </button>
        <button
          onClick={() => setTab('progress')}
          className={`flex-1 rounded-md px-4 py-2 text-sm font-medium ${
            tab === 'progress' ? 'bg-blue-600 text-white' : 'text-gray-400 hover:text-gray-200'
          }`}
        >
          My Progress
        </button>
      </div>

      {tab === 'catalog' ? (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {courses.map((course) => (
            <div key={course.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="mb-3 flex items-start justify-between">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-blue-600/20">
                  <BookOpen className="h-5 w-5 text-blue-400" />
                </div>
                {course.difficulty && (
                  <span className={`rounded-full px-2 py-0.5 text-xs ${difficultyColors[course.difficulty] || ''}`}>
                    {course.difficulty}
                  </span>
                )}
              </div>
              <h3 className="mb-1 font-semibold text-gray-100">{course.title}</h3>
              {course.description && (
                <p className="mb-3 line-clamp-2 text-sm text-gray-400">{course.description}</p>
              )}
              <div className="mb-3 flex items-center gap-3 text-xs text-gray-500">
                {course.duration_min && (
                  <span className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    {course.duration_min} min
                  </span>
                )}
                {course.is_mandatory && (
                  <span className="rounded bg-red-500/20 px-1.5 py-0.5 text-red-400">Required</span>
                )}
              </div>
              {enrolledCourseIds.has(course.id) ? (
                <span className="flex w-full items-center justify-center gap-2 rounded-lg border border-green-600/30 bg-green-600/10 px-4 py-2 text-sm text-green-400">
                  <CheckCircle className="h-4 w-4" />
                  Enrolled
                </span>
              ) : (
                <button
                  onClick={() => enrollMutation.mutate(course.id)}
                  className="flex w-full items-center justify-center gap-2 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
                >
                  <Play className="h-4 w-4" />
                  Enroll
                </button>
              )}
            </div>
          ))}
          {courses.length === 0 && (
            <div className="col-span-full flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <GraduationCap className="mb-4 h-12 w-12 text-gray-600" />
              <p className="text-gray-400">No courses available</p>
            </div>
          )}
        </div>
      ) : (
        <div className="space-y-3">
          {progress.map((p) => (
            <div key={p.id} className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="font-medium text-gray-200">{p.course?.title || p.course_id}</p>
                  <p className="text-sm text-gray-500 capitalize">{p.status.replace('_', ' ')}</p>
                </div>
                <div className="text-right">
                  <p className="text-2xl font-bold text-blue-400">{p.progress_pct}%</p>
                </div>
              </div>
              <div className="mt-3 h-2 rounded-full bg-gray-800">
                <div
                  className="h-full rounded-full bg-blue-600 transition-all"
                  style={{ width: `${p.progress_pct}%` }}
                />
              </div>
              {p.completed_at && (
                <p className="mt-2 flex items-center gap-1 text-xs text-green-400">
                  <Award className="h-3 w-3" />
                  Completed
                </p>
              )}
            </div>
          ))}
          {progress.length === 0 && (
            <div className="flex flex-col items-center justify-center rounded-xl border border-gray-800 bg-gray-900 py-12">
              <p className="text-gray-400">No learning progress yet</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
