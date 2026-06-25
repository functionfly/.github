import { Radar, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, ResponsiveContainer } from 'recharts';

interface ReputationRadarProps {
  scores: Record<string, number>;
  className?: string;
}

const defaultCategories = [
  'Engineering',
  'Innovation',
  'Leadership',
  'Mentorship',
  'Reliability',
  'Collaboration',
];

export function ReputationRadar({ scores, className = '' }: ReputationRadarProps) {
  const data = defaultCategories.map((cat) => ({
    category: cat,
    score: scores[cat] ?? 0,
    fullMark: 100,
  }));

  return (
    <div className={`rounded-xl border border-gray-800 bg-gray-900 p-5 ${className}`}>
      <h3 className="mb-4 text-sm font-medium text-gray-300">Reputation Profile</h3>
      <div className="h-64">
        <ResponsiveContainer width="100%" height="100%">
          <RadarChart data={data} cx="50%" cy="50%" outerRadius="70%">
            <PolarGrid stroke="#374151" />
            <PolarAngleAxis dataKey="category" tick={{ fill: '#9CA3AF', fontSize: 11 }} />
            <PolarRadiusAxis angle={30} domain={[0, 100]} tick={false} axisLine={false} />
            <Radar
              name="Score"
              dataKey="score"
              stroke="#3B82F6"
              fill="#3B82F6"
              fillOpacity={0.25}
              strokeWidth={2}
            />
          </RadarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
