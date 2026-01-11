class ImageUploader {
    constructor() {
        this.input = document.getElementById('image-upload-input');
        this.uploadBtn = document.getElementById('image-upload-btn');
        this.removeBtn = document.getElementById('remove-image-btn');
        this.preview = document.getElementById('image-preview');
        this.previewImage = document.getElementById('preview-image');
        this.progress = document.getElementById('upload-progress');
        this.progressFill = document.getElementById('progress-fill');
        this.progressText = document.getElementById('progress-text');
        this.errorDiv = document.getElementById('upload-error');
        this.errorMessage = document.getElementById('error-message');
        this.successDiv = document.getElementById('upload-success');
        this.hiddenInput = document.getElementById('custom-theme-image-url');
        
        this.maxSize = 5 * 1024 * 1024;
        this.maxDimensions = 4096;
        this.allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
        
        this.init();
    }

    init() {
        if (!this.input) return;
        
        this.attachEventListeners();
    }

    attachEventListeners() {
        if (this.uploadBtn) {
            this.uploadBtn.addEventListener('click', () => {
                this.input.click();
            });
        }

        this.input.addEventListener('change', (e) => {
            const file = e.target.files[0];
            if (file) {
                this.handleFile(file);
            }
        });

        if (this.removeBtn) {
            this.removeBtn.addEventListener('click', () => {
                this.removeImage();
            });
        }

        this.preview.addEventListener('dragover', (e) => {
            e.preventDefault();
            this.preview.classList.add('drag-over');
        });

        this.preview.addEventListener('dragleave', () => {
            this.preview.classList.remove('drag-over');
        });

        this.preview.addEventListener('drop', (e) => {
            e.preventDefault();
            this.preview.classList.remove('drag-over');
            
            const file = e.dataTransfer.files[0];
            if (file) {
                this.handleFile(file);
            }
        });
    }

    async handleFile(file) {
        this.hideError();
        this.hideSuccess();
        
        const validationError = this.validateFile(file);
        if (validationError) {
            this.showError(validationError);
            return;
        }
        
        await this.previewFile(file);
    }

    validateFile(file) {
        if (!this.allowedTypes.includes(file.type)) {
            return 'Only JPEG, PNG, GIF, and WebP images are allowed';
        }
        
        if (file.size > this.maxSize) {
            return 'Image file size cannot exceed 5MB';
        }
        
        return null;
    }

    async previewFile(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();
            reader.onload = (e) => {
                const img = new Image();
                img.onload = () => {
                    if (img.width > this.maxDimensions || img.height > this.maxDimensions) {
                        this.showError(`Image dimensions cannot exceed ${this.maxDimensions}x${this.maxDimensions} pixels`);
                        reject();
                        return;
                    }
                    
                    this.updatePreview(e.target.result);
                    this.uploadFile(file);
                    resolve();
                };
                img.onerror = () => {
                    this.showError('Failed to load image');
                    reject();
                };
                img.src = e.target.result;
            };
            reader.onerror = () => {
                this.showError('Failed to read file');
                reject();
            };
            reader.readAsDataURL(file);
        });
    }

    updatePreview(dataURL) {
        if (this.previewImage) {
            this.previewImage.src = dataURL;
        } else {
            this.preview.innerHTML = '';
            
            const imgElement = document.createElement('img');
            imgElement.id = 'preview-image';
            imgElement.className = 'preview-image';
            imgElement.src = dataURL;
            imgElement.alt = 'Custom header image';
            this.preview.appendChild(imgElement);
            this.previewImage = imgElement;
            
            const removeBtn = document.createElement('button');
            removeBtn.type = 'button';
            removeBtn.className = 'btn-remove-image';
            removeBtn.id = 'remove-image-btn';
            removeBtn.setAttribute('aria-label', 'Remove custom image');
            removeBtn.innerHTML = '<span aria-hidden="true">×</span><span class="sr-only">Remove Image</span>';
            removeBtn.addEventListener('click', () => this.removeImage());
            this.preview.appendChild(removeBtn);
            this.removeBtn = removeBtn;
        }
    }

    async uploadFile(file) {
        const eventID = document.getElementById('event-id')?.value;
        if (!eventID) {
            this.showError('Event ID not found');
            return;
        }
        
        this.showProgress();
        
        const formData = new FormData();
        formData.append('file', file);
        
        const csrfToken = document.querySelector('[name="csrf_token"]')?.value;
        
        try {
            const response = await fetch(`/api/events/${eventID}/images`, {
                method: 'POST',
                body: formData,
                headers: {
                    'X-CSRF-Token': csrfToken || '',
                    'Accept': 'application/json'
                }
            });
            
            if (!response.ok) {
                const error = await response.json();
                throw new Error(error.error?.message || 'Upload failed');
            }
            
            const result = await response.json();
            
            if (this.hiddenInput && result.image?.public_url) {
                this.hiddenInput.value = result.image.public_url;
            }
            
            this.hideProgress();
            this.showSuccess('Image uploaded successfully');
            
        } catch (error) {
            this.hideProgress();
            this.showError(error.message);
        }
    }

    removeImage() {
        this.preview.innerHTML = `
            <div class="image-placeholder">
                <svg class="placeholder-icon" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
                    <rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect>
                    <circle cx="8.5" cy="8.5" r="1.5"></circle>
                    <polyline points="21 15 16 10 5 21"></polyline>
                </svg>
                <p class="placeholder-text">No custom image</p>
            </div>
        `;
        
        if (this.hiddenInput) {
            this.hiddenInput.value = '';
        }
        
        this.input.value = '';
        this.previewImage = null;
        this.removeBtn = null;
        
        this.hideError();
        this.hideSuccess();
    }

    showProgress() {
        if (this.progress) {
            this.progress.hidden = false;
            if (this.progressFill) {
                this.progressFill.style.width = '50%';
            }
        }
        this.hideError();
        this.hideSuccess();
    }

    hideProgress() {
        if (this.progress) {
            this.progress.hidden = true;
            if (this.progressFill) {
                this.progressFill.style.width = '0%';
            }
        }
    }

    showError(message) {
        if (this.errorDiv && this.errorMessage) {
            this.errorMessage.textContent = message;
            this.errorDiv.hidden = false;
        }
    }

    hideError() {
        if (this.errorDiv) {
            this.errorDiv.hidden = true;
        }
    }

    showSuccess(message) {
        if (this.successDiv) {
            const messageEl = this.successDiv.querySelector('.success-message');
            if (messageEl) {
                messageEl.textContent = message;
            }
            this.successDiv.hidden = false;
            
            setTimeout(() => {
                this.hideSuccess();
            }, 3000);
        }
    }

    hideSuccess() {
        if (this.successDiv) {
            this.successDiv.hidden = true;
        }
    }
}

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        if (document.getElementById('image-upload-input')) {
            new ImageUploader();
        }
    });
} else {
    if (document.getElementById('image-upload-input')) {
        new ImageUploader();
    }
}
