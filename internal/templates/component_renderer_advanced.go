package templates

import (
	"fmt"
	"strings"

	"github.com/lenaxia/tinyrsvp/internal/models"
)

func (r *ComponentRenderer) GenerateAnimationCSS(comp *models.Component) string {
	if comp.Animation == nil {
		return ""
	}

	anim := comp.Animation
	var css strings.Builder

	css.WriteString(fmt.Sprintf("animation-name: %s-%s;\n", comp.ID, anim.Type))
	css.WriteString(fmt.Sprintf("animation-duration: %dms;\n", anim.Duration))
	css.WriteString(fmt.Sprintf("animation-delay: %dms;\n", anim.Delay))
	css.WriteString(fmt.Sprintf("animation-timing-function: %s;\n", anim.Easing))

	switch anim.Iteration {
	case models.AnimationIterationOnce:
		css.WriteString("animation-iteration-count: 1;\n")
	case models.AnimationIterationInfinite:
		css.WriteString("animation-iteration-count: infinite;\n")
	case models.AnimationIterationCount:
		if anim.IterationCount != nil {
			css.WriteString(fmt.Sprintf("animation-iteration-count: %d;\n", *anim.IterationCount))
		}
	}

	css.WriteString(fmt.Sprintf("animation-direction: %s;\n", anim.Direction))

	return css.String()
}

func (r *ComponentRenderer) GenerateAnimationKeyframes(animType models.AnimationType) string {
	var keyframes strings.Builder

	switch animType {
	case models.AnimationTypeFade:
		keyframes.WriteString("@keyframes fade {\n")
		keyframes.WriteString("  from { opacity: 0; }\n")
		keyframes.WriteString("  to { opacity: 1; }\n")
		keyframes.WriteString("}\n")

	case models.AnimationTypeSlide:
		keyframes.WriteString("@keyframes slide {\n")
		keyframes.WriteString("  from { transform: translateY(-20px); opacity: 0; }\n")
		keyframes.WriteString("  to { transform: translateY(0); opacity: 1; }\n")
		keyframes.WriteString("}\n")

	case models.AnimationTypeScale:
		keyframes.WriteString("@keyframes scale {\n")
		keyframes.WriteString("  from { transform: scale(0.8); opacity: 0; }\n")
		keyframes.WriteString("  to { transform: scale(1); opacity: 1; }\n")
		keyframes.WriteString("}\n")

	case models.AnimationTypeRotate:
		keyframes.WriteString("@keyframes rotate {\n")
		keyframes.WriteString("  from { transform: rotate(0deg); }\n")
		keyframes.WriteString("  to { transform: rotate(360deg); }\n")
		keyframes.WriteString("}\n")

	case models.AnimationTypeBounce:
		keyframes.WriteString("@keyframes bounce {\n")
		keyframes.WriteString("  0%, 100% { transform: translateY(0); }\n")
		keyframes.WriteString("  50% { transform: translateY(-20px); }\n")
		keyframes.WriteString("}\n")
	}

	return keyframes.String()
}

func (r *ComponentRenderer) GenerateLayoutCSS(comp *models.Component) string {
	if comp.LayoutMode == nil {
		return ""
	}

	var css strings.Builder

	switch *comp.LayoutMode {
	case models.LayoutModeGrid:
		if comp.GridConfig != nil {
			css.WriteString("display: grid;\n")
			css.WriteString(fmt.Sprintf("grid-template-columns: %s;\n", comp.GridConfig.Columns))
			if comp.GridConfig.Rows != "" {
				css.WriteString(fmt.Sprintf("grid-template-rows: %s;\n", comp.GridConfig.Rows))
			}
			if comp.GridConfig.Gap != "" {
				css.WriteString(fmt.Sprintf("gap: %s;\n", comp.GridConfig.Gap))
			}
			css.WriteString(fmt.Sprintf("grid-auto-flow: %s;\n", comp.GridConfig.AutoFlow))
		}

	case models.LayoutModeFlexbox:
		if comp.FlexConfig != nil {
			css.WriteString("display: flex;\n")
			css.WriteString(fmt.Sprintf("flex-direction: %s;\n", comp.FlexConfig.Direction))
			css.WriteString(fmt.Sprintf("flex-wrap: %s;\n", comp.FlexConfig.Wrap))
			css.WriteString(fmt.Sprintf("justify-content: %s;\n", comp.FlexConfig.JustifyContent))
			css.WriteString(fmt.Sprintf("align-items: %s;\n", comp.FlexConfig.AlignItems))
			if comp.FlexConfig.Gap != "" {
				css.WriteString(fmt.Sprintf("gap: %s;\n", comp.FlexConfig.Gap))
			}
		}

	case models.LayoutModeAbsolute:
		css.WriteString("position: absolute;\n")
	}

	return css.String()
}

func (r *ComponentRenderer) GenerateImageEffectsCSS(comp *models.Component, effects *models.ImageEffects) string {
	if effects == nil {
		return ""
	}

	var css strings.Builder

	if effects.Filter != nil && *effects.Filter != "" {
		css.WriteString(fmt.Sprintf("filter: %s;\n", *effects.Filter))
	}

	if effects.Transform != nil && *effects.Transform != "" {
		css.WriteString(fmt.Sprintf("transform: %s;\n", *effects.Transform))
	}

	if effects.BlendMode != nil && *effects.BlendMode != "" {
		css.WriteString(fmt.Sprintf("mix-blend-mode: %s;\n", *effects.BlendMode))
	}

	if effects.Mask != nil && *effects.Mask != "" {
		css.WriteString(fmt.Sprintf("mask-image: url(%s);\n", *effects.Mask))
		css.WriteString(fmt.Sprintf("-webkit-mask-image: url(%s);\n", *effects.Mask))
	}

	if effects.ClipPath != nil && *effects.ClipPath != "" {
		css.WriteString(fmt.Sprintf("clip-path: %s;\n", *effects.ClipPath))
	}

	return css.String()
}

func (r *ComponentRenderer) GenerateTextEffectsCSS(comp *models.Component, effects *models.TextEffects) string {
	if effects == nil {
		return ""
	}

	var css strings.Builder

	if effects.Gradient != nil && *effects.Gradient != "" {
		css.WriteString(fmt.Sprintf("background: %s;\n", *effects.Gradient))
		css.WriteString("-webkit-background-clip: text;\n")
		css.WriteString("background-clip: text;\n")
		css.WriteString("-webkit-text-fill-color: transparent;\n")
	}

	if effects.Stroke != nil && *effects.Stroke != "" {
		css.WriteString(fmt.Sprintf("-webkit-text-stroke: %s;\n", *effects.Stroke))
	}

	if effects.Shadow != nil && *effects.Shadow != "" {
		css.WriteString(fmt.Sprintf("text-shadow: %s;\n", *effects.Shadow))
	}

	if effects.Transform != nil && *effects.Transform != "" {
		css.WriteString(fmt.Sprintf("text-transform: %s;\n", *effects.Transform))
	}

	if effects.LetterSpacing != nil && *effects.LetterSpacing != "" {
		css.WriteString(fmt.Sprintf("letter-spacing: %s;\n", *effects.LetterSpacing))
	}

	if effects.LineHeight != nil && *effects.LineHeight != "" {
		css.WriteString(fmt.Sprintf("line-height: %s;\n", *effects.LineHeight))
	}

	if effects.WordSpacing != nil && *effects.WordSpacing != "" {
		css.WriteString(fmt.Sprintf("word-spacing: %s;\n", *effects.WordSpacing))
	}

	return css.String()
}

func (r *ComponentRenderer) GenerateVisibilityCSS(comp *models.Component) string {
	if comp.Visibility == nil {
		return ""
	}

	vis := comp.Visibility
	var css strings.Builder

	if !vis.ShowOnMobile {
		css.WriteString(fmt.Sprintf("@media (max-width: 767px) {\n  #%s { display: none; }\n}\n", comp.ID))
	}

	if !vis.ShowOnTablet {
		css.WriteString(fmt.Sprintf("@media (min-width: 768px) and (max-width: 1023px) {\n  #%s { display: none; }\n}\n", comp.ID))
	}

	if !vis.ShowOnDesktop {
		css.WriteString(fmt.Sprintf("@media (min-width: 1024px) {\n  #%s { display: none; }\n}\n", comp.ID))
	}

	return css.String()
}

func (r *ComponentRenderer) GenerateComponentCSS(comp *models.Component) string {
	var css strings.Builder

	css.WriteString(fmt.Sprintf("#%s {\n", comp.ID))

	if comp.Animation != nil {
		animCSS := r.GenerateAnimationCSS(comp)
		if animCSS != "" {
			css.WriteString("  ")
			css.WriteString(strings.ReplaceAll(animCSS, "\n", "\n  "))
		}
	}

	if comp.LayoutMode != nil {
		layoutCSS := r.GenerateLayoutCSS(comp)
		if layoutCSS != "" {
			css.WriteString("  ")
			css.WriteString(strings.ReplaceAll(layoutCSS, "\n", "\n  "))
		}
	}

	css.WriteString("}\n")

	if comp.Visibility != nil {
		visCSS := r.GenerateVisibilityCSS(comp)
		if visCSS != "" {
			css.WriteString(visCSS)
		}
	}

	return css.String()
}
