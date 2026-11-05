import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Queuewise | Appointment & Queue",
  description: "Appointment and queue management.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html lang="id">
      <body className="min-h-full flex flex-col">{children}</body>
    </html>
  );
}
