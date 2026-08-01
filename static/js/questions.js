(function() {
    'use strict';

    var container = document.getElementById('questions-container');
    var addBtn = document.getElementById('add-question-btn');
    if (!container || !addBtn) return;

    function renumber() {
        var items = container.querySelectorAll('.question-item');
        items.forEach(function(item, index) {
            item.dataset.index = index;
            var label = item.querySelector('.form-label');
            if (label) {
                label.textContent = 'Question ' + (index + 1);
            }
            item.querySelectorAll('input, select, textarea').forEach(function(field) {
                var name = field.getAttribute('name');
                var id = field.getAttribute('id');
                if (name) field.setAttribute('name', name.replace(/questions\[\d+\]/, 'questions[' + index + ']'));
                if (id) field.setAttribute('id', id.replace(/question_\d+_/, 'question_' + index + '_'));
            });
            item.querySelector('.remove-question').dataset.index = index;
        });
    }

    function toggleOptions(item) {
        var type = item.querySelector('select[name$="[type]"]');
        var optionsField = item.querySelector('.question-options-field');
        if (!type || !optionsField) return;
        optionsField.style.display = (type.value === 'single_choice' || type.value === 'multiple_choice') ? 'block' : 'none';
    }

    function createItem(index) {
        var item = document.createElement('div');
        item.className = 'question-item';
        item.dataset.index = index;
        item.innerHTML =
            '<input type="hidden" name="questions[' + index + '][id]" value="">' +
            '<div class="form-group">' +
                '<label for="question_' + index + '_text" class="form-label">Question ' + (index + 1) + '</label>' +
                '<input type="text" id="question_' + index + '_text" name="questions[' + index + '][text]" class="form-input" placeholder="e.g., Do you have any dietary restrictions?">' +
            '</div>' +
            '<div class="form-group">' +
                '<label for="question_' + index + '_type" class="form-label">Answer Type</label>' +
                '<select id="question_' + index + '_type" name="questions[' + index + '][type]" class="form-select">' +
                    '<option value="text">Text</option>' +
                    '<option value="single_choice">Single Choice</option>' +
                    '<option value="multiple_choice">Multiple Choice</option>' +
                '</select>' +
            '</div>' +
            '<div class="form-group question-options-field" style="display:none;">' +
                '<label for="question_' + index + '_options" class="form-label">Options (one per line)</label>' +
                '<textarea id="question_' + index + '_options" name="questions[' + index + '][options]" class="form-textarea" rows="3"></textarea>' +
            '</div>' +
            '<label class="form-checkbox-label">' +
                '<input type="checkbox" name="questions[' + index + '][required]" class="form-checkbox">' +
                '<span>Required</span>' +
            '</label>' +
            '<button type="button" class="btn btn-danger btn-sm remove-question" data-index="' + index + '">Remove</button>';

        var typeSelect = item.querySelector('select[name$="[type]"]');
        typeSelect.addEventListener('change', function() { toggleOptions(item); });

        var removeBtn = item.querySelector('.remove-question');
        removeBtn.addEventListener('click', function() {
            item.parentNode.removeChild(item);
            renumber();
        });

        return item;
    }

    container.querySelectorAll('.question-item').forEach(function(item) {
        var typeSelect = item.querySelector('select[name$="[type]"]');
        if (typeSelect) typeSelect.addEventListener('change', function() { toggleOptions(item); });
        var removeBtn = item.querySelector('.remove-question');
        if (removeBtn) removeBtn.addEventListener('click', function() {
            item.parentNode.removeChild(item);
            renumber();
        });
        toggleOptions(item);
    });

    addBtn.addEventListener('click', function() {
        var index = container.querySelectorAll('.question-item').length;
        container.appendChild(createItem(index));
    });
})();
