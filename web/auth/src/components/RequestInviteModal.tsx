import React, { useEffect, useRef } from "react";
import { API_ORIGIN } from "../config";

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
    errorEl.className = "modal-error";
    errorEl.textContent = msg;
    const form = document.getElementById("modal-form");
    form?.prepend(errorEl);
  };

  return (
    <>
      <style>{`
        .invite-modal {
          position: fixed;
          inset: 0;
          z-index: 1000;
          width: 100%;
          height: 100%;
          max-width: 100%;
          max-height: 100%;
          padding: 0;
          margin: 0;
          border: none;
          background: transparent;
          display: none;
          align-items: center;
          justify-content: center;
        }
        .invite-modal[open] { display: flex; }
        .invite-modal::backdrop {
          background: rgba(0, 0, 0, 0.75);
          backdrop-filter: blur(4px);
        }
        .modal-inner {
          position: relative;
          background: #18181b;
          border: 1px solid #27272a;
          border-radius: 16px;
          padding: 2rem;
          width: 100%;
          max-width: 420px;
          box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
          animation: modalIn 0.2s ease;
        }
        @keyframes modalIn { from { opacity: 0; transform: scale(0.95) translateY(8px); } to { opacity: 1; transform: scale(1) translateY(0); } }
        .modal-close {
          position: absolute; top: 1rem; right: 1rem;
          width: 32px; height: 32px;
          border-radius: 8px;
          background: transparent;
          border: none;
          color: #71717a;
          cursor: pointer;
          display: flex; align-items: center; justify-content: center;
          transition: background 0.15s, color 0.15s;
        }
        .modal-close:hover { background: #27272a; color: #e4e4e7; }
        .modal-header { text-align: center; margin-bottom: 1.5rem; }
        .modal-icon {
          width: 52px; height: 52px;
          margin: 0 auto 1rem;
          background: rgba(99,102,241,0.12);
          border-radius: 14px;
          display: flex; align-items: center; justify-content: center;
          color: #818cf8;
        }
        .modal-header h2 {
          font-size: 1.25rem; font-weight: 600;
          color: #fafafa;
          margin: 0 0 0.5rem;
        }
        .modal-header p {
          font-size: 0.875rem; color: #a1a1aa;
          margin: 0; line-height: 1.5;
        }
        .modal-form { display: flex; flex-direction: column; gap: 1rem; }
        .modal-field { display: flex; flex-direction: column; gap: 0.375rem; }
        .modal-field label {
          font-size: 0.875rem; font-weight: 500; color: #e4e4e7;
        }
        .modal-field input {
          padding: 0.625rem 0.875rem;
          background: #09090b;
          border: 1px solid #27272a;
          border-radius: 8px;
          color: #fafafa;
          font-size: 0.9375rem;
          font-family: inherit;
          transition: border-color 0.15s, box-shadow 0.15s;
          outline: none;
        }
        .modal-field input:focus {
          border-color: #6366f1;
          box-shadow: 0 0 0 3px rgba(99,102,241,0.15);
        }
        .modal-field input::placeholder { color: #52525b; }
        .modal-error {
          padding: 0.625rem 0.875rem;
          background: rgba(239,68,68,0.1);
          border: 1px solid rgba(239,68,68,0.25);
          border-radius: 8px;
          color: #fca5a5;
          font-size: 0.8125rem;
        }
        .modal-success { text-align: center; }
        .modal-success .success-icon {
          width: 72px; height: 72px;
          margin: 0 auto 1.25rem;
          background: rgba(34,197,94,0.1);
          border-radius: 50%;
          display: flex; align-items: center; justify-content: center;
          color: #22c55e;
        }
        .modal-success h3 {
          font-size: 1.25rem; font-weight: 600; color: #fafafa;
          margin: 0 0 0.5rem;
        }
        .modal-success p {
          font-size: 0.875rem; color: #a1a1aa;
          margin: 0 0 1.5rem; line-height: 1.5;
        }
        .modal-success .btn { margin-top: 0.5rem; }
        .btn {
          display: inline-flex; align-items: center; justify-content: center; gap: 0.5rem;
          padding: 0.625rem 1.25rem; border-radius: 8px;
          font-size: 0.9375rem; font-weight: 500;
          cursor: pointer; transition: background 0.15s, opacity 0.15s;
          border: 1px solid transparent;
          font-family: inherit;
        }
        .btn:disabled { opacity: 0.5; cursor: not-allowed; }
        .btn-primary { background: #6366f1; color: #fff; }
        .btn-primary:hover:not(:disabled) { background: #4f46e5; }
        .btn-full { width: 100%; }
      `}</style>
      <dialog
        ref={dialogRef}
        id="invite-modal"
        className="invite-modal"
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
        <div className="modal-inner">
          <button
            type="button"
            className="modal-close"
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
            className="modal-success"
            style={{ display: "none" }}
          >
            <div className="success-icon">
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
            <h3>You're on the list!</h3>
            <p>
              We'll send you an invite code as soon as we're ready for more
              users. Hang tight!
            </p>
            <button
              type="button"
              className="btn btn-primary btn-full"
              onClick={() => dialogRef.current?.close()}
            >
              Got it
            </button>
          </div>

          <div className="modal-header">
            <div className="modal-icon">
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
            <h2>Request an invite</h2>
            <p>
              Enter your email and we'll get you early access to FunctionFly™.
            </p>
          </div>

          <form id="modal-form" onSubmit={handleSubmit} className="modal-form">
            <div className="modal-field">
              <label htmlFor="modal-email">Email address</label>
              <input
                id="modal-email"
                name="email"
                type="email"
                placeholder="you@example.com"
                required
                autoComplete="email"
              />
            </div>
            <button type="submit" className="btn btn-primary btn-full">
              Send request
            </button>
          </form>
        </div>
      </dialog>
    </>
  );
}
