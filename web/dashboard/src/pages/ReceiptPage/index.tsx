// ReceiptPage — the public, shareable landing page for an execution.
//
// Mounted at:
//   /r/:execId          (canonical)
//   /receipt/:execId    (alias for crawlers / future SEO)
//
// Layout: 2-column on desktop (md:grid-cols-3) with the function header
// spanning both columns. Single column on mobile.
import { ArrowLeft, ServerCrash } from "lucide-react";
import { useEffect, useState } from "react";
import { Helmet } from "react-helmet-async";
import { Link, useParams } from "react-router-dom";

import { Navbar } from "@/components/common/Navbar";
import { ErrorMessage } from "@/components/common/ErrorMessage";
import { LoadingSpinner } from "@/components/common/LoadingSpinner";
import { Button } from "@/components/ui/button";
import { API_URLS } from "@/lib/api-urls";
import { useAuthStore } from "@/stores/authStore";
import { Footer } from "@/pages/LandingPage/components";

import { ReceiptForkCTA } from "./components/ReceiptForkCTA";
import { ReceiptHeader } from "./components/ReceiptHeader";
import { ReceiptInputOutput } from "./components/ReceiptInputOutput";
import { ReceiptPoweredBy } from "./components/ReceiptPoweredBy";
import { ReceiptRunPanel } from "./components/ReceiptRunPanel";
import { ReceiptSchemaViewer } from "./components/ReceiptSchemaViewer";
import { ReceiptShareBar } from "./components/ReceiptShareBar";
import { ReceiptSkeleton } from "./components/ReceiptSkeleton";
import { ReceiptStats } from "./components/ReceiptStats";
import { useReceipt } from "./hooks/useReceipt";

interface ReceiptPageProps {
  /** When true, the page is embedded inside an iframe. Hides the navbar,
   *  footer, and the "Powered by" badge. The embed variant is mounted
   *  at /r/:execId/embed. */
  embed?: boolean;
}

function useIsAuthenticated(): boolean {
  // useAuthStore.getState() reads the current value without subscribing.
  // If the store hasn't been mounted yet (rare in the dashboard), this
  // returns the initial state which has isAuthenticated = false, so
  // the run panel will show the sign-in gate — never a crash.
  return useAuthStore.getState().isAuthenticated === true;
}

function fireViewPing(id: string) {
  // Best-effort. Don't await — never block the page render.
  try {
    fetch(API_URLS.receipt.view(id), { method: "POST", credentials: "omit", keepalive: true });
  } catch {
    // ignore
  }
}

export function ReceiptPage({ embed = false }: ReceiptPageProps) {
  const { execId } = useParams<{ execId: string }>();
  const receiptQ = useReceipt(execId);
  const isAuthenticated = useIsAuthenticated();
  const [hasFiredView, setHasFiredView] = useState(false);

  // Fire the view counter exactly once per mount. Keep the body small.
  useEffect(() => {
    if (!execId || hasFiredView) return;
    fireViewPing(execId);
    setHasFiredView(true);
  }, [execId, hasFiredView]);

  // ── Loading state ─────────────────────────────────────────────────────
  if (receiptQ.isLoading) {
    if (embed) {
      return <div className="p-4"><ReceiptSkeleton /></div>;
    }
    return (
      <div className="flex min-h-screen flex-col bg-bg-primary">
        {!embed && <Navbar variant="landing" />}
        <main className="flex-1 pt-16">
          <div className="container mx-auto max-w-5xl px-4 py-8">
            <ReceiptSkeleton />
          </div>
        </main>
        {!embed && <Footer />}
      </div>
    );
  }

  // ── Error / not-found / revoked state ─────────────────────────────────
  if (receiptQ.isError) {
    const err = receiptQ.error as Error & { status?: number };
    const status = err?.status;
    const message =
      status === 404
        ? "This receipt doesn't exist or has been removed."
        : status === 410
          ? "This receipt was revoked by its owner."
          : err?.message ?? "Failed to load the receipt.";
    if (embed) {
      return <div className="p-4 text-sm text-muted-foreground">{message}</div>;
    }
    return (
      <div className="flex min-h-screen flex-col bg-bg-primary">
        <Navbar variant="landing" />
        <main className="flex-1 pt-16">
          <div className="container mx-auto max-w-3xl px-4 py-12">
            <div className="rounded-lg border border-border/40 bg-card p-8 text-center">
              <ServerCrash className="mx-auto mb-4 h-10 w-10 text-muted-foreground" aria-hidden />
              <h1 className="text-xl font-semibold">Receipt unavailable</h1>
              <p className="mt-2 text-sm text-muted-foreground">{message}</p>
              <Button asChild variant="outline" className="mt-6">
                <Link to="/registry">
                  <ArrowLeft className="mr-2 h-4 w-4" aria-hidden /> Browse the registry
                </Link>
              </Button>
            </div>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  const receipt = receiptQ.data;
  if (!receipt) {
    return <ErrorMessage error={new Error("No receipt data")} />;
  }

  // ── Loaded state ──────────────────────────────────────────────────────
  return (
    <>
      {!embed ? (
        <Helmet>
          <title>{receipt.share.og_meta.title}</title>
          <meta name="description" content={receipt.share.og_meta.description} />
          {/* Open Graph */}
          <meta property="og:type" content="website" />
          <meta property="og:title" content={receipt.share.og_meta.title} />
          <meta property="og:description" content={receipt.share.og_meta.description} />
          <meta property="og:url" content={receipt.share.url} />
          <meta property="og:image" content={receipt.share.og_meta.image} />
          {/* Twitter Card */}
          <meta name="twitter:card" content="summary_large_image" />
          <meta name="twitter:title" content={receipt.share.og_meta.title} />
          <meta name="twitter:description" content={receipt.share.og_meta.description} />
          <meta name="twitter:image" content={receipt.share.og_meta.image} />
          {/* Embeddable for Slack/Discord unfurls */}
          <link rel="canonical" href={receipt.share.url} />
        </Helmet>
      ) : null}

      <div className={`flex min-h-screen flex-col bg-bg-primary ${embed ? "" : ""}`}>
        {!embed && <Navbar variant="landing" />}
        <main className={`flex-1 ${embed ? "" : "pt-16"}`}>
          <div
            className={
              embed
                ? "space-y-4 p-3"
                : "container mx-auto max-w-5xl space-y-6 px-4 py-8"
            }
          >
            {!embed && <ReceiptHeader receipt={receipt} />}
            {!embed && <ReceiptStats
              durationMs={receipt.execution.duration_ms}
              cached={receipt.execution.cached}
              createdAt={receipt.execution.created_at}
            />}

            <div className={embed ? "space-y-4" : "grid grid-cols-1 gap-6 md:grid-cols-3"}>
              <div className={embed ? "" : "md:col-span-2 space-y-6"}>
                <ReceiptInputOutput
                  input={receipt.execution.input}
                  output={receipt.execution.output}
                />
                {!embed && <ReceiptSchemaViewer receipt={receipt} />}
                {!embed && <ReceiptRunPanel receipt={receipt} isAuthenticated={isAuthenticated} />}
              </div>
              <div className={embed ? "" : "space-y-6"}>
                {!embed && <ReceiptShareBar receipt={receipt} />}
                {!embed && <ReceiptForkCTA receipt={receipt} isAuthenticated={isAuthenticated} />}
              </div>
            </div>
          </div>
        </main>
        {!embed && <Footer />}
        {!embed && <ReceiptPoweredBy />}
      </div>
    </>
  );
}

/**
 * ReceiptEmbedPage — minimal, iframe-safe variant of the receipt page.
 * No navbar/footer/powered-by. Mounted at /r/:execId/embed.
 */
export function ReceiptEmbedPage() {
  return <ReceiptPage embed />;
}
