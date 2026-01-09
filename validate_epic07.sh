#!/bin/bash

set -e

echo "=========================================="
echo "Epic 07 Frontend Validation Script"
echo "=========================================="
echo ""

PASS_COUNT=0
FAIL_COUNT=0
TOTAL_STORIES=22

validate_story() {
    local story_num=$1
    local story_name=$2
    local story_file="docs/00_BACKLOG/07_STORY_${story_num}_${story_name}.md"
    
    echo "Validating Story ${story_num}: ${story_name}"
    
    if [ ! -f "$story_file" ]; then
        echo "  ❌ Story file not found: $story_file"
        ((FAIL_COUNT++))
        return 1
    fi
    
    local status=$(grep -E "^\*\*Status:\*\*" "$story_file" | head -1 | sed 's/.*Status:\*\* *//' | tr -d '\r')
    echo "  Status: $status"
    
    if [[ "$status" == "Complete"* ]]; then
        echo "  ✅ PASS"
        ((PASS_COUNT++))
        return 0
    else
        echo "  ❌ FAIL - Status is not Complete"
        ((FAIL_COUNT++))
        return 1
    fi
}

echo "Phase 1: Design System (Stories 00-03)"
echo "----------------------------------------"
validate_story "00" "css_variables"
validate_story "01" "typography"
validate_story "02" "color_system"
validate_story "03" "spacing_system"
echo ""

echo "Phase 2: Layout Components (Stories 04-07)"
echo "----------------------------------------"
validate_story "04" "responsive_grid"
validate_story "05" "navigation"
validate_story "06" "forms"
validate_story "07" "buttons"
echo ""

echo "Phase 3: Admin UI (Stories 08-12)"
echo "----------------------------------------"
validate_story "08" "dashboard_ui"
validate_story "09" "event_list_ui"
validate_story "10" "event_form_ui"
validate_story "11" "invite_list_ui"
validate_story "12" "rsvp_summary_ui"
echo ""

echo "Phase 4: Guest UI (Stories 13-15)"
echo "----------------------------------------"
validate_story "13" "rsvp_page_ui"
validate_story "14" "confirmation_ui"
validate_story "15" "mobile_optimization"
echo ""

echo "Phase 5: Interactivity (Stories 16-18)"
echo "----------------------------------------"
validate_story "16" "form_validation_js"
validate_story "17" "loading_states"
validate_story "18" "error_display"
echo ""

echo "Phase 6: Accessibility (Stories 19-21)"
echo "----------------------------------------"
validate_story "19" "keyboard_navigation"
validate_story "20" "screen_reader"
validate_story "21" "focus_management"
echo ""

echo "=========================================="
echo "Validation Summary"
echo "=========================================="
echo "Total Stories: $TOTAL_STORIES"
echo "Passed: $PASS_COUNT"
echo "Failed: $FAIL_COUNT"
echo ""

if [ $FAIL_COUNT -eq 0 ]; then
    echo "✅ ALL STORIES COMPLETE!"
    exit 0
else
    echo "❌ SOME STORIES INCOMPLETE"
    exit 1
fi
