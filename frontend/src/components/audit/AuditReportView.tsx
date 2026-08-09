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
  if (tag === "fast") return "bg-emerald-500/15 text-emerald-700";
  if (tag === "hard") return "bg-red-500/15 text-red-700";
  return "bg-amber-500/15 text-amber-800";
}

function paceLabel(pace?: string) {
  if (pace === "slow") return "Yavaş";
  if (pace === "fast") return "Önerilen";
  return "Orta";
}

function BarRow({ label, pct, count, tone }: { label: string; pct: number; count: number; tone: string }) {
  return (
    <div className="space-y-1 text-sm">
      <div className="flex items-center justify-between gap-2">
        <span>{label}</span>
        <strong className="text-xs">
          {count} · %{num(pct, 1)}
        </strong>
      </div>
      <div className="h-2.5 overflow-hidden rounded-full bg-[#0f1c2414]">
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
    <div className="consult-report text-[#0f1c24]">
      <style jsx global>{`
        .consult-report {
          font-family: "Manrope", system-ui, sans-serif;
          background:
            radial-gradient(circle at 12% 18%, rgba(13, 110, 110, 0.16), transparent 34%),
            radial-gradient(circle at 88% 8%, rgba(196, 92, 38, 0.12), transparent 28%),
            linear-gradient(180deg, #eef3f1 0%, #f3f6f4 45%, #e8eeeb 100%);
          border-radius: 1.25rem;
          padding: 2rem 1.5rem 2.5rem;
        }
        .consult-report h1,
        .consult-report h2 {
          font-family: "DM Serif Display", Georgia, serif;
          font-weight: 400;
        }
      `}</style>

      <header className="border-b border-[#0f1c241f] pb-8">
        <p className="text-xs font-bold uppercase tracking-[0.18em] text-[#0d6e6e]">
          InferReview · Müşteri Raporu
        </p>
        <h1 className="mt-3 max-w-2xl text-4xl leading-tight">{audit.client_name}</h1>
        <p className="mt-3 max-w-2xl text-[#5a6b75]">{audit.app_display_name}</p>
        <div className="mt-4 flex flex-wrap gap-x-5 gap-y-1 text-sm text-[#5a6b75]">
          <span>
            <strong className="text-[#0f1c24]">Play:</strong> {playCount} yorum
          </span>
          <span>
            <strong className="text-[#0f1c24]">App Store:</strong> {appStoreCount} yorum
          </span>
          <span>
            <strong className="text-[#0f1c24]">Mod:</strong> {audit.mode}
          </span>
        </div>
        {callout && (
          <div className="mt-4 rounded-r-xl border-l-[3px] border-[#0d6e6e] bg-[#0d6e6e14] px-4 py-3 text-sm leading-relaxed">
            {callout}
          </div>
        )}
      </header>

      <section className="mt-8 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[
          { label: "Yazılı yorum", value: String(stats.total_reviews ?? audit.total_reviews), hint: "Analiz edilen metin" },
          { label: "Ort. puan", value: num(currentAvg), hint: "Yazılı yorum ortalaması" },
          { label: "Mevcut taban", value: num(currentAvg), hint: "İyileştirme başlangıcı" },
          { label: "Hedef bandı", value: `~${num(stretchGoal, 1)}`, hint: "Veriye göre anlamlı hedef" },
        ].map((kpi) => (
          <article
            key={kpi.label}
            className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)] backdrop-blur"
          >
            <p className="text-[0.78rem] font-bold uppercase tracking-[0.08em] text-[#5a6b75]">{kpi.label}</p>
            <p className="mt-2 font-serif text-4xl leading-none">{kpi.value}</p>
            <p className="mt-2 text-sm text-[#5a6b75]">{kpi.hint}</p>
          </article>
        ))}
      </section>

      {ratingDist.length > 0 && (
        <section className="mt-10">
          <h2 className="text-3xl">Yazılı yorum dağılımı</h2>
          <p className="mt-2 max-w-3xl text-[#5a6b75]">Metin içeren erişilebilir yorumların yıldız kırılımı.</p>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            <div className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
              <h3 className="mb-4 text-sm font-semibold">Yıldız bazlı kırılım</h3>
              <div className="space-y-3">
                {ratingDist.map((b) => (
                  <BarRow
                    key={b.star}
                    label={`${b.star} ★`}
                    pct={Number(b.pct ?? 0)}
                    count={Number(b.count ?? 0)}
                    tone={
                      (b.star ?? 0) <= 2
                        ? "bg-[#b42318]"
                        : (b.star ?? 0) === 3
                          ? "bg-[#c9a227]"
                          : "bg-[#0d6e6e]"
                    }
                  />
                ))}
              </div>
            </div>
            <div className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
              <h3 className="mb-4 text-sm font-semibold">Duygu özeti (yazılı)</h3>
              <div className="space-y-4">
                {[
                  { key: "negative", label: "Olumsuz", sub: "1–2 yıldız", cls: "bg-[#b423181f] text-[#b42318]" },
                  { key: "neutral", label: "Nötr", sub: "3 yıldız", cls: "bg-[#9a67001f] text-[#9a6700]" },
                  { key: "positive", label: "Olumlu", sub: "4–5 yıldız", cls: "bg-[#1f7a4c1f] text-[#1f7a4c]" },
                ].map((s) => {
                  const b = sentiment[s.key] ?? {};
                  return (
                    <div key={s.key} className="flex items-end justify-between border-b border-[#0f1c241f] pb-3 last:border-0">
                      <div>
                        <span className={`rounded-md px-2 py-0.5 text-xs font-bold uppercase ${s.cls}`}>{s.label}</span>
                        <p className="mt-1 text-sm text-[#5a6b75]">{s.sub}</p>
                      </div>
                      <div className="text-right">
                        <strong className="text-2xl">{b.count ?? 0}</strong>
                        <p className="text-sm text-[#5a6b75]">%{num(b.pct, 1)}</p>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </section>
      )}

      {(themes.length > 0 || rootCauses.length > 0) && (
        <section className="mt-10">
          <h2 className="text-3xl">Kök nedenler</h2>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            {themes.length > 0 && (
              <div className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
                <h3 className="mb-4 text-sm font-semibold">Tema yoğunluğu</h3>
                <div className="space-y-3">
                  {themes.map((t, i) => (
                    <div key={t.theme ?? i} className="grid grid-cols-[1fr_56px] items-center gap-3 text-sm">
                      <div>
                        <div className="mb-1 flex justify-between gap-2">
                          <span>{t.label ?? t.theme}</span>
                          <strong>{t.count ?? 0}</strong>
                        </div>
                        <div className="h-2.5 overflow-hidden rounded-full bg-[#0f1c2414]">
                          <div className="h-full rounded-full bg-[#b42318]" style={{ width: `${Math.min(Number(t.pct ?? 0), 100)}%` }} />
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
            <div className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
              <h3 className="mb-4 text-sm font-semibold">Detaylı bulgular</h3>
              <ul className="space-y-4">
                {(rootCauses.length ? rootCauses : themes).map((raw, i) => {
                  const rc = raw as Record<string, unknown>;
                  return (
                  <li key={i} className="border-b border-[#0f1c241f] pb-4 last:border-0">
                    <p className="font-semibold">{String(rc.theme ?? rc.label ?? `Tema ${i + 1}`)}</p>
                    <p className="mt-1 text-sm text-[#5a6b75]">
                      {String(rc.description ?? "")}
                      {rc.sample_count != null ? ` (${String(rc.sample_count)} örnek)` : ""}
                    </p>
                  </li>
                  );
                })}
              </ul>
            </div>
          </div>
        </section>
      )}

      {scenarios.length > 0 && (
        <section className="mt-10">
          <h2 className="text-3xl">Hedefe giden senaryolar</h2>
          <p className="mt-2 max-w-3xl text-[#5a6b75]">
            Mevcut {num(currentAvg)} ortalamadan ~{num(stretchGoal, 1)} bandına giden gerçekçi yollar (sabit puan hedefi yok).
          </p>
          <div className="mt-5 grid gap-4 lg:grid-cols-3">
            {scenarios.map((s) => (
              <article key={s.id ?? s.title} className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
                <span className={`inline-block rounded-md px-2 py-0.5 text-[0.72rem] font-extrabold uppercase ${tagClass(s.pace === "slow" ? "hard" : s.pace === "fast" ? "fast" : "mid")}`}>
                  {paceLabel(s.pace) || s.label}
                </span>
                <h3 className="mt-3 text-base font-semibold">{s.title}</h3>
                <p className="mt-2 text-sm text-[#5a6b75]">{s.summary}</p>
                {s.highlight && <p className="mt-3 text-sm font-semibold text-[#0d6e6e]">{s.highlight}</p>}
                {s.timeline && <p className="mt-2 text-xs text-[#5a6b75]">Süre: {s.timeline}</p>}
              </article>
            ))}
          </div>
        </section>
      )}

      {(timeline.length > 0 || (insights.action_plan?.length ?? 0) > 0) && (
        <section className="mt-10">
          <h2 className="text-3xl">90 günlük aksiyon planı</h2>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            {(timeline.length ? timeline : insights.action_plan ?? []).map((item, i) => {
              const t = timeline.length ? (item as TimelineItem) : null;
              const ap = !timeline.length ? (item as Record<string, unknown>) : null;
              return (
                <div key={i} className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
                  <span className={`inline-block rounded-md px-2 py-0.5 text-[0.72rem] font-extrabold uppercase ${tagClass(String(t?.tag ?? ap?.tag ?? "mid"))}`}>
                    {String(t?.horizon ?? ap?.horizon_days ?? "90 gün")}
                  </span>
                  <h4 className="mt-3 font-semibold">{String(t?.title ?? ap?.title ?? ap?.action ?? `Aksiyon ${i + 1}`)}</h4>
                  <p className="mt-2 text-sm text-[#5a6b75]">{String(t?.body ?? ap?.action ?? ap?.expected_impact ?? "")}</p>
                </div>
              );
            })}
          </div>
        </section>
      )}

      {priorities.length > 0 && (
        <section className="mt-10">
          <h2 className="text-3xl">Öncelik sırası</h2>
          <div className="mt-5 rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
            {priorities.map((p) => (
              <div key={p.rank ?? p.title} className="grid grid-cols-[42px_1fr] gap-3 border-b border-[#0f1c241f] py-4 last:border-0">
                <div className="flex h-[42px] w-[42px] items-center justify-center rounded-xl bg-[#0d6e6e1a] font-extrabold text-[#0d6e6e]">
                  {String(p.rank ?? "").padStart(2, "0")}
                </div>
                <div>
                  <h4 className="font-semibold">{p.title}</h4>
                  <p className="mt-1 text-sm text-[#5a6b75]">{p.body}</p>
                </div>
              </div>
            ))}
          </div>
        </section>
      )}

      <section className="mt-10">
        <h2 className="text-3xl">Yönetim özeti</h2>
        <div className="mt-5 rounded-[18px] border border-[#0f1c241f] bg-white/75 p-5 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
          <p className="whitespace-pre-wrap text-sm leading-relaxed">{insights.executive_summary}</p>
        </div>
        {findings.length > 0 && (
          <div className="mt-4 space-y-3">
            {findings.map((f, i) => (
              <div key={i} className="rounded-[18px] border border-[#0f1c241f] bg-white/75 p-4 shadow-[0_18px_50px_rgba(15,28,36,0.08)]">
                <h4 className="font-semibold">{f.title}</h4>
                <p className="mt-1 text-sm text-[#5a6b75]">{f.body}</p>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
