export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type Catalog = { branches: { id: string; name: string; slug: string; address: string }[]; services: { id: string; name: string; description: string; duration_minutes: number; price_amount: number }[]; staff: { id: string; name: string; service_id: string }[] };
export type Appointment = { id: string; booking_code: string; branch_name: string; branch_slug: string; customer_name: string; service_name: string; staff_name: string; scheduled_start_at: string; status: string; queue_id?: string; queue_number?: number };
export type Queue = { branch_name: string; current?: { id: string; queue_number: number; status: string; customer_name: string; service_name: string }; waiting: { id: string; queue_number: number; customer_name: string; service_name: string }[] };

export async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, { headers: { "Content-Type": "application/json" }, ...options });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) throw new Error(payload.error ?? "Terjadi kesalahan.");
  return payload.data as T;
}

export const money = (value: number) => new Intl.NumberFormat("id-ID", { style: "currency", currency: "IDR", maximumFractionDigits: 0 }).format(value);
export const dateTime = (value: string) => new Intl.DateTimeFormat("id-ID", { dateStyle: "medium", timeStyle: "short" }).format(new Date(value));
