import { Suspense } from "react";
import { BookingDetailApp } from "@/components/queuewise-app";
export default function BookingDetailPage() { return <Suspense fallback={<main className="public-shell"><div className="panel">Memuat booking…</div></main>}><BookingDetailApp /></Suspense>; }
