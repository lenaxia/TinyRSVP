#!/bin/bash
#
# Pre-Commit Hook Setup Script
# 
# This script sets up pre-commit hooks for the TinyRSVP project.
# The hooks enforce code quality standards before allowing commits.
#

set -e

HOOK_DIR=".git/hooks"
HOOK_FILE="${HOOK_DIR}/pre-commit"

echo "Setting up TinyRSVP pre-commit hooks..."
echo ""

# Check if .git directory exists
if [ ! -d ".git" ]; then
    echo "ERROR: This script must be run from the repository root"
    echo "Current directory: $(pwd)"
    exit 1
fi

# Create hook
cat > "${HOOK_FILE}" << 'EOF'
#!/bin/bash

set -e

echo "🔍 Running pre-commit checks..."
echo ""

# Step 1: Format code
echo "📝 Formatting code..."
go fmt ./...
if [ $? -ne 0 ]; then
    echo "❌ go fmt failed"
    exit 1
fi
echo "✅ Code formatted"
echo ""

# Step 2: Run go vet
echo "🔎 Running static analysis..."
go vet ./...
if [ $? -ne 0 ]; then
    echo "❌ go vet failed"
    exit 1
fi
echo "✅ Static analysis passed"
echo ""

# Step 3: Run tests (MANDATORY: use -timeout 30s per README-LLM.md)
if [ "$SKIP_TESTS" != "1" ]; then
    echo "🧪 Running tests with 30s timeout..."
    go test -timeout 30s ./...
    if [ $? -ne 0 ]; then
        echo "❌ Tests failed"
        echo ""
        echo "Fix the failing tests before committing."
        echo "To skip tests in emergency: SKIP_TESTS=1 git commit -m \"message\""
        exit 1
    fi
    echo "✅ All tests passed"
    echo ""
else
    echo "⚠️  Skipping tests (SKIP_TESTS=1 set)"
    echo ""
fi

# Step 4: Check for common issues
echo "🔍 Checking for common issues..."

# Check for debugging prints (excluding test files)
DEBUG_PRINTS=$(git diff --cached --name-only --diff-filter=ACM | grep "\.go$" | grep -v "_test\.go$" | xargs grep -Hn "fmt\.Println\|log\.Println" 2>/dev/null || true)
if [ -n "$DEBUG_PRINTS" ]; then
    echo "⚠️  Warning: Found fmt.Println or log.Println in staged files:"
    echo "$DEBUG_PRINTS"
    echo "Consider using proper logging instead"
    echo ""
fi

# Check for TODO/FIXME/HACK
TODO_ITEMS=$(git diff --cached --name-only --diff-filter=ACM | grep "\.go$" | xargs grep -Hn "TODO\|FIXME\|HACK" 2>/dev/null || true)
if [ -n "$TODO_ITEMS" ]; then
    echo "⚠️  Warning: Found TODO/FIXME/HACK comments in staged files:"
    echo "$TODO_ITEMS"
    echo "Consider addressing these before committing"
    echo ""
fi

# Check for map[string]interface{} usage (anti-pattern per README-LLM.md)
UNTYPED_MAPS=$(git diff --cached --name-only --diff-filter=ACM | grep "\.go$" | grep -v "_test\.go$" | xargs grep -Hn "map\[string\]interface{}" 2>/dev/null || true)
if [ -n "$UNTYPED_MAPS" ]; then
    echo "⚠️  Warning: Found map[string]interface{} in staged files:"
    echo "$UNTYPED_MAPS"
    echo "Per README-LLM.md: Use strongly-typed structs instead"
    echo ""
fi

echo "✅ Pre-commit checks complete!"
echo ""
EOF

# Make hook executable
chmod +x "${HOOK_FILE}"

echo "✅ Pre-commit hook installed at: ${HOOK_FILE}"
echo ""
echo "The hook will run automatically on every commit and will:"
echo "  1. Format code with go fmt"
echo "  2. Run static analysis with go vet"
echo "  3. Run all tests with -timeout 30s"
echo "  4. Warn about debug prints, TODOs, and map[string]interface{}"
echo ""
echo "To bypass the hook in emergencies:"
echo "  git commit --no-verify -m \"message\""
echo ""
echo "To skip tests temporarily:"
echo "  SKIP_TESTS=1 git commit -m \"message\""
echo ""
echo "See .git/hooks/README.md for more details."
echo ""
echo "Setup complete!"
