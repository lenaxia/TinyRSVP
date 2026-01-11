class PropertiesPanelAdvanced {
    constructor(propertiesPanel) {
        this.panel = propertiesPanel;
    }

    addAnimationControls(form, component) {
        const fields = [];
        
        if (component.animation) {
            fields.push(
                {
                    type: 'select',
                    id: 'prop-animation-type',
                    label: 'Animation Type',
                    value: component.animation.type || 'fade',
                    options: [
                        { value: 'fade', label: 'Fade' },
                        { value: 'slide', label: 'Slide' },
                        { value: 'scale', label: 'Scale' },
                        { value: 'rotate', label: 'Rotate' },
                        { value: 'bounce', label: 'Bounce' }
                    ]
                },
                {
                    type: 'number',
                    id: 'prop-animation-duration',
                    label: 'Duration (ms)',
                    value: component.animation.duration || 1000,
                    min: 0,
                    max: 10000,
                    step: 100
                },
                {
                    type: 'number',
                    id: 'prop-animation-delay',
                    label: 'Delay (ms)',
                    value: component.animation.delay || 0,
                    min: 0,
                    max: 5000,
                    step: 100
                },
                {
                    type: 'select',
                    id: 'prop-animation-easing',
                    label: 'Easing',
                    value: component.animation.easing || 'ease-in-out',
                    options: [
                        { value: 'linear', label: 'Linear' },
                        { value: 'ease-in', label: 'Ease In' },
                        { value: 'ease-out', label: 'Ease Out' },
                        { value: 'ease-in-out', label: 'Ease In-Out' }
                    ]
                },
                {
                    type: 'select',
                    id: 'prop-animation-iteration',
                    label: 'Iteration',
                    value: component.animation.iteration || 'once',
                    options: [
                        { value: 'once', label: 'Once' },
                        { value: 'infinite', label: 'Infinite' },
                        { value: 'count', label: 'Count' }
                    ]
                },
                {
                    type: 'select',
                    id: 'prop-animation-direction',
                    label: 'Direction',
                    value: component.animation.direction || 'normal',
                    options: [
                        { value: 'normal', label: 'Normal' },
                        { value: 'reverse', label: 'Reverse' },
                        { value: 'alternate', label: 'Alternate' }
                    ]
                }
            );
            
            if (component.animation.iteration === 'count') {
                fields.push({
                    type: 'number',
                    id: 'prop-animation-count',
                    label: 'Iteration Count',
                    value: component.animation.iterationCount || 1,
                    min: 1,
                    max: 100
                });
            }
        } else {
            fields.push({
                type: 'button',
                id: 'btn-add-animation',
                label: 'Add Animation',
                text: '+ Add Animation'
            });
        }
        
        if (fields.length > 0) {
            const section = this.panel.createSection('Animation', fields);
            form.appendChild(section);
        }
    }

    addAdvancedLayoutControls(form, component) {
        if (component.type !== 'Container') return;
        
        const fields = [
            {
                type: 'select',
                id: 'prop-layout-mode',
                label: 'Layout Mode',
                value: component.layoutMode || 'flexbox',
                options: [
                    { value: 'flexbox', label: 'Flexbox' },
                    { value: 'grid', label: 'Grid' },
                    { value: 'absolute', label: 'Absolute' }
                ]
            }
        ];
        
        if (component.layoutMode === 'grid' && component.gridConfig) {
            fields.push(
                {
                    type: 'text',
                    id: 'prop-grid-columns',
                    label: 'Grid Columns',
                    value: component.gridConfig.columns || 'repeat(3, 1fr)',
                    placeholder: 'e.g., repeat(3, 1fr)'
                },
                {
                    type: 'text',
                    id: 'prop-grid-rows',
                    label: 'Grid Rows',
                    value: component.gridConfig.rows || 'auto',
                    placeholder: 'e.g., auto'
                },
                {
                    type: 'text',
                    id: 'prop-grid-gap',
                    label: 'Grid Gap',
                    value: component.gridConfig.gap || '20px',
                    placeholder: 'e.g., 20px'
                },
                {
                    type: 'select',
                    id: 'prop-grid-autoflow',
                    label: 'Auto Flow',
                    value: component.gridConfig.autoFlow || 'row',
                    options: [
                        { value: 'row', label: 'Row' },
                        { value: 'column', label: 'Column' },
                        { value: 'dense', label: 'Dense' }
                    ]
                }
            );
        } else if (component.layoutMode === 'flexbox' && component.flexConfig) {
            fields.push(
                {
                    type: 'select',
                    id: 'prop-flex-direction',
                    label: 'Flex Direction',
                    value: component.flexConfig.direction || 'row',
                    options: [
                        { value: 'row', label: 'Row' },
                        { value: 'column', label: 'Column' },
                        { value: 'row-reverse', label: 'Row Reverse' },
                        { value: 'column-reverse', label: 'Column Reverse' }
                    ]
                },
                {
                    type: 'select',
                    id: 'prop-flex-wrap',
                    label: 'Flex Wrap',
                    value: component.flexConfig.wrap || 'nowrap',
                    options: [
                        { value: 'nowrap', label: 'No Wrap' },
                        { value: 'wrap', label: 'Wrap' },
                        { value: 'wrap-reverse', label: 'Wrap Reverse' }
                    ]
                },
                {
                    type: 'select',
                    id: 'prop-flex-justify',
                    label: 'Justify Content',
                    value: component.flexConfig.justifyContent || 'flex-start',
                    options: [
                        { value: 'flex-start', label: 'Start' },
                        { value: 'center', label: 'Center' },
                        { value: 'flex-end', label: 'End' },
                        { value: 'space-between', label: 'Space Between' },
                        { value: 'space-around', label: 'Space Around' },
                        { value: 'space-evenly', label: 'Space Evenly' }
                    ]
                },
                {
                    type: 'select',
                    id: 'prop-flex-align',
                    label: 'Align Items',
                    value: component.flexConfig.alignItems || 'flex-start',
                    options: [
                        { value: 'flex-start', label: 'Start' },
                        { value: 'center', label: 'Center' },
                        { value: 'flex-end', label: 'End' },
                        { value: 'stretch', label: 'Stretch' },
                        { value: 'baseline', label: 'Baseline' }
                    ]
                },
                {
                    type: 'text',
                    id: 'prop-flex-gap',
                    label: 'Gap',
                    value: component.flexConfig.gap || '20px',
                    placeholder: 'e.g., 20px'
                }
            );
        }
        
        if (fields.length > 1) {
            const section = this.panel.createSection('Advanced Layout', fields);
            form.appendChild(section);
        }
    }

    addImageEffectsControls(form, component) {
        if (component.type !== 'Image') return;
        
        const fields = [
            {
                type: 'text',
                id: 'prop-image-filter',
                label: 'Filter',
                value: component.content?.filter || '',
                placeholder: 'e.g., blur(5px), grayscale(100%)'
            },
            {
                type: 'text',
                id: 'prop-image-transform',
                label: 'Transform',
                value: component.content?.transform || '',
                placeholder: 'e.g., rotate(45deg), scale(1.2)'
            },
            {
                type: 'select',
                id: 'prop-image-blend',
                label: 'Blend Mode',
                value: component.content?.blendMode || 'normal',
                options: [
                    { value: 'normal', label: 'Normal' },
                    { value: 'multiply', label: 'Multiply' },
                    { value: 'screen', label: 'Screen' },
                    { value: 'overlay', label: 'Overlay' },
                    { value: 'darken', label: 'Darken' },
                    { value: 'lighten', label: 'Lighten' }
                ]
            },
            {
                type: 'text',
                id: 'prop-image-mask',
                label: 'Mask URL',
                value: component.content?.mask || '',
                placeholder: '/images/mask.png'
            },
            {
                type: 'text',
                id: 'prop-image-clippath',
                label: 'Clip Path',
                value: component.content?.clipPath || '',
                placeholder: 'e.g., circle(50%), polygon(...)'
            }
        ];
        
        const section = this.panel.createSection('Image Effects', fields);
        form.appendChild(section);
    }

    addTextEffectsControls(form, component) {
        if (component.type !== 'TextBox') return;
        
        const fields = [
            {
                type: 'text',
                id: 'prop-text-gradient',
                label: 'Text Gradient',
                value: component.content?.gradient || '',
                placeholder: 'linear-gradient(90deg, #ff0000, #00ff00)'
            },
            {
                type: 'text',
                id: 'prop-text-stroke',
                label: 'Text Stroke',
                value: component.content?.stroke || '',
                placeholder: '2px #000000'
            },
            {
                type: 'text',
                id: 'prop-text-shadow',
                label: 'Text Shadow',
                value: component.content?.shadow || '',
                placeholder: '2px 2px 4px rgba(0,0,0,0.5)'
            },
            {
                type: 'select',
                id: 'prop-text-transform',
                label: 'Text Transform',
                value: component.content?.transform || 'none',
                options: [
                    { value: 'none', label: 'None' },
                    { value: 'uppercase', label: 'Uppercase' },
                    { value: 'lowercase', label: 'Lowercase' },
                    { value: 'capitalize', label: 'Capitalize' }
                ]
            },
            {
                type: 'text',
                id: 'prop-text-letter-spacing',
                label: 'Letter Spacing',
                value: component.content?.letterSpacing || '',
                placeholder: 'e.g., 0.1em'
            },
            {
                type: 'text',
                id: 'prop-text-line-height',
                label: 'Line Height',
                value: component.content?.lineHeight || '',
                placeholder: 'e.g., 1.5'
            },
            {
                type: 'text',
                id: 'prop-text-word-spacing',
                label: 'Word Spacing',
                value: component.content?.wordSpacing || '',
                placeholder: 'e.g., 0.2em'
            }
        ];
        
        const section = this.panel.createSection('Text Effects', fields);
        form.appendChild(section);
    }

    addVisibilityControls(form, component) {
        const visibility = component.visibility || {
            showOnMobile: true,
            showOnTablet: true,
            showOnDesktop: true
        };
        
        const fields = [
            {
                type: 'checkbox',
                id: 'prop-visibility-mobile',
                label: 'Show on Mobile',
                checked: visibility.showOnMobile !== false
            },
            {
                type: 'checkbox',
                id: 'prop-visibility-tablet',
                label: 'Show on Tablet',
                checked: visibility.showOnTablet !== false
            },
            {
                type: 'checkbox',
                id: 'prop-visibility-desktop',
                label: 'Show on Desktop',
                checked: visibility.showOnDesktop !== false
            },
            {
                type: 'text',
                id: 'prop-visibility-when',
                label: 'Show When (Expression)',
                value: visibility.showWhen || '',
                placeholder: '{{.Event.HasImage}}'
            }
        ];
        
        const section = this.panel.createSection('Visibility Rules', fields);
        form.appendChild(section);
    }

    extendShowComponent(originalMethod) {
        return function(component) {
            originalMethod.call(this, component);
            
            const form = this.panel.querySelector('.properties-form');
            if (!form) return;
            
            const advanced = new PropertiesPanelAdvanced(this);
            advanced.addAnimationControls(form, component);
            advanced.addAdvancedLayoutControls(form, component);
            
            if (component.type === 'Image') {
                advanced.addImageEffectsControls(form, component);
            }
            
            if (component.type === 'TextBox') {
                advanced.addTextEffectsControls(form, component);
            }
            
            advanced.addVisibilityControls(form, component);
        };
    }

    extendParsePropertyPath(originalMethod) {
        return function(property, value) {
            const updates = originalMethod.call(this, property, value);
            const parts = property.split('-');
            
            if (parts[0] === 'animation') {
                if (!updates.animation) updates.animation = {};
                if (parts[1] === 'type') updates.animation.type = value;
                else if (parts[1] === 'duration') updates.animation.duration = parseInt(value);
                else if (parts[1] === 'delay') updates.animation.delay = parseInt(value);
                else if (parts[1] === 'easing') updates.animation.easing = value;
                else if (parts[1] === 'iteration') updates.animation.iteration = value;
                else if (parts[1] === 'count') updates.animation.iterationCount = parseInt(value);
                else if (parts[1] === 'direction') updates.animation.direction = value;
            } else if (parts[0] === 'grid') {
                if (!updates.gridConfig) updates.gridConfig = {};
                if (parts[1] === 'columns') updates.gridConfig.columns = value;
                else if (parts[1] === 'rows') updates.gridConfig.rows = value;
                else if (parts[1] === 'gap') updates.gridConfig.gap = value;
                else if (parts[1] === 'autoflow') updates.gridConfig.autoFlow = value;
            } else if (parts[0] === 'flex') {
                if (!updates.flexConfig) updates.flexConfig = {};
                if (parts[1] === 'direction') updates.flexConfig.direction = value;
                else if (parts[1] === 'wrap') updates.flexConfig.wrap = value;
                else if (parts[1] === 'justify') updates.flexConfig.justifyContent = value;
                else if (parts[1] === 'align') updates.flexConfig.alignItems = value;
                else if (parts[1] === 'gap') updates.flexConfig.gap = value;
            } else if (parts[0] === 'image') {
                if (!updates.content) updates.content = {};
                if (parts[1] === 'filter') updates.content.filter = value;
                else if (parts[1] === 'transform') updates.content.transform = value;
                else if (parts[1] === 'blend') updates.content.blendMode = value;
                else if (parts[1] === 'mask') updates.content.mask = value;
                else if (parts[1] === 'clippath') updates.content.clipPath = value;
            } else if (parts[0] === 'text') {
                if (!updates.content) updates.content = {};
                if (parts[1] === 'gradient') updates.content.gradient = value;
                else if (parts[1] === 'stroke') updates.content.stroke = value;
                else if (parts[1] === 'shadow') updates.content.shadow = value;
                else if (parts[1] === 'transform') updates.content.transform = value;
                else if (parts[1] === 'letter' && parts[2] === 'spacing') updates.content.letterSpacing = value;
                else if (parts[1] === 'line' && parts[2] === 'height') updates.content.lineHeight = value;
                else if (parts[1] === 'word' && parts[2] === 'spacing') updates.content.wordSpacing = value;
            } else if (parts[0] === 'visibility') {
                if (!updates.visibility) updates.visibility = {};
                if (parts[1] === 'mobile') updates.visibility.showOnMobile = value;
                else if (parts[1] === 'tablet') updates.visibility.showOnTablet = value;
                else if (parts[1] === 'desktop') updates.visibility.showOnDesktop = value;
                else if (parts[1] === 'when') updates.visibility.showWhen = value;
            } else if (parts[0] === 'layout' && parts[1] === 'mode') {
                updates.layoutMode = value;
            }
            
            return updates;
        };
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = PropertiesPanelAdvanced;
}
