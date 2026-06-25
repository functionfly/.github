export interface GaugeData {
  value: number | string;
  label: string;
}

export interface GaugeProps {
  data: GaugeData;
  isFirst?: boolean;
}

export function Gauge({ data, isFirst = false }: GaugeProps) {
  return (
    <div className={`gauge ${isFirst ? 'gauge--first' : ''}`}>
      <div className="gauge__value">
        <span className="gauge__dot" aria-hidden="true" />
        <span className="gauge__number">{data.value}</span>
      </div>
      <div className="gauge__label">{data.label}</div>
    </div>
  );
}

export interface GaugeStripProps {
  children: React.ReactNode;
}

export function GaugeStrip({ children }: GaugeStripProps) {
  return <div className="gauge-strip">{children}</div>;
}
