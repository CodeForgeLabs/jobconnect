interface ProgressRingProps {
  value: number;
  size?: number;
  stroke?: number;
  label?: string;
}

export default function ProgressRing({ value, size = 180, stroke = 12, label }: ProgressRingProps) {
  const bounded = Math.max(0, Math.min(100, value));
  const radius = (size - stroke) / 2;
  const circumference = 2 * Math.PI * radius;
  const dashOffset = circumference - (bounded / 100) * circumference;

  return (
    <div className="relative inline-flex items-center justify-center" style={{ width: size, height: size }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="rgba(34,69,53,0.13)"
          strokeWidth={stroke}
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="var(--jc-accent)"
          strokeWidth={stroke}
          strokeLinecap="round"
          strokeDasharray={circumference}
          strokeDashoffset={dashOffset}
          transform={`rotate(-90 ${size / 2} ${size / 2})`}
        />
      </svg>
      <div className="absolute text-center">
        <p className="text-3xl font-semibold text-[var(--jc-ink)]">{bounded}%</p>
        {label ? <p className="text-xs text-[var(--jc-ink-muted)]">{label}</p> : null}
      </div>
    </div>
  );
}

