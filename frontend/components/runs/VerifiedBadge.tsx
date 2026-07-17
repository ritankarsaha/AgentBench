type VerifiedBadgeProps = {
  verified: boolean;
  hasTrace: boolean;
};

export function VerifiedBadge({ verified, hasTrace }: VerifiedBadgeProps) {
  if (verified) {
    return (
      <span className="whitespace-nowrap rounded-badge border border-accent-verified/40 bg-accent-verified/10 px-2 py-0.5 text-xs font-medium text-accent-verified">
        ⚡ Verified
      </span>
    );
  }

  if (hasTrace) {
    return (
      <span className="whitespace-nowrap rounded-badge border border-border px-2 py-0.5 text-xs text-text-muted">
        ✓ Trace-Backed
      </span>
    );
  }

  return null;
}
