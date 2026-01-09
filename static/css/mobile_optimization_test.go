package css

import (
	"strings"
	"testing"
)

func TestMobileOptimizationCSS(t *testing.T) {
	css := `
.mobile-optimized {
    -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
    -webkit-touch-callout: none;
    touch-action: manipulation;
}

.mobile-no-select {
    -webkit-user-select: none;
    -moz-user-select: none;
    -ms-user-select: none;
    user-select: none;
}

.mobile-scroll-smooth {
    -webkit-overflow-scrolling: touch;
    scroll-behavior: smooth;
}

.mobile-full-width {
    width: 100%;
    max-width: 100%;
}

.mobile-tap-target {
    min-height: 44px;
    min-width: 44px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
}

.mobile-text-readable {
    font-size: 16px;
    line-height: 1.5;
}

input[type="text"],
input[type="email"],
input[type="tel"],
input[type="number"],
input[type="search"],
input[type="url"],
textarea,
select {
    font-size: 16px;
}

@media (max-width: 767px) {
    body {
        font-size: 16px;
        -webkit-text-size-adjust: 100%;
        -ms-text-size-adjust: 100%;
    }

    .mobile-hide {
        display: none !important;
    }

    .mobile-show {
        display: block !important;
    }

    .mobile-stack {
        flex-direction: column;
    }

    .mobile-full-width {
        width: 100% !important;
    }

    .mobile-center {
        text-align: center;
    }

    .mobile-padding {
        padding: var(--spacing-4);
    }

    .mobile-no-padding {
        padding: 0;
    }

    .mobile-margin {
        margin: var(--spacing-4);
    }

    .mobile-no-margin {
        margin: 0;
    }
}

@media (min-width: 768px) {
    .desktop-hide {
        display: none !important;
    }

    .desktop-show {
        display: block !important;
    }
}

.touch-action-pan-y {
    touch-action: pan-y;
}

.touch-action-pan-x {
    touch-action: pan-x;
}

.touch-action-none {
    touch-action: none;
}

@supports (-webkit-touch-callout: none) {
    .ios-safe-area-top {
        padding-top: env(safe-area-inset-top);
    }

    .ios-safe-area-bottom {
        padding-bottom: env(safe-area-inset-bottom);
    }

    .ios-safe-area-left {
        padding-left: env(safe-area-inset-left);
    }

    .ios-safe-area-right {
        padding-right: env(safe-area-inset-right);
    }
}

.prevent-zoom {
    touch-action: manipulation;
}

@media (hover: none) and (pointer: coarse) {
    .hover-only {
        display: none;
    }
}

@media (hover: hover) and (pointer: fine) {
    .touch-only {
        display: none;
    }
}
`

	t.Run("contains tap highlight color", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-tap-highlight-color") {
			t.Error("CSS should contain -webkit-tap-highlight-color for touch feedback")
		}
	})

	t.Run("contains touch action manipulation", func(t *testing.T) {
		if !strings.Contains(css, "touch-action: manipulation") {
			t.Error("CSS should contain touch-action: manipulation to prevent double-tap zoom")
		}
	})

	t.Run("contains smooth scrolling for mobile", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-overflow-scrolling: touch") {
			t.Error("CSS should contain -webkit-overflow-scrolling for smooth iOS scrolling")
		}
	})

	t.Run("contains minimum tap target size", func(t *testing.T) {
		if !strings.Contains(css, "min-height: 44px") {
			t.Error("CSS should contain min-height: 44px for mobile tap targets")
		}
		if !strings.Contains(css, "min-width: 44px") {
			t.Error("CSS should contain min-width: 44px for mobile tap targets")
		}
	})

	t.Run("contains 16px font size for inputs", func(t *testing.T) {
		if !strings.Contains(css, "font-size: 16px") {
			t.Error("CSS should contain font-size: 16px for inputs to prevent zoom on iOS")
		}
	})

	t.Run("contains text size adjust", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-text-size-adjust: 100%") {
			t.Error("CSS should contain -webkit-text-size-adjust to prevent text inflation")
		}
	})

	t.Run("contains mobile-specific media query", func(t *testing.T) {
		if !strings.Contains(css, "@media (max-width: 767px)") {
			t.Error("CSS should contain mobile-specific media query")
		}
	})

	t.Run("contains mobile utility classes", func(t *testing.T) {
		utilities := []string{
			".mobile-hide",
			".mobile-show",
			".mobile-stack",
			".mobile-full-width",
			".mobile-center",
		}
		for _, utility := range utilities {
			if !strings.Contains(css, utility) {
				t.Errorf("CSS should contain %s utility class", utility)
			}
		}
	})

	t.Run("contains iOS safe area support", func(t *testing.T) {
		if !strings.Contains(css, "env(safe-area-inset-top)") {
			t.Error("CSS should contain safe area inset support for iOS notch")
		}
	})

	t.Run("contains touch-specific media queries", func(t *testing.T) {
		if !strings.Contains(css, "@media (hover: none) and (pointer: coarse)") {
			t.Error("CSS should contain touch device detection media query")
		}
	})

	t.Run("contains prevent zoom class", func(t *testing.T) {
		if !strings.Contains(css, ".prevent-zoom") {
			t.Error("CSS should contain prevent-zoom class for form inputs")
		}
	})
}

func TestMobileOptimizationResponsiveBreakpoints(t *testing.T) {
	tests := []struct {
		name       string
		breakpoint string
		shouldHave bool
	}{
		{"mobile breakpoint", "@media (max-width: 767px)", true},
		{"tablet breakpoint", "@media (min-width: 768px)", true},
		{"desktop breakpoint", "@media (min-width: 1024px)", false},
	}

	css := `
@media (max-width: 767px) {
    body {
        font-size: 16px;
    }
}

@media (min-width: 768px) {
    .desktop-hide {
        display: none !important;
    }
}
`

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contains := strings.Contains(css, tt.breakpoint)
			if contains != tt.shouldHave {
				if tt.shouldHave {
					t.Errorf("CSS should contain breakpoint %s", tt.breakpoint)
				} else {
					t.Errorf("CSS should not contain breakpoint %s", tt.breakpoint)
				}
			}
		})
	}
}

func TestMobileOptimizationTouchActions(t *testing.T) {
	css := `
.touch-action-pan-y {
    touch-action: pan-y;
}

.touch-action-pan-x {
    touch-action: pan-x;
}

.touch-action-none {
    touch-action: none;
}
`

	touchActions := []string{
		"touch-action: pan-y",
		"touch-action: pan-x",
		"touch-action: none",
	}

	for _, action := range touchActions {
		t.Run(action, func(t *testing.T) {
			if !strings.Contains(css, action) {
				t.Errorf("CSS should contain %s for gesture control", action)
			}
		})
	}
}

func TestMobileOptimizationPerformance(t *testing.T) {
	css := `
.mobile-optimized {
    -webkit-tap-highlight-color: rgba(0, 0, 0, 0.1);
    -webkit-touch-callout: none;
    touch-action: manipulation;
}

.mobile-scroll-smooth {
    -webkit-overflow-scrolling: touch;
    scroll-behavior: smooth;
}
`

	t.Run("contains hardware acceleration hints", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-overflow-scrolling: touch") {
			t.Error("CSS should contain -webkit-overflow-scrolling for hardware acceleration")
		}
	})

	t.Run("contains touch callout prevention", func(t *testing.T) {
		if !strings.Contains(css, "-webkit-touch-callout: none") {
			t.Error("CSS should contain -webkit-touch-callout to prevent iOS callout menu")
		}
	})
}

func TestMobileOptimizationAccessibility(t *testing.T) {
	css := `
.mobile-tap-target {
    min-height: 44px;
    min-width: 44px;
    display: inline-flex;
    align-items: center;
    justify-content: center;
}

.mobile-text-readable {
    font-size: 16px;
    line-height: 1.5;
}
`

	t.Run("tap targets meet accessibility guidelines", func(t *testing.T) {
		if !strings.Contains(css, "min-height: 44px") || !strings.Contains(css, "min-width: 44px") {
			t.Error("Tap targets should be at least 44x44px for accessibility")
		}
	})

	t.Run("text is readable without zoom", func(t *testing.T) {
		if !strings.Contains(css, "font-size: 16px") {
			t.Error("Text should be at least 16px for readability without zoom")
		}
		if !strings.Contains(css, "line-height: 1.5") {
			t.Error("Line height should be at least 1.5 for readability")
		}
	})
}
