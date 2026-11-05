import { Suspense } from "react";
import { QueueApp } from "@/components/queuewise-app";
export default function QueuePage() { return <Suspense fallback={<main className="public-shell"><div className="panel">Memuat antrean…</div></main>}><QueueApp /></Suspense>; }
