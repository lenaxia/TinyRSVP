#!/usr/bin/env bash
# test_route_command.sh — Test suite for route-command.sh
#
# Validates the AI-command routing logic across:
#   - Command detection (12 commands, prefix and inline positions)
#   - NOTE extraction (command + --no-merge stripped)
#   - --no-merge flag handling (trailing holds, mid-position does not)
#   - HOLD_MERGE flag for /fix, /implement, /test, /security
#   - Prompt assembly (context + core-rules + command-specific file)
#
# Usage:
#   .github/scripts/test_route_command.sh
#
# Exits non-zero on any test failure.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROUTER="$SCRIPT_DIR/route-command.sh"
PROMPTS_DIR="$SCRIPT_DIR/../prompts"

# Per-test scratch directory so each run is isolated.
SCRATCH="$(mktemp -d)"
trap 'rm -rf "$SCRATCH"' EXIT

PASS=0
FAIL=0
FAILED_TESTS=()

# run_router <comment_body> [pr_url] [event_name]
# Sets globals: COMMAND, NOTE, HOLD_MERGE, OUT_FILE, OUT_CONTENT
run_router() {
    local comment="$1"
    local pr_url="${2:-}"
    local event_name="${3:-}"
    local out_file="$SCRATCH/prompt.txt"

    # Reset all globals so tests are independent.
    COMMAND=""
    NOTE=""
    HOLD_MERGE=""
    OUT_CONTENT=""

    # Execute the router in a clean subshell so its exports don't leak,
    # then re-parse the KEY=VALUE lines it prints.
    #
    # Output format is:
    #   COMMAND=<value>\nNOTE=<value, possibly multi-line>\nHOLD_MERGE=<value>\nOUT_FILE=<value>
    #
    # NOTE is the only field that may itself contain newlines, so we
    # extract each fixed-name field and treat everything between NOTE= and
    # the trailing \nHOLD_MERGE= as the note body.
    local output
    output="$(env -i \
        COMMENT_BODY="$comment" \
        PR_URL="$pr_url" \
        EVENT_NAME="$event_name" \
        OUT_FILE="$out_file" \
        PROMPTS_DIR="$PROMPTS_DIR" \
        bash "$ROUTER" 2>/dev/null)"

    COMMAND="$(printf '%s' "$output" | sed -n 's/^COMMAND=//p')"
    HOLD_MERGE="$(printf '%s' "$output" | sed -n 's/^HOLD_MERGE=//p')"
    NOTE="$(printf '%s' "$output" | sed -n '/^NOTE=/{ s/^NOTE=//; p; n; :loop; /^HOLD_MERGE=/q; p; n; b loop; }')"
    # Trim a trailing newline that the loop-based sed leaves behind.
    NOTE="${NOTE%$'\n'}"

    OUT_CONTENT=""
    if [ -f "$out_file" ]; then
        OUT_CONTENT="$(cat "$out_file")"
    fi
}

# assert_eq <description> <actual> <expected>
assert_eq() {
    local desc="$1" actual="$2" expected="$3"
    if [ "$actual" = "$expected" ]; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        FAILED_TESTS+=("$desc")
        printf '  FAIL: %s\n    expected: <%s>\n    actual:   <%s>\n' \
            "$desc" "$expected" "$actual" >&2
    fi
}

# assert_contains <description> <haystack> <needle>
assert_contains() {
    local desc="$1" haystack="$2" needle="$3"
    if printf '%s' "$haystack" | grep -Fq -- "$needle"; then
        PASS=$((PASS + 1))
    else
        FAIL=$((FAIL + 1))
        FAILED_TESTS+=("$desc")
        printf '  FAIL: %s\n    expected to contain: <%s>\n' \
            "$desc" "$needle" >&2
    fi
}

# assert_not_contains <description> <haystack> <needle>
assert_not_contains() {
    local desc="$1" haystack="$2" needle="$3"
    if printf '%s' "$haystack" | grep -Fq -- "$needle"; then
        FAIL=$((FAIL + 1))
        FAILED_TESTS+=("$desc")
        printf '  FAIL: %s\n    should NOT contain: <%s>\n' \
            "$desc" "$needle" >&2
    else
        PASS=$((PASS + 1))
    fi
}

echo "=== route-command.sh test suite ==="
echo ""

# ---------------------------------------------------------------------------
# 1. Command detection: prefix position (command at start of comment)
# ---------------------------------------------------------------------------
echo "-- 1. Prefix command detection --"

for cmd in /implement /security /analyze /review /explain /triage /design /merge /test /fix /help /ai; do
    run_router "$cmd some detail"
    assert_eq "prefix '$cmd' resolves to $cmd" "$COMMAND" "$cmd"
done

# Bare command with no arguments
for cmd in /implement /security /analyze /review /explain /triage /design /merge /test /fix /help /ai; do
    run_router "$cmd"
    assert_eq "bare '$cmd' (no args) resolves to $cmd" "$COMMAND" "$cmd"
done

# /test must not match /testing, /fix must not match /fixing, etc.
run_router "/testing something"
assert_eq "/testing is NOT /test" "$COMMAND" "/ai"
run_router "/fixing a bug"
assert_eq "/fixing is NOT /fix" "$COMMAND" "/ai"
run_router "/implementing a thing"
assert_eq "/implementing is NOT /implement" "$COMMAND" "/ai"

# ---------------------------------------------------------------------------
# 2. Command detection: inline position (command mid-comment)
# ---------------------------------------------------------------------------
echo "-- 2. Inline command detection --"

for cmd in /implement /security /analyze /review /explain /triage /design /merge /test /fix /help /ai; do
    run_router "Hey team $cmd and the rest"
    assert_eq "inline '$cmd' resolves to $cmd" "$COMMAND" "$cmd"
done

# Inline without surrounding spaces must NOT match (prevents false positives
# like "the/test/path" being treated as /test).
run_router "see https://example.com/review/policy"
assert_eq "URL path '/review/' is NOT /review" "$COMMAND" "/ai"

# ---------------------------------------------------------------------------
# 3. NOTE extraction
# ---------------------------------------------------------------------------
echo "-- 3. NOTE extraction --"

run_router "/fix the login bug"
assert_eq "NOTE strips command prefix" "$NOTE" "the login bug"

run_router "/explain the auth flow"
assert_eq "NOTE strips /explain prefix" "$NOTE" "the auth flow"

run_router "/ai What is the data model?"
assert_eq "NOTE preserves question marks" "$NOTE" "What is the data model?"

run_router "/test internal/handlers/"
assert_eq "NOTE preserves slashes" "$NOTE" "internal/handlers/"

run_router "/review  extra   spaces  "
assert_eq "NOTE trims surrounding whitespace" "$NOTE" "extra   spaces"

# Inline command: NOTE is everything after the command token.
run_router "please /explain the session layer"
assert_eq "NOTE from inline command" "$NOTE" "the session layer"

# Empty NOTE when command has no text.
run_router "/fix"
assert_eq "empty NOTE for bare /fix" "$NOTE" ""

# ---------------------------------------------------------------------------
# 4. --no-merge flag handling
# ---------------------------------------------------------------------------
echo "-- 4. --no-merge flag handling --"

# Trailing --no-merge on the four code-change commands sets HOLD_MERGE=1.
for cmd in /fix /implement /test /security; do
    run_router "$cmd do the thing --no-merge"
    assert_eq "trailing --no-merge on $cmd sets HOLD_MERGE=1" "$HOLD_MERGE" "1"
    assert_eq "--no-merge stripped from NOTE on $cmd" "$NOTE" "do the thing"
done

# Trailing --no-merge on non-code-change commands is stripped from NOTE
# but does NOT set HOLD_MERGE.
for cmd in /analyze /review /explain /triage /ai /help /merge; do
    run_router "$cmd something --no-merge"
    assert_eq "trailing --no-merge on $cmd keeps HOLD_MERGE=0" "$HOLD_MERGE" "0"
    assert_eq "--no-merge stripped from NOTE on $cmd" "$NOTE" "something"
done

# /design always holds via its own prompt, --no-merge does not add a hold.
run_router "/design the new API --no-merge"
assert_eq "/design with --no-merge keeps HOLD_MERGE=0 (holds via prompt)" "$HOLD_MERGE" "0"

# Mid-sentence --no-merge must NOT be treated as the flag.
run_router "/fix use the --no-merge flag carefully"
assert_eq "mid-position --no-merge does NOT set HOLD_MERGE on /fix" "$HOLD_MERGE" "0"
assert_eq "mid-position --no-merge stays in NOTE" "$NOTE" "use the --no-merge flag carefully"

run_router "/implement handle --no-merge option"
assert_eq "mid-position --no-merge keeps HOLD_MERGE=0 on /implement" "$HOLD_MERGE" "0"

# --no-merge as the entire comment body after a bare command.
run_router "/test --no-merge"
assert_eq "bare /test --no-merge sets HOLD_MERGE=1" "$HOLD_MERGE" "1"
assert_eq "bare /test --no-merge NOTE is empty" "$NOTE" ""

# ---------------------------------------------------------------------------
# 5. Prompt assembly — context + core-rules always present
# ---------------------------------------------------------------------------
echo "-- 5. Prompt assembly --"

run_router "/fix a bug"
assert_contains "prompt always includes context.md header" "$OUT_CONTENT" "TinyRSVP"
assert_contains "prompt always includes core-rules.md" "$OUT_CONTENT" "Core Rules"
assert_contains "prompt includes code-change-workflow.md for /fix" "$OUT_CONTENT" "Code Change Workflow"
assert_contains "prompt includes fix.md for /fix" "$OUT_CONTENT" "You are fixing a bug"
assert_contains "prompt includes the bug note" "$OUT_CONTENT" "a bug"

run_router "/review focus on error paths"
assert_contains "prompt includes pr-review.md for /review" "$OUT_CONTENT" "code reviewer"
assert_contains "prompt includes review focus note" "$OUT_CONTENT" "focus on error paths"
assert_not_contains "/review prompt does NOT include code-change-workflow.md" "$OUT_CONTENT" "Code Change Workflow"

run_router "/design the new module"
assert_contains "prompt includes design.md for /design" "$OUT_CONTENT" "design document"
assert_contains "prompt includes code-change-workflow.md for /design" "$OUT_CONTENT" "Code Change Workflow"

run_router "/merge"
assert_contains "prompt includes merge.md for /merge" "$OUT_CONTENT" "finalizing an already-reviewed"
assert_not_contains "/merge prompt does NOT include code-change-workflow.md" "$OUT_CONTENT" "Code Change Workflow"

run_router "/help"
assert_contains "prompt includes help.md for /help" "$OUT_CONTENT" "Post a comment on the issue or PR"

# /ai on an issue with no note -> issue responder prompt.
run_router "/ai"
assert_contains "/ai on issue (no note) -> issue-responder.md" "$OUT_CONTENT" "triggered you on a GitHub issue"
assert_not_contains "/ai on issue does NOT include code-change-workflow.md" "$OUT_CONTENT" "Code Change Workflow"

# /ai on a PR with no note -> PR review prompt.
run_router "/ai" "https://github.com/org/repo/pull/1" ""
assert_contains "/ai on PR (no note) -> pr-review.md" "$OUT_CONTENT" "code reviewer"

# /ai with a note -> note prompt regardless of issue/PR.
run_router "/ai explain the data layer"
assert_contains "/ai with note surfaces the note" "$OUT_CONTENT" "explain the data layer"

# HOLD_MERGE directive appears in the assembled prompt when set.
run_router "/fix a bug --no-merge"
assert_contains "--no-merge adds MERGE HOLD directive to prompt" "$OUT_CONTENT" "MERGE HOLD"

run_router "/fix a bug"
assert_not_contains "no MERGE HOLD when --no-merge absent" "$OUT_CONTENT" "MERGE HOLD"

# ---------------------------------------------------------------------------
# 6. Edge cases
# ---------------------------------------------------------------------------
echo "-- 6. Edge cases --"

# Comment with no recognizable command defaults to /ai.
run_router "just wondering how the invite tokens work"
assert_eq "unrecognized text defaults to /ai" "$COMMAND" "/ai"
assert_eq "NOTE is the full body when defaulting to /ai" "$NOTE" "just wondering how the invite tokens work"

# Empty comment body.
run_router ""
assert_eq "empty body defaults to /ai" "$COMMAND" "/ai"
assert_eq "empty body NOTE is empty" "$NOTE" ""

# Unicode in NOTE survives intact.
run_router "/fix café résumé naïve"
assert_eq "unicode preserved in NOTE" "$NOTE" "café résumé naïve"

# ---------------------------------------------------------------------------
# Summary
# ---------------------------------------------------------------------------
echo ""
echo "=== Results ==="
printf '  passed: %d\n  failed: %d\n' "$PASS" "$FAIL"

if [ "$FAIL" -gt 0 ]; then
    echo ""
    echo "FAILED tests:"
    for t in "${FAILED_TESTS[@]}"; do
        printf '  - %s\n' "$t"
    done
    exit 1
fi

echo "All tests passed."
