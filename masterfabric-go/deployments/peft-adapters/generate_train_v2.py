#!/usr/bin/env python3
"""Generate train_v2.jsonl aligned with production Analyze prompt (Go mlc client)."""

import json
from pathlib import Path

PROMPT_PREFIX = """You are a strict classifier for app store reviews (any language).
Classify into exactly one category and one sentiment.
Categories: bug, feature, praise, spam, other.
Sentiments: positive, negative, neutral.

Rules:
- Match the reviewer's tone: complaints and dissatisfaction → negative; compliments → positive; factual/neutral → neutral.
- bug: crashes, errors, broken or slow functionality.
- feature: requests for new capability.
- praise: explicit compliments.
- spam: promotional junk or fake reviews.
- other: general feedback that does not fit above (still use the correct sentiment).

Examples:
{"category":"bug","sentiment":"negative"} — "App keeps crashing"
{"category":"other","sentiment":"negative"} — "This app is terrible" / "Kötü bir uygulama"
{"category":"praise","sentiment":"positive"} — "Love this app!"

Respond with ONLY a JSON object and nothing else.

Review: """

# (review_text, category, sentiment)
EXAMPLES: list[tuple[str, str, str]] = [
    # --- bug (48) ---
    ("App keeps crashing when I open settings", "bug", "negative"),
    ("Uygulama açılır açılmaz kapanıyor", "bug", "negative"),
    ("Can't login, always shows error 500", "bug", "negative"),
    ("Giriş yapamıyorum sürekli hata veriyor", "bug", "negative"),
    ("Very slow and freezes on my phone", "bug", "negative"),
    ("Telefonumda çok yavaş donuyor", "bug", "negative"),
    ("Notifications stopped working after update", "bug", "negative"),
    ("Güncellemeden sonra bildirimler gelmiyor", "bug", "negative"),
    ("Payment failed but money was deducted", "bug", "negative"),
    ("Ödeme alındı ama işlem başarısız diyor", "bug", "negative"),
    ("Screen goes black when uploading photos", "bug", "negative"),
    ("Fotoğraf yüklerken ekran kararıyor", "bug", "negative"),
    ("Sync never completes, stuck at 99%", "bug", "negative"),
    ("Senkronizasyon hiç bitmiyor", "bug", "negative"),
    ("Audio cuts out during calls", "bug", "negative"),
    ("Aramalarda ses kesiliyor", "bug", "negative"),
    ("Battery drain is insane since last patch", "bug", "negative"),
    ("Son güncellemeden beri pil inanılmaz hızlı bitiyor", "bug", "negative"),
    ("Maps won't load, blank screen", "bug", "negative"),
    ("Haritalar açılmıyor boş ekran", "bug", "negative"),
    ("Keyboard covers the text field on iOS", "bug", "negative"),
    ("Klavye yazı alanını kapatıyor", "bug", "negative"),
    ("Logout button does nothing", "bug", "negative"),
    ("Çıkış yap butonu çalışmıyor", "bug", "negative"),
    ("Camera permission granted but camera still fails", "bug", "negative"),
    ("Kamera izni var ama yine de açılmıyor", "bug", "negative"),
    ("App crashes on Android 14", "bug", "negative"),
    ("Android 14'te sürekli çöküyor", "bug", "negative"),
    ("Data loss after force close", "bug", "negative"),
    ("Zorla kapatınca verilerim silindi", "bug", "negative"),
    ("WiFi only mode broken, uses mobile data", "bug", "negative"),
    ("Sadece wifi modu bozuk mobil veri harcıyor", "bug", "negative"),
    ("Search returns no results even for exact match", "bug", "negative"),
    ("Tam arama yapsam da sonuç çıkmıyor", "bug", "negative"),
    ("Widget shows wrong information", "bug", "negative"),
    ("Widget yanlış bilgi gösteriyor", "bug", "negative"),
    ("Export PDF corrupted file", "bug", "negative"),
    ("PDF dışa aktarma bozuk dosya üretiyor", "bug", "negative"),
    ("Touch targets misaligned after UI update", "bug", "negative"),
    ("Arayüz güncellemesinden sonra dokunma bozuk", "bug", "negative"),
    ("Sometimes works sometimes doesn't very frustrating", "bug", "negative"),
    ("Bazen çalışıyor bazen çalışmıyor sinir bozucu", "bug", "negative"),
    ("Registration email never arrives", "bug", "negative"),
    ("Kayıt maili hiç gelmiyor", "bug", "negative"),
    ("App size 2GB but basic features missing", "bug", "negative"),
    ("2GB yer kaplıyor temel özellikler eksik", "bug", "negative"),
    ("Bluetooth pairing fails every time", "bug", "negative"),
    ("Bluetooth eşleşmesi her seferinde başarısız", "bug", "negative"),
    # --- feature (48) ---
    ("Please add dark mode", "feature", "neutral"),
    ("Karanlık mod ekleyin lütfen", "feature", "neutral"),
    ("Would love offline support for flights", "feature", "positive"),
    ("Uçak modunda da çalışsa harika olur", "feature", "positive"),
    ("Need iPad version with split screen", "feature", "neutral"),
    ("iPad için split screen desteği lazım", "feature", "neutral"),
    ("Add widget for home screen", "feature", "neutral"),
    ("Ana ekran widget'ı ekleyin", "feature", "neutral"),
    ("Export to Excel would help my work", "feature", "positive"),
    ("Excel'e aktarma işimi kolaylaştırır", "feature", "positive"),
    ("Please support multiple languages in settings", "feature", "neutral"),
    ("Ayarlarda daha fazla dil olsun", "feature", "neutral"),
    ("Biometric login missing on Android", "feature", "negative"),
    ("Android'de parmak izi girişi yok yazık", "feature", "negative"),
    ("Why no Apple Watch app yet?", "feature", "negative"),
    ("Apple Watch uygulaması neden yok?", "feature", "negative"),
    ("Add filters to search results", "feature", "neutral"),
    ("Arama sonuçlarına filtre ekleyin", "feature", "neutral"),
    ("Collaboration mode for teams please", "feature", "positive"),
    ("Takım çalışması modu gelsin", "feature", "positive"),
    ("Schedule reminders would be useful", "feature", "positive"),
    ("Hatırlatıcı planlama özelliği faydalı olur", "feature", "positive"),
    ("Need two-factor authentication", "feature", "neutral"),
    ("İki faktörlü doğrulama eklenmeli", "feature", "neutral"),
    ("Please add CSV import", "feature", "neutral"),
    ("CSV içe aktarma olsun", "feature", "neutral"),
    ("Voice control for accessibility", "feature", "positive"),
    ("Erişilebilirlik için sesli kontrol", "feature", "positive"),
    ("Cloud backup option missing", "feature", "negative"),
    ("Bulut yedekleme seçeneği yok", "feature", "negative"),
    ("Add price alerts feature", "feature", "positive"),
    ("Fiyat alarmı özelliği ekleyin", "feature", "positive"),
    ("Would pay for premium if you add API access", "feature", "positive"),
    ("API erişimi eklerseniz premium alırım", "feature", "positive"),
    ("Need family sharing plan", "feature", "neutral"),
    ("Aile planı lazım", "feature", "neutral"),
    ("Sort reviews by date please", "feature", "neutral"),
    ("Yorumları tarihe göre sıralayın", "feature", "neutral"),
    ("Add undo for delete actions", "feature", "neutral"),
    ("Silme işlemi için geri al olsun", "feature", "neutral"),
    ("Custom themes would be nice", "feature", "positive"),
    ("Özel temalar güzel olurdu", "feature", "positive"),
    ("Integrate with Google Calendar", "feature", "neutral"),
    ("Google Takvim entegrasyonu", "feature", "neutral"),
    ("Batch edit for multiple items", "feature", "neutral"),
    ("Toplu düzenleme özelliği", "feature", "neutral"),
    ("Add landscape mode support", "feature", "neutral"),
    ("Yatay mod desteği ekleyin", "feature", "neutral"),
    # --- praise (40) ---
    ("Love this app, works perfectly!", "praise", "positive"),
    ("Harika uygulama kusursuz çalışıyor", "praise", "positive"),
    ("Best app in this category", "praise", "positive"),
    ("Bu kategorideki en iyi uygulama", "praise", "positive"),
    ("Clean UI and super fast", "praise", "positive"),
    ("Arayüz temiz ve çok hızlı", "praise", "positive"),
    ("Customer support replied quickly, impressed", "praise", "positive"),
    ("Destek ekibi hızlı döndü etkilendim", "praise", "positive"),
    ("Five stars well deserved", "praise", "positive"),
    ("Beş yıldızı hak ediyor", "praise", "positive"),
    ("Makes my daily workflow so much easier", "praise", "positive"),
    ("Günlük işlerimi çok kolaylaştırıyor", "praise", "positive"),
    ("Amazing update, love the new design", "praise", "positive"),
    ("Muhteşem güncelleme yeni tasarım harika", "praise", "positive"),
    ("Reliable and intuitive", "praise", "positive"),
    ("Güvenilir ve kullanımı kolay", "praise", "positive"),
    ("Worth every penny of the subscription", "praise", "positive"),
    ("Abonelik fiyatına değer", "praise", "positive"),
    ("My kids love it too", "praise", "positive"),
    ("Çocuklarım da bayılıyor", "praise", "positive"),
    ("Finally an app that just works", "praise", "positive"),
    ("Sonunda düzgün çalışan bir uygulama", "praise", "positive"),
    ("Great job developers keep it up", "praise", "positive"),
    ("Elinize sağlık devam edin", "praise", "positive"),
    ("Smooth animations and thoughtful UX", "praise", "positive"),
    ("Akıcı animasyonlar düşünülmüş UX", "praise", "positive"),
    ("Recommended to all my colleagues", "praise", "positive"),
    ("Tüm iş arkadaşlarıma önerdim", "praise", "positive"),
    ("Privacy focused and transparent, thank you", "praise", "positive"),
    ("Gizlilik odaklı ve şeffaf teşekkürler", "praise", "positive"),
    ("Offline mode works great on trips", "praise", "positive"),
    ("Offline mod seyahatte harika", "praise", "positive"),
    ("Simple but powerful exactly what I needed", "praise", "positive"),
    ("Basit ama güçlü tam ihtiyacım olan", "praise", "positive"),
    ("Updates are frequent and meaningful", "praise", "positive"),
    ("Güncellemeler sık ve anlamlı", "praise", "positive"),
    ("Beautiful typography and colors", "praise", "positive"),
    ("Tipografi ve renkler çok güzel", "praise", "positive"),
    ("Top tier app no complaints", "praise", "positive"),
    ("Üst düzey uygulama şikayet yok", "praise", "positive"),
    # --- spam (32) ---
    ("Download my crypto app now get rich!!!", "spam", "neutral"),
    ("Hemen kripto uygulamamı indir zengin ol!!!", "spam", "neutral"),
    ("FREE V-BUCKS click here http://spam.test", "spam", "neutral"),
    ("BEDAVA V-BUCKS tıkla http://spam.test", "spam", "neutral"),
    ("Follow me on instagram for hacks", "spam", "neutral"),
    ("Hile için instagramdan takip et", "spam", "neutral"),
    ("5 stars only because promo code promised", "spam", "neutral"),
    ("Promosyon kodu vaat ettiler diye 5 yıldız", "spam", "neutral"),
    ("Buy followers cheap DM me", "spam", "neutral"),
    ("Ucuz takipçi satın al DM at", "spam", "neutral"),
    ("This is an ad not a real review", "spam", "neutral"),
    ("Bu gerçek yorum değil reklam", "spam", "neutral"),
    ("Check out my YouTube channel link below", "spam", "neutral"),
    ("YouTube kanalıma abone olun link altta", "spam", "neutral"),
    ("Win iPhone click survey now", "spam", "neutral"),
    ("iPhone kazan anket için tıkla", "spam", "neutral"),
    ("MLM opportunity join my team", "spam", "neutral"),
    ("Network marketing fırsatı ekibime katıl", "spam", "neutral"),
    ("Fake review bot posting", "spam", "neutral"),
    ("Sahte yorum botu", "spam", "neutral"),
    ("Get mod apk unlimited money", "spam", "neutral"),
    ("Mod apk sınırsız para indir", "spam", "neutral"),
    ("Casino bonus use code WIN123", "spam", "neutral"),
    ("Casino bonusu WIN123 kodu kullan", "spam", "neutral"),
    ("SEO backlink service cheap rates", "spam", "neutral"),
    ("Ucuz backlink hizmeti", "spam", "neutral"),
    ("Plagiarized copy paste spam text", "spam", "neutral"),
    ("Kopyala yapıştır spam metin", "spam", "neutral"),
    ("Earn money from home no effort", "spam", "neutral"),
    ("Evden para kazan sıfır emek", "spam", "neutral"),
    ("Telegram group for leaked accounts", "spam", "neutral"),
    ("Sızdırılmış hesaplar telegram grubu", "spam", "neutral"),
    # --- other (48) — mixed sentiment ---
    ("This app is terrible", "other", "negative"),
    ("Kötü bir uygulama hiç beğenmedim", "other", "negative"),
    ("Overpriced for what it offers", "other", "negative"),
    ("Verdiği hizmete göre pahalı", "other", "negative"),
    ("Confusing onboarding took forever", "other", "negative"),
    ("Karışık tanıtım çok uzun sürdü", "other", "negative"),
    ("Meh nothing special", "other", "neutral"),
    ("Eh idare eder özel bir şey yok", "other", "neutral"),
    ("Ok app I guess", "other", "neutral"),
    ("Fena değil herhalde", "other", "neutral"),
    ("Too many permissions requested", "other", "negative"),
    ("Çok fazla izin istiyor", "other", "negative"),
    ("Subscription model is annoying", "other", "negative"),
    ("Abonelik modeli sinir bozucu", "other", "negative"),
    ("UI redesign made it worse honestly", "other", "negative"),
    ("Yeni arayüz daha kötü oldu", "other", "negative"),
    ("Average experience not bad not great", "other", "neutral"),
    ("Ortalama ne iyi ne kötü", "other", "neutral"),
    ("Used it once probably won't again", "other", "negative"),
    ("Bir kez kullandım bir daha kullanmam", "other", "negative"),
    ("Terms of service too aggressive", "other", "negative"),
    ("Kullanım şartları çok agresif", "other", "negative"),
    ("Decent but competitors are better", "other", "neutral"),
    ("Fena değil ama rakipler daha iyi", "other", "neutral"),
    ("Mixed feelings about privacy policy", "other", "neutral"),
    ("Gizlilik politikası konusunda kararsızım", "other", "neutral"),
    ("Too many ads in free version", "other", "negative"),
    ("Ücretsiz sürümde çok reklam var", "other", "negative"),
    ("Good concept poor execution overall", "other", "negative"),
    ("Fikir iyi uygulama zayıf", "other", "negative"),
    ("Not sure if I recommend it yet", "other", "neutral"),
    ("Henüz tavsiye eder miyim emin değilim", "other", "neutral"),
    ("Feels outdated compared to alternatives", "other", "negative"),
    ("Alternatiflere göre eski hissettiriyor", "other", "negative"),
    ("Fair price fair product", "other", "neutral"),
    ("Fiyatına göre idare eder", "other", "neutral"),
    ("Support is slow but app is fine", "other", "neutral"),
    ("Destek yavaş ama uygulama idare eder", "other", "neutral"),
    ("Wish it had better documentation", "other", "neutral"),
    ("Daha iyi dokümantasyon olmalı", "other", "neutral"),
    ("Company ethics concern me", "other", "negative"),
    ("Şirket etiği konusunda endişeliyim", "other", "negative"),
    ("It's alright for beginners", "other", "neutral"),
    ("Yeni başlayanlar için fena değil", "other", "neutral"),
    ("Deleted after a week", "other", "negative"),
    ("Bir hafta sonra sildim", "other", "negative"),
    ("Could be worse could be better", "other", "neutral"),
    ("Daha kötüsü de olabilir daha iyisi de", "other", "neutral"),
    # --- edge cases (24) — tricky boundaries ---
    ("Crash fixed in latest update thanks", "praise", "positive"),
    ("Son güncellemede çökme düzeldi teşekkürler", "praise", "positive"),
    ("Please fix crashing bug ASAP", "bug", "negative"),
    ("Çökme hatasını acil düzeltin", "bug", "negative"),
    ("Love the app but needs dark mode", "feature", "positive"),
    ("Uygulamayı seviyorum ama dark mode lazım", "feature", "positive"),
    ("Great app no bugs found", "praise", "positive"),
    ("Harika uygulama hata bulamadım", "praise", "positive"),
    ("Is this a scam app?", "other", "negative"),
    ("Bu uygulama dolandırıcılık mı?", "other", "negative"),
    ("1 star until you add widgets", "feature", "negative"),
    ("Widget gelene kadar 1 yıldız", "feature", "negative"),
    ("Works fine on wifi only issue on 5G", "bug", "negative"),
    ("Wifi'de iyi 5G'de sorun var", "bug", "negative"),
    ("Not a bug just confusing UX", "other", "neutral"),
    ("Hata değil sadece kafa karıştırıcı arayüz", "other", "neutral"),
    ("Spammy notifications every hour stop it", "bug", "negative"),
    ("Saat başı spam bildirim göndermeyin", "bug", "negative"),
    ("Best feature request forum ever jk this is store review", "other", "neutral"),
    ("Şaka bir yana mağaza yorumuyum", "other", "neutral"),
    ("Refund please app doesn't work", "bug", "negative"),
    ("Para iadesi uygulama çalışmıyor", "bug", "negative"),
    ("Five stars amazing!!! buy now limited offer", "spam", "neutral"),
    ("Beş yıldız muhteşem hemen al sınırlı kampanya", "spam", "neutral"),
]


def main() -> None:
    out = Path(__file__).parent / "train_v2.jsonl"
    rows = []
    for review, category, sentiment in EXAMPLES:
        user = PROMPT_PREFIX + repr(review)
        assistant = json.dumps({"category": category, "sentiment": sentiment}, separators=(",", ":"))
        rows.append({"messages": [{"role": "user", "content": user}, {"role": "assistant", "content": assistant}]})

    with out.open("w", encoding="utf-8") as f:
        for row in rows:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")

    from collections import Counter

    cats = Counter(c for _, c, _ in EXAMPLES)
    sents = Counter(s for _, _, s in EXAMPLES)
    print(f"Wrote {len(rows)} examples to {out}")
    print("Categories:", dict(cats))
    print("Sentiments:", dict(sents))


if __name__ == "__main__":
    main()
