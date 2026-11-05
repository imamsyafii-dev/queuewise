import Link from "next/link";

export default function Home() {
  return <main className="landing"><nav className="topbar container"><Link className="brand" href="/">queue<span>wise</span></Link><Link className="text-link" href="/admin">Portal staff →</Link></nav><section className="hero container"><div><p className="eyebrow">APPOINTMENT & QUEUE, SIMPLIFIED</p><h1>Waktu pelanggan Anda pantas dikelola dengan baik.</h1><p className="lead">Booking yang sederhana untuk pelanggan. Antrean yang tenang untuk tim Anda.</p><div className="button-row"><Link className="button primary" href="/booking">Buat appointment</Link><Link className="button secondary" href="/queue/jakarta-selatan">Lihat antrean</Link></div></div><div className="hero-card"><p>Sedang dipanggil</p><strong>A-12</strong><span>Konsultasi Umum · Ruang 02</span><hr /><small>Berikutnya&nbsp; A-13 · A-14 · A-15</small></div></section></main>;
}
