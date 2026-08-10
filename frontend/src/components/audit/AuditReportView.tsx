import type { Audit, AuditInsights } from "@/types";
import { DistributionBars } from "@/components/RichResult";
import { categoryBadge, sentimentBadge } from "@/lib/badges";

type Bucket = { star?: number; count?: number; pct?: number };
type SentimentBucket = { count?: number; pct?: number };
type ReviewQuote = { text?: string; rating?: number; store?: string; sentiment?: string };
type CategoryInsight = {
  category?: string;
  label?: string;
  count?: number;
  pct?: number;
  reviews?: ReviewQuote[];
};
type FeedbackSuggestion = {
  title?: string;
  summary?: string;
  category?: string;
  priority?: string;
  supporting_reviews?: ReviewQuote[];
};
type FeaturedReview = {
  text?: string;
  rating?: number;
  store?: string;
  category?: string;
  sentiment?: string;
  highlight?: string;
};
type TimelineItem = { horizon?: string; tag?: string; title?: string; body?: string };
type Priority = { rank?: number; title?: string; body?: string };
type StoreInsightBlock = {
  store?: string;
  label?: string;
  statistics?: Record<string, unknown>;
  category_insights?: CategoryInsight[];
  feature_suggestions?: FeedbackSuggestion[];
  bug_suggestions?: FeedbackSuggestion[];
  featured_reviews?: FeaturedReview[];
};
type Finding = { title?: string; body?: string };

function num(v: unknown, digits = 2) {
  const n = Number(v);
  return Number.isFinite(n) ? n.toFixed(digits) : "0";
}

function list<T>(v: unknown): T[] {
  return Array.isArray(v) ? (v as T[]) : [];
}

function tagClass(tag?: string) {
  if (tag === "fast") return "bg-emerald-500/10 text-emerald-500 border-emerald-500/20";
  if (tag === "hard") return "bg-red-500/10 text-red-500 border-red-500/20";
  return "bg-amber-500/10 text-amber-500 border-amber-500/20";
}

function Panel({ children, className = "" }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`rounded-2xl border border-border bg-surface p-5 ${className}`}>{children}</div>
  );
}

function Stat({ label, value, hint, accent }: { label: string; value: string; hint?: string; accent?: boolean }) {
  return (
    <Panel className={accent ? "border-accent/30 bg-accent/5" : ""}>
      <p className="font-mono text-xs text-muted">{label}</p>
      <p className={`mt-2 font-mono text-4xl font-medium ${accent ? "text-accent" : "text-foreground"}`}>
        {value}
      </p>
      {hint && <p className="mt-2 text-sm text-muted">{hint}</p>}
    </Panel>
  );
}

function BarRow({ label, pct, count, tone }: { label: string; pct: number; count: number; tone: string }) {
  return (
    <div className="space-y-1 text-sm">
      <div className="flex items-center justify-between gap-2">
        <span className="text-foreground">{label}</span>
        <span className="font-mono text-xs text-muted">
          {count} · %{num(pct, 1)}
        </span>
      </div>
      <div className="h-2 overflow-hidden rounded-full bg-surface-2">
        <div className={`h-full rounded-full ${tone}`} style={{ width: `${Math.min(pct, 100)}%` }} />
      </div>
    </div>
  );
}

function ReviewQuotes({ quotes }: { quotes: ReviewQuote[] }) {
  if (!quotes.length) return null;
  return (
    <div className="mt-3 space-y-2">
      {quotes.map((q, i) => (
        <blockquote
          key={i}
          className="rounded-lg border border-border bg-background/60 px-3 py-2 text-sm text-muted"
        >
          <div className="mb-1 flex flex-wrap items-center gap-2">
            <span className="font-mono text-xs text-foreground">{q.rating ?? "?"}★</span>
            {q.store && <span className="font-mono text-[10px] uppercase text-muted">{q.store}</span>}
            {q.sentiment && (
              <span className={`rounded border px-1.5 py-0.5 text-[10px] ${sentimentBadge(String(q.sentiment))}`}>
                {q.sentiment}
              </span>
            )}
          </div>
          <p className="leading-relaxed text-foreground">&ldquo;{q.text}&rdquo;</p>
        </blockquote>
      ))}
    </div>
  );
}

function CategoryChart({ insights: cats, total }: { insights: CategoryInsight[]; total: number }) {
  const counts: Record<string, number> = {};
  const labels: Record<string, string> = {};
  for (const c of cats) {
    const key = c.category || "other";
    counts[key] = Number(c.count ?? 0);
    labels[key] = String(c.label || key);
  }
  if (Object.keys(counts).length === 0) return null;

  const entries = Object.entries(counts).sort((a, b) => b[1] - a[1]);
  const max = entries[0]?.[1] || 1;

  return (
    <div className="grid gap-6 lg:grid-cols-2">
      <Panel>
        <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Kategori dağılımı</h3>
        <DistributionBars counts={counts} total={total || 1} />
      </Panel>
      <Panel>
        <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Kategori yoğunluğu</h3>
        <div className="flex items-end gap-3" style={{ minHeight: 180 }}>
          {entries.map(([key, count]) => {
            const pct = max > 0 ? (count / max) * 100 : 0;
            const color =
              key === "bug"
                ? "bg-red-500"
                : key === "feature"
                  ? "bg-blue-500"
                  : key === "praise"
                    ? "bg-emerald-500"
                    : "bg-accent";
            return (
              <div key={key} className="flex flex-1 flex-col items-center gap-2">
                <span className="font-mono text-xs text-foreground">{count}</span>
                <div className="flex w-full flex-col justify-end rounded-t-md bg-surface-2" style={{ height: 140 }}>
                  <div className={`w-full rounded-t-md ${color}`} style={{ height: `${Math.max(pct, 6)}%` }} />
                </div>
                <span className="text-center text-[10px] leading-tight text-muted">{labels[key] || key}</span>
              </div>
            );
          })}
        </div>
      </Panel>
    </div>
  );
}

function StoreInsightSection({ block }: { block: StoreInsightBlock }) {
  const blockStats = block.statistics ?? {};
  const ratingDist = list<Bucket>(blockStats.rating_distribution);
  const sentiment = (blockStats.sentiment_breakdown ?? {}) as Record<string, SentimentBucket>;
  const categoryInsights = list<CategoryInsight>(block.category_insights);
  const featureSuggestions = list<FeedbackSuggestion>(block.feature_suggestions);
  const bugSuggestions = list<FeedbackSuggestion>(block.bug_suggestions);
  const featuredReviews = list<FeaturedReview>(block.featured_reviews);
  const totalReviews = Number(blockStats.total_reviews ?? 0);
  const avgRating = Number(blockStats.avg_rating ?? 0);

  return (
    <section className="space-y-6 rounded-2xl border border-border bg-surface/40 p-5">
      <div>
        <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">{block.label ?? block.store}</p>
        <h2 className="mt-2 text-lg font-medium text-foreground">{block.label ?? block.store} analizi</h2>
        <p className="mt-1 text-sm text-muted">
          {totalReviews} yorum · ort. {num(avgRating)}★
        </p>
      </div>

      {ratingDist.length > 0 && (
        <div className="grid gap-4 lg:grid-cols-2">
          <Panel>
            <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Yıldız dağılımı</h3>
            <div className="space-y-3">
              {ratingDist.map((b) => (
                <BarRow
                  key={b.star}
                  label={`${b.star} ★`}
                  pct={Number(b.pct ?? 0)}
                  count={Number(b.count ?? 0)}
                  tone={
                    (b.star ?? 0) <= 2 ? "bg-red-500" : (b.star ?? 0) === 3 ? "bg-amber-500" : "bg-accent"
                  }
                />
              ))}
            </div>
          </Panel>
          <Panel>
            <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Duygu özeti</h3>
            <div className="space-y-4">
              {[
                { key: "negative", label: "Olumsuz", cls: "bg-red-500/10 text-red-500 border-red-500/20" },
                { key: "neutral", label: "Nötr", cls: "bg-amber-500/10 text-amber-500 border-amber-500/20" },
                { key: "positive", label: "Olumlu", cls: "bg-emerald-500/10 text-emerald-500 border-emerald-500/20" },
              ].map((s) => {
                const b = sentiment[s.key] ?? {};
                return (
                  <div key={s.key} className="flex items-end justify-between border-b border-border pb-3 last:border-0">
                    <span className={`rounded-md border px-2 py-0.5 text-xs font-medium uppercase ${s.cls}`}>
                      {s.label}
                    </span>
                    <div className="text-right">
                      <p className="font-mono text-xl text-foreground">{b.count ?? 0}</p>
                      <p className="text-sm text-muted">%{num(b.pct, 1)}</p>
                    </div>
                  </div>
                );
              })}
            </div>
          </Panel>
        </div>
      )}

      {categoryInsights.length > 0 && (
        <div>
          <h3 className="font-medium text-foreground">Kategori kırılımı</h3>
          <div className="mt-4 grid gap-4 lg:grid-cols-2">
            {categoryInsights.map((cat) => (
              <Panel key={cat.category ?? cat.label}>
                <div className="flex items-center justify-between gap-2">
                  <h4 className="font-medium text-foreground">{cat.label ?? cat.category}</h4>
                  <span className="font-mono text-xs text-muted">
                    {cat.count ?? 0} · %{num(cat.pct, 1)}
                  </span>
                </div>
                <ReviewQuotes quotes={list<ReviewQuote>(cat.reviews)} />
              </Panel>
            ))}
          </div>
        </div>
      )}

      {(featureSuggestions.length > 0 || bugSuggestions.length > 0) && (
        <div className="grid gap-4 lg:grid-cols-2">
          <SuggestionList title="Özellik önerileri" items={featureSuggestions} tone="feature" />
          <SuggestionList title="Hata / şikâyetler" items={bugSuggestions} tone="bug" />
        </div>
      )}

      {featuredReviews.length > 0 && (
        <div className="grid gap-4 lg:grid-cols-2">
          {featuredReviews.slice(0, 4).map((rv, i) => (
            <Panel key={i}>
              <div className="mb-2 flex flex-wrap items-center gap-2">
                <span className="font-mono text-sm text-foreground">{rv.rating ?? "?"}★</span>
                {rv.category && (
                  <span className={`rounded border px-1.5 py-0.5 text-[10px] ${categoryBadge(String(rv.category))}`}>
                    {rv.category}
                  </span>
                )}
              </div>
              <p className="text-sm leading-relaxed text-foreground">&ldquo;{rv.text}&rdquo;</p>
            </Panel>
          ))}
        </div>
      )}
    </section>
  );
}

function SuggestionList({
  title,
  items,
  tone,
}: {
  title: string;
  items: FeedbackSuggestion[];
  tone: "feature" | "bug";
}) {
  if (!items.length) return null;
  return (
    <Panel>
      <h3 className="font-mono text-xs uppercase tracking-wider text-muted">{title}</h3>
      <div className="mt-4 space-y-4">
        {items.map((item, i) => (
          <div key={i} className="border-b border-border pb-4 last:border-0">
            <div className="flex flex-wrap items-start justify-between gap-2">
              <h4 className="font-medium text-foreground">{item.title}</h4>
              <div className="flex gap-2">
                <span className={`rounded border px-2 py-0.5 text-[10px] uppercase ${categoryBadge(tone)}`}>
                  {tone}
                </span>
                {item.priority && (
                  <span className="rounded border border-accent/30 bg-accent/5 px-2 py-0.5 font-mono text-[10px] text-accent">
                    {item.priority}
                  </span>
                )}
              </div>
            </div>
            {item.summary && <p className="mt-2 text-sm text-muted">{item.summary}</p>}
            <ReviewQuotes quotes={list<ReviewQuote>(item.supporting_reviews)} />
          </div>
        ))}
      </div>
    </Panel>
  );
}

export default function AuditReportView({ audit, insights }: { audit: Audit; insights: AuditInsights }) {
  const stats = insights.statistics ?? {};
  const meta = (insights.report_meta ?? stats.report_meta ?? {}) as Record<string, unknown>;
  const ratingDist = list<Bucket>(stats.rating_distribution);
  const sentiment = (stats.sentiment_breakdown ?? {}) as Record<string, SentimentBucket>;
  const categoryInsights = list<CategoryInsight>(insights.category_insights);
  const featureSuggestions = list<FeedbackSuggestion>(insights.feature_suggestions);
  const bugSuggestions = list<FeedbackSuggestion>(insights.bug_suggestions);
  const featuredReviews = list<FeaturedReview>(insights.featured_reviews);
  const timeline = list<TimelineItem>(meta.timeline);
  const priorities = list<Priority>(meta.priorities);
  const findings = list<Finding>(meta.management_findings);
  const currentAvg = Number(meta.current_avg_rating ?? stats.avg_rating ?? 0);
  const stretchGoal = Number(meta.stretch_goal_rating ?? Math.min(5, currentAvg + 0.45));
  const callout = String(meta.callout ?? "");
  const totalReviews = Number(stats.total_reviews ?? audit.total_reviews ?? 0);

  const playCount = Number(stats.play_count ?? audit.play_fetched ?? 0);
  const appStoreCount = Number(stats.appstore_count ?? audit.appstore_fetched ?? 0);
  const storeBreakdown = list<StoreInsightBlock>(stats.store_breakdown);
  const hasDualStore = storeBreakdown.length >= 2;
  const showPlay = Boolean(audit.play_app_id) || playCount > 0;
  const showAppStore = Boolean(audit.appstore_app_id) || appStoreCount > 0;

  return (
    <div className="space-y-8">
      <header className="border-b border-border pb-8">
        <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">InferReview · Müşteri Raporu</p>
        <h1 className="mt-3 text-2xl font-medium text-foreground">{audit.client_name}</h1>
        <p className="mt-2 text-sm text-muted">{audit.app_display_name}</p>
        <div className="mt-4 flex flex-wrap gap-x-5 gap-y-1 font-mono text-xs text-muted">
          {showPlay && (
            <span>
              Play Store: <span className="text-foreground">{playCount}</span>
            </span>
          )}
          {showAppStore && (
            <span>
              App Store: <span className="text-foreground">{appStoreCount}</span>
            </span>
          )}
          <span>
            Mod: <span className="text-foreground">{audit.mode}</span>
          </span>
          {audit.country && (
            <span>
              Ülke: <span className="text-foreground">{audit.country.toUpperCase()}</span>
            </span>
          )}
        </div>
        {callout && (
          <div className="mt-4 rounded-xl border border-accent/30 bg-accent/5 px-4 py-3 text-sm leading-relaxed text-foreground">
            {callout}
          </div>
        )}
      </header>

      {hasDualStore && (
        <section className="space-y-6">
          <div>
            <h2 className="text-lg font-medium text-foreground">Mağaza bazlı analiz</h2>
            <p className="mt-1 text-sm text-muted">Her mağaza için ayrı istatistik ve geri bildirim özeti.</p>
          </div>
          {storeBreakdown.map((block) => (
            <StoreInsightSection key={block.store ?? block.label} block={block} />
          ))}
        </section>
      )}

      {hasDualStore && (
        <div className="border-t border-border pt-2">
          <h2 className="text-lg font-medium text-foreground">Birleşik analiz</h2>
          <p className="mt-1 text-sm text-muted">Her iki mağazanın yorumları birlikte değerlendirildi.</p>
        </div>
      )}

      <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Stat label="Yazılı yorum" value={String(totalReviews)} hint="Analiz edilen metin" />
        <Stat label="Ort. puan" value={num(currentAvg)} hint="Yazılı yorum ortalaması" accent />
        <Stat label="Mevcut taban" value={num(currentAvg)} hint="İyileştirme başlangıcı" />
        <Stat label="Hedef bandı" value={`~${num(stretchGoal, 1)}`} hint="Veriye göre anlamlı hedef" accent />
      </section>

      {ratingDist.length > 0 && (
        <section>
          <h2 className="text-lg font-medium text-foreground">Yazılı yorum dağılımı</h2>
          <p className="mt-1 text-sm text-muted">Metin içeren erişilebilir yorumların yıldız kırılımı.</p>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            <Panel>
              <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Yıldız bazlı kırılım</h3>
              <div className="space-y-3">
                {ratingDist.map((b) => (
                  <BarRow
                    key={b.star}
                    label={`${b.star} ★`}
                    pct={Number(b.pct ?? 0)}
                    count={Number(b.count ?? 0)}
                    tone={
                      (b.star ?? 0) <= 2
                        ? "bg-red-500"
                        : (b.star ?? 0) === 3
                          ? "bg-amber-500"
                          : "bg-accent"
                    }
                  />
                ))}
              </div>
            </Panel>
            <Panel>
              <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Duygu özeti (yazılı)</h3>
              <div className="space-y-4">
                {[
                  { key: "negative", label: "Olumsuz", sub: "1–2 yıldız", cls: "bg-red-500/10 text-red-500 border-red-500/20" },
                  { key: "neutral", label: "Nötr", sub: "3 yıldız", cls: "bg-amber-500/10 text-amber-500 border-amber-500/20" },
                  { key: "positive", label: "Olumlu", sub: "4–5 yıldız", cls: "bg-emerald-500/10 text-emerald-500 border-emerald-500/20" },
                ].map((s) => {
                  const b = sentiment[s.key] ?? {};
                  return (
                    <div key={s.key} className="flex items-end justify-between border-b border-border pb-3 last:border-0">
                      <div>
                        <span className={`rounded-md border px-2 py-0.5 text-xs font-medium uppercase ${s.cls}`}>
                          {s.label}
                        </span>
                        <p className="mt-1 text-sm text-muted">{s.sub}</p>
                      </div>
                      <div className="text-right">
                        <p className="font-mono text-2xl text-foreground">{b.count ?? 0}</p>
                        <p className="text-sm text-muted">%{num(b.pct, 1)}</p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </Panel>
          </div>
        </section>
      )}

      {categoryInsights.length > 0 && (
        <section>
          <h2 className="text-lg font-medium text-foreground">Kategori analizi</h2>
          <p className="mt-1 text-sm text-muted">
            Sınıflandırılmış yorum kategorileri ve her kategoriden örnek alıntılar.
          </p>
          <div className="mt-5 space-y-4">
            <CategoryChart insights={categoryInsights} total={totalReviews} />
            <div className="grid gap-4 lg:grid-cols-2">
              {categoryInsights.map((cat) => (
                <Panel key={cat.category ?? cat.label}>
                  <div className="flex items-center justify-between gap-2">
                    <h3 className="font-medium text-foreground">{cat.label ?? cat.category}</h3>
                    <span className="font-mono text-xs text-muted">
                      {cat.count ?? 0} · %{num(cat.pct, 1)}
                    </span>
                  </div>
                  <ReviewQuotes quotes={list<ReviewQuote>(cat.reviews)} />
                </Panel>
              ))}
            </div>
          </div>
        </section>
      )}

      {(featureSuggestions.length > 0 || bugSuggestions.length > 0) && (
        <section>
          <h2 className="text-lg font-medium text-foreground">Kullanıcı geri bildirimi</h2>
          <p className="mt-1 text-sm text-muted">
            Yorumlarla desteklenen özellik önerileri ve hata / şikâyet başlıkları.
          </p>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            <SuggestionList title="Özellik önerileri (feature)" items={featureSuggestions} tone="feature" />
            <SuggestionList title="Hata ve şikâyetler (bug)" items={bugSuggestions} tone="bug" />
          </div>
        </section>
      )}

      {featuredReviews.length > 0 && (
        <section>
          <h2 className="text-lg font-medium text-foreground">Öne çıkan yorumlar</h2>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            {featuredReviews.map((rv, i) => (
              <Panel key={i}>
                <div className="mb-2 flex flex-wrap items-center gap-2">
                  <span className="font-mono text-sm text-foreground">{rv.rating ?? "?"}★</span>
                  {rv.store && <span className="font-mono text-[10px] uppercase text-muted">{rv.store}</span>}
                  {rv.category && (
                    <span className={`rounded border px-1.5 py-0.5 text-[10px] ${categoryBadge(String(rv.category))}`}>
                      {rv.category}
                    </span>
                  )}
                  {rv.highlight && (
                    <span className="rounded border border-accent/30 bg-accent/5 px-2 py-0.5 text-[10px] text-accent">
                      {rv.highlight}
                    </span>
                  )}
                </div>
                <p className="text-sm leading-relaxed text-foreground">&ldquo;{rv.text}&rdquo;</p>
              </Panel>
            ))}
          </div>
        </section>
      )}

      {priorities.length > 0 && (
        <section>
          <h2 className="text-lg font-medium text-foreground">Öncelik sırası</h2>
          <Panel className="mt-5">
            {priorities.map((p) => (
              <div key={p.rank ?? p.title} className="grid grid-cols-[42px_1fr] gap-3 border-b border-border py-4 last:border-0">
                <div className="flex h-[42px] w-[42px] items-center justify-center rounded-xl border border-accent/30 bg-accent/5 font-mono text-sm font-medium text-accent">
                  {String(p.rank ?? "").padStart(2, "0")}
                </div>
                <div>
                  <h4 className="font-medium text-foreground">{p.title}</h4>
                  <p className="mt-1 text-sm text-muted">{p.body}</p>
                </div>
              </div>
            ))}
          </Panel>
        </section>
      )}

      {(timeline.length > 0 || (insights.action_plan?.length ?? 0) > 0) && (
        <section>
          <h2 className="text-lg font-medium text-foreground">90 günlük aksiyon planı</h2>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            {(timeline.length ? timeline : insights.action_plan ?? []).map((item, i) => {
              const t = timeline.length ? (item as TimelineItem) : null;
              const ap = !timeline.length ? (item as Record<string, unknown>) : null;
              return (
                <Panel key={i}>
                  <span
                    className={`inline-block rounded-md border px-2 py-0.5 text-[0.72rem] font-medium uppercase ${tagClass(
                      String(t?.tag ?? ap?.tag ?? "mid")
                    )}`}
                  >
                    {String(t?.horizon ?? ap?.horizon_days ?? "90 gün")}
                  </span>
                  <h4 className="mt-3 font-medium text-foreground">
                    {String(t?.title ?? ap?.title ?? ap?.action ?? `Aksiyon ${i + 1}`)}
                  </h4>
                  <p className="mt-2 text-sm text-muted">{String(t?.body ?? ap?.action ?? ap?.expected_impact ?? "")}</p>
                </Panel>
              );
            })}
          </div>
        </section>
      )}

      <section>
        <h2 className="text-lg font-medium text-foreground">Yönetim özeti</h2>
        <Panel className="mt-5">
          <p className="whitespace-pre-wrap text-sm leading-relaxed text-foreground">{insights.executive_summary}</p>
        </Panel>
        {findings.length > 0 && (
          <div className="mt-4 space-y-3">
            {findings.map((f, i) => (
              <Panel key={i}>
                <h4 className="font-medium text-foreground">{f.title}</h4>
                <p className="mt-1 text-sm text-muted">{f.body}</p>
              </Panel>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
