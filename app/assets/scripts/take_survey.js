
document.addEventListener("DOMContentLoaded", () => {
    const surveyForm = document.querySelector('form[action^="/surveys/"]');
    if (!surveyForm) {
        return;
    }

    // Extract surveyId from the form's action and hidden input
    const pathParts = new URL(surveyForm.action).pathname.split('/');
    const surveyId = pathParts[pathParts.length - 1];

    const assignmentIdInput = surveyForm.querySelector('input[name="assignment_id"]');
    if (!surveyId || !assignmentIdInput) {
        console.error("Could not determine survey or assignment ID for progress saving.");
        // We can continue for progress bar functionality
    }
    const assignmentId = assignmentIdInput ? assignmentIdInput.value : null;

    let debounceTimer;

    const saveProgress = () => {
        if (!assignmentId) return; // Can't save progress without assignmentId

        const formData = new FormData(surveyForm);
        const data = {};
        
        // This logic correctly handles single, multiple (checkboxes), and grouped (radio) inputs
        for (const [key, value] of formData.entries()) {
            // We are only interested in answers, not assignment_id
            if (key === 'assignment_id') continue;

            if (data.hasOwnProperty(key)) {
                if (!Array.isArray(data[key])) {
                    data[key] = [data[key]];
                }
                data[key].push(value);
            } else {
                data[key] = value;
            }
        }

        // We need to package this into the structure the backend will expect
        const payload = {
            AssignmentID: assignmentId,
            Answers: data
        };
        
        fetch(`/surveys/${surveyId}/progress`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json',
            },
            body: JSON.stringify(payload),
        })
        .then(response => {
            if (response.ok) {
                console.log('Progress saved.');
                showSaveIndicator();
            } else {
                console.error('Failed to save progress.');
            }
        })
        .catch(error => {
            console.error('Error saving progress:', error);
        });
    };

    const showSaveIndicator = () => {
        let indicator = document.getElementById('save-indicator');
        if (!indicator) {
            indicator = document.createElement('div');
            indicator.id = 'save-indicator';
            indicator.className = 'fixed bottom-5 right-5 bg-green-x-dark text-white px-4 py-2 rounded-md shadow-lg transition-opacity duration-500 opacity-0';
            document.body.appendChild(indicator);
        }
        
        indicator.textContent = 'Progress saved';
        
        // Use requestAnimationFrame to ensure the initial opacity-0 is applied before transitioning
        requestAnimationFrame(() => {
            indicator.style.opacity = '1';
        });

        setTimeout(() => {
            indicator.style.opacity = '0';
        }, 2000);
    };
    
    // --- PROGRESS BAR & AUTOSAVE LOGIC ---

    const progressBar = document.getElementById('progress-bar');
    const progressText = document.getElementById('progress-text');
    const hasProgressBar = progressBar && progressText;

    let totalUnits = 0;
    const questionConfigs = [];

    if (hasProgressBar) {
        const questionElements = surveyForm.querySelectorAll('.question-block');
        questionElements.forEach(questionEl => {
            const questionType = questionEl.dataset.questionType;
            let units = 0;
            if (questionType === 'multi-grid-radio') {
                units = questionEl.querySelectorAll('tbody tr').length;
            } else if (questionType) {
                units = 1;
            }
            if (units > 0) {
                totalUnits += units;
                questionConfigs.push({ element: questionEl, type: questionType });
            }
        });
    }

    const updateProgress = () => {
        if (!hasProgressBar) return;
        
        let answeredUnits = 0;
        questionConfigs.forEach(config => {
            const questionEl = config.element;
            const type = config.type;

            if (type === 'multi-grid-radio') {
                const answeredRows = new Set();
                questionEl.querySelectorAll('input[type="radio"]:checked').forEach(radio => {
                    answeredRows.add(radio.name);
                });
                answeredUnits += answeredRows.size;
            } else {
                const inputs = questionEl.querySelectorAll('input:not([type=hidden]), textarea, select');
                let isAnswered = false;
                if (inputs.length > 0) {
                    const firstInput = inputs[0];
                    if (firstInput.type === 'checkbox' || firstInput.type === 'radio') {
                         isAnswered = Array.from(inputs).some(input => input.checked);
                    } else if (firstInput.tagName.toLowerCase() === 'select') {
                        isAnswered = firstInput.value !== '';
                    } else { // text, textarea
                         isAnswered = firstInput.value.trim() !== '';
                    }
                }
                if (isAnswered) {
                    answeredUnits += 1;
                }
            }
        });

        const percentage = totalUnits > 0 ? Math.round((answeredUnits / totalUnits) * 100) : 0;
        progressBar.style.width = `${percentage}%`;
        progressText.textContent = `${percentage}%`;
    };

    surveyForm.addEventListener('input', () => {
        // Debounce save progress
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(saveProgress, 1500);
        
        // Update progress bar immediately
        updateProgress();
    });

    // Initial calculation on page load
    updateProgress();
});
