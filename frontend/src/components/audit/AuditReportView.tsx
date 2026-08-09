import type { Audit, AuditInsights } from "@/types";

type Bucket = { star?: number; count?: number; pct?: number };
type SentimentBucket = { count?: number; pct?: number };
type Theme = { theme?: string; label?: string; count?: number; pct?: number };
type Scenario = {
  id?: string;
  label?: string;
  pace?: string;
  title?: string;
  summary?: string;
  timeline?: string;
  highlight?: string;
};
type TimelineItem = { horizon?: string; tag?: string; title?: string; body?: string };
type Priority = { rank?: number; title?: string; body?: string };
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

function paceLabel(pace?: string) {
  if (pace === "slow") return "Yavaş";
  if (pace === "fast") return "Önerilen";
  return "Orta";
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

export default function AuditReportView({ audit, insights }: { audit: Audit; insights: AuditInsights }) {
  const stats = insights.statistics ?? {};
  const meta = (insights.report_meta ?? stats.report_meta ?? {}) as Record<string, unknown>;
  const ratingDist = list<Bucket>(stats.rating_distribution);
  const sentiment = (stats.sentiment_breakdown ?? {}) as Record<string, SentimentBucket>;
  const themes = list<Theme>(stats.theme_intensity);
  const rootCauses = insights.root_causes ?? [];
  const scenarios = list<Scenario>(meta.scenarios);
  const timeline = list<TimelineItem>(meta.timeline);
  const priorities = list<Priority>(meta.priorities);
  const findings = list<Finding>(meta.management_findings);
  const currentAvg = Number(meta.current_avg_rating ?? stats.avg_rating ?? 0);
  const stretchGoal = Number(meta.stretch_goal_rating ?? Math.min(5, currentAvg + 0.45));
  const callout = String(meta.callout ?? "");

  const playCount = Number(stats.play_count ?? audit.play_fetched ?? 0);
  const appStoreCount = Number(stats.appstore_count ?? audit.appstore_fetched ?? 0);

  return (
    <div className="space-y-8">
      <header className="border-b border-border pb-8">
        <p className="font-mono text-xs uppercase tracking-[0.2em] text-accent">InferReview · Müşteri Raporu</p>
        <h1 className="mt-3 text-2xl font-medium text-foreground">{audit.client_name}</h1>
        <p className="mt-2 text-sm text-muted">{audit.app_display_name}</p>
        <div className="mt-4 flex flex-wrap gap-x-5 gap-y-1 font-mono text-xs text-muted">
          <span>
            Play: <span className="text-foreground">{playCount}</span>
          </span>
          <span>
            App Store: <span className="text-foreground">{appStoreCount}</span>
          </span>
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

      <section className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Stat label="Yazılı yorum" value={String(stats.total_reviews ?? audit.total_reviews)} hint="Analiz edilen metin" />
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

      {(themes.length > 0 || rootCauses.length > 0) && (
        <section>
          <h2 className="text-lg font-medium text-foreground">Kök nedenler</h2>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            {themes.length > 0 && (
              <Panel>
                <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Tema yoğunluğu</h3>
                <div className="space-y-3">
                  {themes.map((t, i) => (
                    <div key={t.theme ?? i} className="text-sm">
                      <div className="mb-1 flex justify-between gap-2">
                        <span className="text-foreground">{t.label ?? t.theme}</span>
                        <span className="font-mono text-xs text-muted">{t.count ?? 0}</span>
                      </div>
                      <div className="h-2 overflow-hidden rounded-full bg-surface-2">
                        <div
                          className="h-full rounded-full bg-red-500"
                          style={{ width: `${Math.min(Number(t.pct ?? 0), 100)}%` }}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </Panel>
            )}
            <Panel>
              <h3 className="mb-4 font-mono text-xs uppercase tracking-wider text-muted">Detaylı bulgular</h3>
              <ul className="space-y-4">
                {(rootCauses.length ? rootCauses : themes).map((raw, i) => {
                  const rc = raw as Record<string, unknown>;
                  return (
                    <li key={i} className="border-b border-border pb-4 last:border-0">
                      <p className="font-medium text-foreground">{String(rc.theme ?? rc.label ?? `Tema ${i + 1}`)}</p>
                      <p className="mt-1 text-sm text-muted">
                        {String(rc.description ?? "")}
                        {rc.sample_count != null ? ` (${String(rc.sample_count)} örnek)` : ""}
                      </p>
                    </li>
                  );
                })}
              </ul>
            </Panel>
          </div>
        </section>
      )}

      {scenarios.length > 0 && (
        <section>
          <h2 className="text-lg font-medium text-foreground">Hedefe giden senaryolar</h2>
          <p className="mt-1 text-sm text-muted">
            Mevcut {num(currentAvg)} ortalamadan ~{num(stretchGoal, 1)} bandına giden gerçekçi yollar.
          </p>
          <div className="mt-5 grid gap-4 lg:grid-cols-3">
            {scenarios.map((s) => (
              <Panel key={s.id ?? s.title}>
                <span
                  className={`inline-block rounded-md border px-2 py-0.5 text-[0.72rem] font-medium uppercase ${tagClass(
                    s.pace === "slow" ? "hard" : s.pace === "fast" ? "fast" : "mid"
                  )}`}
                >
                  {paceLabel(s.pace) || s.label}
                </span>
                <h3 className="mt-3 text-base font-medium text-foreground">{s.title}</h3>
                <p className="mt-2 text-sm text-muted">{s.summary}</p>
                {s.highlight && <p className="mt-3 text-sm font-medium text-accent">{s.highlight}</p>}
                {s.timeline && <p className="mt-2 text-xs text-muted">Süre: {s.timeline}</p>}
              </Panel>
            ))}
          </div>
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
