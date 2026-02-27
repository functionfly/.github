import { useQuery } from '@tanstack/react-query'
import { useNavigate, useParams } from 'react-router-dom'
import { getFunctionVersions } from '../../api/functions'
import './VersionSelector.css'

export default function VersionSelector() {
  const { author, name } = useParams<{ author: string; name: string }>()
  const navigate = useNavigate()

  const { data: versions } = useQuery({
    queryKey: ['versions', author, name],
    queryFn: () => getFunctionVersions(author!, name!),
    enabled: !!author && !!name,
  })

  if (!versions || versions.length <= 1) {
    return null
  }

  const currentVersion = versions[0]?.version

  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    navigate(`/functions/${author}/${name}?version=${e.target.value}`)
  }

  return (
    <div className="version-selector">
      <label htmlFor="version">Version:</label>
      <select
        id="version"
        value={currentVersion}
        onChange={handleChange}
      >
        {versions.map((v) => (
          <option key={v.version} value={v.version}>
            {v.version}
          </option>
        ))}
      </select>
    </div>
  )
}
