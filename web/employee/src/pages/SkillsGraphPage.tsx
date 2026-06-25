import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { skillsGraphApi, type SkillGraphNode } from '@/api/skills_graph';
import { GitBranch, TrendingUp, AlertTriangle, Search } from 'lucide-react';
import {
  ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, ZAxis,
} from 'recharts';

const categoryColors: Record<string, string> = {
  technical: 'bg-blue-500/20 text-blue-400',
  soft: 'bg-purple-500/20 text-purple-400',
  domain: 'bg-green-500/20 text-green-400',
  tool: 'bg-yellow-500/20 text-yellow-400',
};

const categories = ['', 'technical', 'soft', 'domain', 'tool'];

export function SkillsGraphPage() {
  const [categoryFilter, setCategoryFilter] = useState('');
  const [search, setSearch] = useState('');

  const { data: skillsData, isLoading } = useQuery({
    queryKey: ['skills-graph', categoryFilter],
    queryFn: () => skillsGraphApi.get(categoryFilter ? { category: categoryFilter } : undefined),
  });

  const { data: gapData } = useQuery({
    queryKey: ['skills-graph', 'gaps'],
    queryFn: () => skillsGraphApi.getGap(),
  });

  const skills = skillsData?.data?.skills || [];
  const gaps = gapData?.data?.gaps || [];

  const filteredSkills = skills.filter((s) =>
    !search || s.skill_name.toLowerCase().includes(search.toLowerCase())
  );

  const sortedSkills = [...filteredSkills].sort((a, b) => b.gap_score - a.gap_score);

  const scatterData = skills.map((s) => ({
    name: s.skill_name,
    supply: s.supply_score,
    demand: s.demand_score,
    gap: s.gap_score,
    employees: s.total_employees,
  }));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <GitBranch className="h-6 w-6 text-green-400" />
          <h1 className="text-2xl font-bold">Skills Graph</h1>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <div className="relative flex-1 sm:max-w-xs">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-500" />
          <input
            type="text"
            placeholder="Search skills..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-gray-700 bg-gray-800 py-2 pl-9 pr-3 text-sm text-gray-100 placeholder-gray-500"
          />
        </div>
        <div className="flex gap-2">
          {categories.map((cat) => (
            <button
              key={cat}
              onClick={() => setCategoryFilter(cat)}
              className={`rounded-lg px-3 py-2 text-xs font-medium transition-colors ${
                categoryFilter === cat
                  ? 'bg-blue-600 text-white'
                  : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {cat || 'All'}
            </button>
          ))}
        </div>
      </div>

      {gaps.length > 0 && (
        <div className="space-y-2">
          <h3 className="flex items-center gap-2 text-sm font-medium text-gray-300">
            <AlertTriangle className="h-4 w-4 text-yellow-400" />
            Top Skill Gaps
          </h3>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
            {gaps.slice(0, 6).map((gap) => (
              <div key={gap.id} className="rounded-xl border border-yellow-500/30 bg-yellow-500/5 p-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm font-medium text-gray-100">{gap.skill_name}</span>
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${categoryColors[gap.category] || ''}`}>
                    {gap.category}
                  </span>
                </div>
                <div className="mt-2 flex items-center gap-4 text-xs text-gray-400">
                  <span>Supply: {gap.supply_score}</span>
                  <span>Demand: {gap.demand_score}</span>
                  <span className="font-medium text-yellow-400">Gap: {gap.gap_score.toFixed(1)}</span>
                </div>
                <div className="mt-2 h-1.5 rounded-full bg-gray-800">
                  <div
                    className="h-full rounded-full bg-yellow-500"
                    style={{ width: `${Math.min(gap.gap_score * 10, 100)}%` }}
                  />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-12">
          <div className="h-8 w-8 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <h3 className="mb-4 text-sm font-medium text-gray-300">Supply vs Demand</h3>
              <div className="h-72">
                <ResponsiveContainer width="100%" height="100%">
                  <ScatterChart>
                    <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
                    <XAxis type="number" dataKey="supply" name="Supply" stroke="#9ca3af" fontSize={12} />
                    <YAxis type="number" dataKey="demand" name="Demand" stroke="#9ca3af" fontSize={12} />
                    <ZAxis type="number" dataKey="gap" range={[40, 200]} name="Gap" />
                    <Tooltip
                      cursor={{ strokeDasharray: '3 3' }}
                      contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '8px' }}
                      labelStyle={{ color: '#f3f4f6' }}
                      formatter={(value, name) => [value, name]}
                    />
                    <Scatter data={scatterData} fill="#3b82f6" />
                  </ScatterChart>
                </ResponsiveContainer>
              </div>
            </div>

            <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
              <h3 className="mb-4 text-sm font-medium text-gray-300">Trending Skills</h3>
              <div className="space-y-3">
                {skills.filter((s) => s.trending).slice(0, 8).map((skill) => (
                  <div key={skill.id} className="flex items-center justify-between rounded-lg bg-gray-800/50 px-4 py-3">
                    <div className="flex items-center gap-3">
                      <TrendingUp className="h-4 w-4 text-green-400" />
                      <span className="text-sm font-medium text-gray-100">{skill.skill_name}</span>
                      <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${categoryColors[skill.category] || ''}`}>
                        {skill.category}
                      </span>
                    </div>
                    <span className="text-sm text-gray-400">{skill.total_employees} employees</span>
                  </div>
                ))}
                {skills.filter((s) => s.trending).length === 0 && (
                  <p className="py-4 text-center text-sm text-gray-500">No trending skills</p>
                )}
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-gray-800 bg-gray-900 p-4">
            <h3 className="mb-4 text-sm font-medium text-gray-300">All Skills ({sortedSkills.length})</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-800 text-left text-xs text-gray-500">
                    <th className="pb-3 pr-4">Skill</th>
                    <th className="pb-3 pr-4">Category</th>
                    <th className="pb-3 pr-4">Employees</th>
                    <th className="pb-3 pr-4">Proficiency</th>
                    <th className="pb-3 pr-4">Supply</th>
                    <th className="pb-3 pr-4">Demand</th>
                    <th className="pb-3">Gap</th>
                  </tr>
                </thead>
                <tbody>
                  {sortedSkills.map((skill) => (
                    <tr key={skill.id} className="border-b border-gray-800/50">
                      <td className="py-3 pr-4">
                        <div className="flex items-center gap-2">
                          <span className="font-medium text-gray-100">{skill.skill_name}</span>
                          {skill.trending && <TrendingUp className="h-3 w-3 text-green-400" />}
                        </div>
                      </td>
                      <td className="py-3 pr-4">
                        <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${categoryColors[skill.category] || 'bg-gray-500/20 text-gray-400'}`}>
                          {skill.category}
                        </span>
                      </td>
                      <td className="py-3 pr-4 text-gray-300">{skill.total_employees}</td>
                      <td className="py-3 pr-4 text-gray-300">{(skill.avg_proficiency * 100).toFixed(0)}%</td>
                      <td className="py-3 pr-4 text-gray-300">{skill.supply_score}</td>
                      <td className="py-3 pr-4 text-gray-300">{skill.demand_score}</td>
                      <td className="py-3">
                        <span className={skill.gap_score > 30 ? 'font-medium text-yellow-400' : 'text-gray-300'}>
                          {skill.gap_score.toFixed(1)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
