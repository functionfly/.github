import React, { useEffect, useRef } from "react";
import { API_ORIGIN } from "../config";
import { cn } from "../lib/utils";

export default function RequestInviteModal() {
  const dialogRef = useRef<HTMLDialogElement>(null);

  useEffect(() => {
    const openBtn = document.getElementById("open-invite-modal");
    const dialog = dialogRef.current;
    if (!openBtn || !dialog) return;
    const handleOpen = () => dialog.showModal();
    openBtn.addEventListener("click", handleOpen);
    return () => openBtn.removeEventListener("click", handleOpen);
  }, []);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const form = e.currentTarget;
    const formData = new FormData(form);
    const email = formData.get("email") as string;
    if (!email) return;

    const submitBtn = form.querySelector(
      "button[type=submit]",
    ) as HTMLButtonElement;
    submitBtn.disabled = true;
    submitBtn.textContent = "Sending…";

    try {
      const res = await fetch(`${API_ORIGIN}/v1/waitlist`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, source: "signup_modal" }),
      });

      if (res.ok) {
        form.style.display = "none";
        const successEl = document.getElementById("modal-success");
        if (successEl) {
          successEl.style.display = "block";
        }
      } else {
        const data = await res.json().catch(() => ({}));
        showError(data.message || "Something went wrong. Please try again.");
        submitBtn.disabled = false;
        submitBtn.textContent = "Send request";
      }
    } catch {
      showError("Network error. Please check your connection and try again.");
      submitBtn.disabled = false;
      submitBtn.textContent = "Send request";
    }
  };

  const showError = (msg: string) => {
    const existing = document.getElementById("modal-error");
    if (existing) existing.remove();
    const errorEl = document.createElement("div");
    errorEl.id = "modal-error";
    errorEl.className = cn(
      "p-2.5 mb-3 rounded-lg text-sm",
      "bg-red-500/10 border border-red-500/25 text-[var(--ff-error)]"
    );
    errorEl.textContent = msg;
    const form = document.getElementById("modal-form");
    form?.prepend(errorEl);
  };

  return (
    <dialog
      ref={dialogRef}
      id="invite-modal"
      className="fixed inset-0 z-[1000] w-full h-full max-w-full max-h-full p-0 m-0 border-none bg-transparent items-center justify-center [&[open]]:flex [&:not([open])]:hidden backdrop-blur-sm"
      style={{ backgroundColor: "rgba(13, 17, 23, 0.9)" }}
      onClick={(e) => {
        const rect = dialogRef.current?.getBoundingClientRect();
        if (
          rect &&
          (e.clientX < rect.left ||
            e.clientX > rect.right ||
            e.clientY < rect.top ||
            e.clientY > rect.bottom)
        ) {
          dialogRef.current?.close();
        }
      }}
    >
      <div className="ff-modal__content max-w-[420px] p-8">
        <button
          type="button"
          className="absolute top-4 right-4 w-8 h-8 rounded-lg bg-transparent border-none text-[var(--ff-muted-text)] cursor-pointer flex items-center justify-center transition-colors hover:bg-[#1C2128] hover:text-[var(--ff-flame)]"
          onClick={() => dialogRef.current?.close()}
          aria-label="Close"
        >
          <svg
            width="18"
            height="18"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <line x1="18" y1="6" x2="6" y2="18" />
            <line x1="6" y1="6" x2="18" y2="18" />
          </svg>
        </button>

        <div
          id="modal-success"
          className="text-center"
          style={{ display: "none" }}
        >
          <div className="w-[72px] h-[72px] mx-auto mb-5 bg-[rgba(0,255,157,0.1)] rounded-full flex items-center justify-center text-[var(--ff-success)]">
            <svg
              width="40"
              height="40"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
              <polyline points="22 4 12 14.01 9 11.01" />
            </svg>
          </div>
          <h3 className="text-xl font-semibold text-[var(--ff-primary-text)] mb-2">
            You're on the list!
          </h3>
          <p className="text-sm text-[var(--ff-secondary-text)] mb-6 leading-relaxed">
            We'll send you an invite code as soon as we're ready for more
            users. Hang tight!
          </p>
          <button
            type="button"
            className="ff-btn ff-btn--primary w-full mt-2"
            onClick={() => dialogRef.current?.close()}
          >
            Got it
          </button>
        </div>

        <div className="text-center mb-6">
          <div className="w-[52px] h-[52px] mx-auto mb-4 bg-[rgba(255,107,53,0.12)] rounded-[14px] flex items-center justify-center text-[var(--ff-flame)]">
            <svg
              width="22"
              height="22"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
            >
              <path d="M21 2l-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0l3 3L22 7l-3-3m-3.5 3.5L19 4" />
            </svg>
          </div>
          <h2 className="text-lg font-semibold text-[var(--ff-primary-text)] mb-1">
            Request an invite
          </h2>
          <p className="text-sm text-[var(--ff-secondary-text)] leading-relaxed">
            Enter your email and we'll get you early access to FunctionFly™.
          </p>
        </div>

        <form 
          id="modal-form" 
          onSubmit={handleSubmit} 
          className="flex flex-col gap-4"
        >
          <div className="ff-field mb-0">
            <label className="ff-field__label" htmlFor="modal-email">
              Email address
            </label>
            <input
              id="modal-email"
              name="email"
              type="email"
              placeholder="you@example.com"
              required
              autoComplete="email"
              className="ff-input"
            />
          </div>
          <button type="submit" className="ff-btn ff-btn--primary w-full">
            Send request
          </button>
        </form>
      </div>
    </dialog>
  );
}
