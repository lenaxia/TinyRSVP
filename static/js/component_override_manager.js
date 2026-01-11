/**
 * Component Override Manager
 * Tracks and manages component customizations
 */

class ComponentOverrideManager {
    constructor(eventCustomization) {
        this.customization = eventCustomization;
        this.init();
    }

    init() {
        this.setupOverrideTracking();
    }

    setupOverrideTracking() {
        document.getElementById('btn-show-overrides')?.addEventListener('click', () => {
            this.showOverridesSummary();
        });

        document.getElementById('btn-export-overrides')?.addEventListener('click', () => {
            this.exportOverrides();
        });

        document.getElementById('btn-import-overrides')?.addEventListener('click', () => {
            this.importOverrides();
        });
    }

    getOverrideSummary() {
        const overrides = this.customization.customizationData.eventOverrides;
        
        if (!overrides) {
            return {
                totalOverrides: 0,
                modifiedComponents: [],
                addedComponents: [],
                removedComponents: []
            };
        }

        return {
            totalOverrides: (overrides.overrides?.length || 0) + 
                          (overrides.additions?.length || 0) + 
                          (overrides.removals?.length || 0),
            modifiedComponents: overrides.overrides?.map(o => o.id) || [],
            addedComponents: overrides.additions?.map(a => a.id) || [],
            removedComponents: overrides.removals || []
        };
    }

    showOverridesSummary() {
        const summary = this.getOverrideSummary();
        const modal = this.createSummaryModal(summary);
        document.body.appendChild(modal);
    }

    createSummaryModal(summary) {
        const modal = document.createElement('div');
        modal.className = 'modal-overlay';
        modal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Customization Summary</h3>
                    <button type="button" class="modal-close">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="summary-stat">
                        <strong>Total Changes:</strong> ${summary.totalOverrides}
                    </div>
                    
                    ${summary.modifiedComponents.length > 0 ? `
                        <div class="summary-section">
                            <h4>Modified Components (${summary.modifiedComponents.length})</h4>
                            <ul class="summary-list">
                                ${summary.modifiedComponents.map(id => `
                                    <li>
                                        <span>${this.customization.formatComponentName(id)}</span>
                                        <button type="button" class="btn-view-component" data-component-id="${id}">View</button>
                                        <button type="button" class="btn-reset-single" data-component-id="${id}">Reset</button>
                                    </li>
                                `).join('')}
                            </ul>
                        </div>
                    ` : ''}
                    
                    ${summary.addedComponents.length > 0 ? `
                        <div class="summary-section">
                            <h4>Added Components (${summary.addedComponents.length})</h4>
                            <ul class="summary-list">
                                ${summary.addedComponents.map(id => `
                                    <li>
                                        <span>${this.customization.formatComponentName(id)}</span>
                                        <button type="button" class="btn-view-component" data-component-id="${id}">View</button>
                                        <button type="button" class="btn-remove-single" data-component-id="${id}">Remove</button>
                                    </li>
                                `).join('')}
                            </ul>
                        </div>
                    ` : ''}
                    
                    ${summary.removedComponents.length > 0 ? `
                        <div class="summary-section">
                            <h4>Removed Components (${summary.removedComponents.length})</h4>
                            <ul class="summary-list">
                                ${summary.removedComponents.map(id => `
                                    <li>
                                        <span>${this.customization.formatComponentName(id)}</span>
                                        <button type="button" class="btn-restore-single" data-component-id="${id}">Restore</button>
                                    </li>
                                `).join('')}
                            </ul>
                        </div>
                    ` : ''}
                    
                    ${summary.totalOverrides === 0 ? `
                        <p class="summary-empty">No customizations yet</p>
                    ` : ''}
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn-cancel" id="summary-close">Close</button>
                </div>
            </div>
        `;

        modal.querySelector('.modal-close').addEventListener('click', () => modal.remove());
        modal.querySelector('#summary-close').addEventListener('click', () => modal.remove());

        modal.querySelectorAll('.btn-view-component').forEach(btn => {
            btn.addEventListener('click', () => {
                this.customization.selectComponent(btn.dataset.componentId);
                modal.remove();
            });
        });

        modal.querySelectorAll('.btn-reset-single').forEach(btn => {
            btn.addEventListener('click', () => {
                this.resetSingleOverride(btn.dataset.componentId);
                modal.remove();
                this.showOverridesSummary();
            });
        });

        modal.querySelectorAll('.btn-remove-single').forEach(btn => {
            btn.addEventListener('click', () => {
                this.removeSingleAddition(btn.dataset.componentId);
                modal.remove();
                this.showOverridesSummary();
            });
        });

        modal.querySelectorAll('.btn-restore-single').forEach(btn => {
            btn.addEventListener('click', () => {
                this.restoreSingleComponent(btn.dataset.componentId);
                modal.remove();
                this.showOverridesSummary();
            });
        });

        return modal;
    }

    resetSingleOverride(componentID) {
        if (!this.customization.customizationData.eventOverrides) return;

        this.customization.customizationData.eventOverrides.overrides = 
            this.customization.customizationData.eventOverrides.overrides.filter(
                o => o.id !== componentID
            );

        this.customization.isDirty = true;
        this.customization.renderComponentList();
        this.customization.renderPreview();
        this.customization.showSuccess('Override reset');
    }

    removeSingleAddition(componentID) {
        if (!this.customization.customizationData.eventOverrides) return;

        this.customization.customizationData.eventOverrides.additions = 
            this.customization.customizationData.eventOverrides.additions.filter(
                a => a.id !== componentID
            );

        this.customization.isDirty = true;
        this.customization.renderComponentList();
        this.customization.renderPreview();
        this.customization.showSuccess('Component removed');
    }

    restoreSingleComponent(componentID) {
        if (!this.customization.customizationData.eventOverrides) return;

        this.customization.customizationData.eventOverrides.removals = 
            this.customization.customizationData.eventOverrides.removals.filter(
                id => id !== componentID
            );

        this.customization.isDirty = true;
        this.customization.renderComponentList();
        this.customization.renderPreview();
        this.customization.showSuccess('Component restored');
    }

    exportOverrides() {
        const overrides = this.customization.customizationData.eventOverrides;
        
        if (!overrides || this.getOverrideSummary().totalOverrides === 0) {
            this.customization.showError('No customizations to export');
            return;
        }

        const dataStr = JSON.stringify(overrides, null, 2);
        const dataBlob = new Blob([dataStr], { type: 'application/json' });
        const url = URL.createObjectURL(dataBlob);
        
        const link = document.createElement('a');
        link.href = url;
        link.download = `event-${this.customization.eventID}-customization.json`;
        link.click();
        
        URL.revokeObjectURL(url);
        this.customization.showSuccess('Customization exported');
    }

    importOverrides() {
        const input = document.createElement('input');
        input.type = 'file';
        input.accept = 'application/json';
        
        input.addEventListener('change', async (e) => {
            const file = e.target.files[0];
            if (!file) return;

            try {
                const text = await file.text();
                const overrides = JSON.parse(text);
                
                if (!this.validateOverrides(overrides)) {
                    throw new Error('Invalid customization file');
                }

                this.customization.customizationData.eventOverrides = overrides;
                this.customization.isDirty = true;
                this.customization.renderComponentList();
                this.customization.renderPreview();
                this.customization.showSuccess('Customization imported');
            } catch (error) {
                console.error('Import error:', error);
                this.customization.showError('Failed to import customization');
            }
        });

        input.click();
    }

    validateOverrides(overrides) {
        if (!overrides || typeof overrides !== 'object') return false;
        if (!overrides.version) return false;
        if (!Array.isArray(overrides.overrides)) return false;
        if (!Array.isArray(overrides.additions)) return false;
        if (!Array.isArray(overrides.removals)) return false;
        return true;
    }

    getOriginalValue(componentID, propertyPath) {
        const templateComponent = this.customization.customizationData.templateConfig?.components?.find(
            c => c.id === componentID
        );

        if (!templateComponent) return null;

        let value = templateComponent;
        for (const key of propertyPath) {
            if (value && typeof value === 'object') {
                value = value[key];
            } else {
                return null;
            }
        }

        return value;
    }

    getCustomizedValue(componentID, propertyPath) {
        const override = this.customization.customizationData.eventOverrides?.overrides?.find(
            o => o.id === componentID
        );

        if (!override) return null;

        let value = override.updates;
        for (const key of propertyPath) {
            if (value && typeof value === 'object') {
                value = value[key];
            } else {
                return null;
            }
        }

        return value;
    }

    compareValues(componentID) {
        const component = this.customization.getComponent(componentID);
        if (!component) return {};

        const comparison = {};
        const properties = ['content', 'position', 'dimensions', 'zIndex', 'visible'];

        properties.forEach(prop => {
            const original = this.getOriginalValue(componentID, [prop]);
            const customized = this.getCustomizedValue(componentID, [prop]);
            
            if (customized !== null && JSON.stringify(original) !== JSON.stringify(customized)) {
                comparison[prop] = {
                    original,
                    customized,
                    changed: true
                };
            }
        });

        return comparison;
    }
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = ComponentOverrideManager;
}
