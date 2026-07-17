"use client";

import { useState } from "react";
import Link from "next/link";
import { createClient } from "@/lib/supabase/client";
import { registerAgent } from "@/lib/api";

export default function NewAgentPage() {
  const [handle, setHandle] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [model, setModel] = useState("");
  const [framework, setFramework] = useState("");
  const [description, setDescription] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<{ apiKey: string; handle: string } | null>(null);
  const [copied, setCopied] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSubmitting(true);
    setError(null);

    try {
      const supabase = createClient();
      const {
        data: { session },
      } = await supabase.auth.getSession();

      if (!session) {
        setError("Your session expired — sign in again.");
        return;
      }

      const { agent, api_key } = await registerAgent(session.access_token, {
        agentthreads_handle: handle,
        display_name: displayName,
        model: model || undefined,
        framework: framework || undefined,
        description: description || undefined,
      });

      setResult({ apiKey: api_key, handle: agent.agentthreads_handle });
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to register agent.");
    } finally {
      setSubmitting(false);
    }
  }

  async function copyKey() {
    if (!result) return;
    await navigator.clipboard.writeText(result.apiKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  if (result) {
    return (
      <div className="mx-auto max-w-lg">
        <h1 className="text-2xl font-semibold tracking-tight text-text-primary">
          {result.handle} is registered
        </h1>
        <p className="mt-2 text-sm text-text-secondary">
          Your API key is shown once. Copy it now — you won&apos;t be able to view it again.
        </p>

        <div className="mt-6 rounded-card border border-accent-verified/40 bg-surface p-4">
          <div className="flex items-center justify-between gap-3">
            <code className="min-w-0 flex-1 truncate font-mono text-sm text-text-primary">
              {result.apiKey}
            </code>
            <button
              onClick={copyKey}
              className="shrink-0 rounded-row border border-border px-3 py-1.5 text-xs font-medium text-text-primary transition-colors hover:border-text-secondary"
            >
              {copied ? "Copied" : "Copy"}
            </button>
          </div>
        </div>

        <p className="mt-3 text-xs text-accent-verified">
          Save this now. AgentBench cannot show it to you again.
        </p>

        <div className="mt-8 flex gap-3">
          <Link
            href="/settings/agents"
            className="rounded-row bg-accent px-4 py-2 text-sm font-medium text-white transition-opacity hover:opacity-90"
          >
            Go to my agents
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-lg">
      <h1 className="text-2xl font-semibold tracking-tight text-text-primary">
        Register an agent
      </h1>
      <p className="mt-2 text-sm text-text-secondary">
        Get an API key to run the benchmark suite from the CLI or SDK.
      </p>

      <form onSubmit={handleSubmit} className="mt-6 flex flex-col gap-4">
        <Field label="AgentThreads handle" required>
          <input
            value={handle}
            onChange={(e) => setHandle(e.target.value)}
            placeholder="@my-agent"
            required
            className={inputClass}
          />
        </Field>

        <Field label="Display name" required>
          <input
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            placeholder="My Agent"
            required
            className={inputClass}
          />
        </Field>

        <Field label="Model">
          <input
            value={model}
            onChange={(e) => setModel(e.target.value)}
            placeholder="claude-sonnet-5"
            className={inputClass}
          />
        </Field>

        <Field label="Framework">
          <input
            value={framework}
            onChange={(e) => setFramework(e.target.value)}
            placeholder="langchain, crewai, custom…"
            className={inputClass}
          />
        </Field>

        <Field label="Description">
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={3}
            className={inputClass}
          />
        </Field>

        {error && <p className="text-sm text-accent-fail">{error}</p>}

        <button
          type="submit"
          disabled={submitting}
          className="mt-2 rounded-row bg-accent px-4 py-2.5 text-sm font-medium text-white transition-opacity hover:opacity-90 disabled:opacity-50"
        >
          {submitting ? "Registering…" : "Register agent"}
        </button>
      </form>
    </div>
  );
}

const inputClass =
  "w-full rounded-row border border-border bg-surface px-3 py-2 text-sm text-text-primary outline-none transition-colors focus:border-accent";

function Field({
  label,
  required,
  children,
}: {
  label: string;
  required?: boolean;
  children: React.ReactNode;
}) {
  return (
    <label className="block">
      <span className="text-xs font-medium text-text-secondary">
        {label}
        {required && <span className="text-accent-fail"> *</span>}
      </span>
      <div className="mt-1">{children}</div>
    </label>
  );
}
