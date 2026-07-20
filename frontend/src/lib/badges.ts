export function categoryBadge(category: string): string {
  const map: Record<string, string> = {
    bug: "bg-red-500/10 text-red-500 border-red-500/20",
    feature: "bg-blue-500/10 text-blue-500 border-blue-500/20",
    praise: "bg-emerald-500/10 text-emerald-500 border-emerald-500/20",
    spam: "bg-amber-500/10 text-amber-500 border-amber-500/20",
    other: "bg-zinc-500/10 text-zinc-400 border-zinc-500/20",
  };
  return map[category] || map.other;
}

export function sentimentBadge(sentiment: string): string {
  const map: Record<string, string> = {
    positive: "bg-emerald-500/10 text-emerald-500 border-emerald-500/20",
    negative: "bg-red-500/10 text-red-500 border-red-500/20",
    neutral: "bg-zinc-500/10 text-zinc-400 border-zinc-500/20",
  };
  return map[sentiment] || map.neutral;
}
