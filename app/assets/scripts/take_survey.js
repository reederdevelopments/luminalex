
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
        return;
    }
    const assignmentId = assignmentIdInput.value;

    let debounceTimer;

    const saveProgress = () => {
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

    surveyForm.addEventListener('input', () => {
        clearTimeout(debounceTimer);
        debounceTimer = setTimeout(saveProgress, 1500); // Save progress 1.5 seconds after input
    });
});
