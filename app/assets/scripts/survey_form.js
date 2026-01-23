
document.addEventListener("DOMContentLoaded", () => {
    const questionsContainer = document.getElementById("questions-container");
    const groupHeadingsContainer = document.getElementById("group-headings-container");

    if (!questionsContainer) {
        return;
    }

    // Function to re-index all question blocks to ensure contiguous indices (0, 1, 2, ...)
    function reindexQuestions() {
        const questionBlocks = questionsContainer.querySelectorAll(".question-block");
        questionBlocks.forEach((questionBlock, questionIndex) => {
            // Update all input/select/textarea names that use the array index
            const inputs = questionBlock.querySelectorAll('[name^="questions["]');
            inputs.forEach(input => {
                input.name = input.name.replace(/questions\[\d+\]/, `questions[${questionIndex}]`);
            });

            // Update IDs and `for` attributes that depend on the index
            const idElements = questionBlock.querySelectorAll('[id^="is_required_"], [id^="question_type_"], [id^="question_text_"], [id^="question_group_"]');
            idElements.forEach(el => {
                el.id = el.id.replace(/_\d+$/, `_${questionIndex}`);
            });

            const labelElements = questionBlock.querySelectorAll('[for^="is_required_"], [for^="question_type_"], [for^="question_text_"], [for^="question_group_"]');
            labelElements.forEach(el => {
                el.setAttribute('for', el.getAttribute('for').replace(/_\d+$/, `_${questionIndex}`));
            });


            // Update the data-question-index for the options list
            const optionsList = questionBlock.querySelector('.options-list');
            if (optionsList) {
                optionsList.dataset.questionIndex = questionIndex;
            }

            // Update option placeholder text
            const optionInputs = questionBlock.querySelectorAll('.option-input input[type="text"]');
            optionInputs.forEach((input, optionIndex) => {
                input.placeholder = `Option ${optionIndex + 1}`;
            });
        });
    }

    function reindexGroupHeadings() {
        if (!groupHeadingsContainer) return;
        const headingBlocks = groupHeadingsContainer.querySelectorAll(".group-heading-block");
        headingBlocks.forEach((block, index) => {
            const label = block.querySelector('label');
            const textarea = block.querySelector('textarea');

            if (label) {
                label.setAttribute('for', `group_heading_${index}`);
                label.textContent = `Group ${index + 1} Heading`;
            }

            if (textarea) {
                textarea.id = `group_heading_${index}`;
            }
        });
    }

    // Function to add a new option input to a question
    function addOption(optionsDiv) {
        const questionIndex = optionsDiv.dataset.questionIndex;
        const optionIndex = optionsDiv.querySelectorAll(".option-input").length;

        const newOption = document.createElement("div");
        newOption.className = "flex items-center gap-2 mt-2 option-input";
        newOption.innerHTML = `
            <input type="text" name="questions[${questionIndex}].Options" class="block w-full rounded-md border-gray-300 shadow-sm focus:border-green-x-dark focus:ring-green-x-dark sm:text-sm" placeholder="Option ${optionIndex + 1}" value="">
            <button type="button" class="remove-option-btn text-red-x-dark hover:text-red-x-light">
                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 pointer-events-none">
                    <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                </svg>
            </button>
        `;
        optionsDiv.appendChild(newOption);
        newOption.querySelector('input').focus();
    }

    // Function to show/hide options based on question type
    function toggleOptions(questionDiv) {
        const typeSelector = questionDiv.querySelector(".question-type-selector");
        const optionsContainer = questionDiv.querySelector(".options-container");

        if (!typeSelector || !optionsContainer) {
            return;
        }

        const selectedType = typeSelector.value;
        const hasOptions = ["multiple-choice", "checkbox", "dropdown"].includes(selectedType);

        if (hasOptions) {
            optionsContainer.classList.remove("hidden");
        } else {
            optionsContainer.classList.add("hidden");
        }
    }

    // Event delegation for all dynamic elements
    document.body.addEventListener("click", (e) => {
        // Handle remove question
        const removeQuestionBtn = e.target.closest(".remove-question-btn");
        if (removeQuestionBtn) {
            removeQuestionBtn.closest(".question-block").remove();
            reindexQuestions();
            return;
        }

        // Handle remove group heading
        const removeGroupHeadingBtn = e.target.closest(".remove-group-heading-btn");
        if (removeGroupHeadingBtn) {
            removeGroupHeadingBtn.closest(".group-heading-block").remove();
            reindexGroupHeadings();
            return;
        }


        // Handle add option
        const addOptionBtn = e.target.closest(".add-option-btn");
        if (addOptionBtn) {
            const optionsDiv = addOptionBtn.closest(".options-container").querySelector(".options-list");
            addOption(optionsDiv);
            return;
        }

        // Handle remove option
        const removeOptionBtn = e.target.closest(".remove-option-btn");
        if (removeOptionBtn) {
            const optionsDiv = removeOptionBtn.closest('.options-list');
            removeOptionBtn.closest(".option-input").remove();
            
            // Re-index placeholder text for remaining options
            const optionInputs = optionsDiv.querySelectorAll('.option-input input[type="text"]');
            optionInputs.forEach((input, optionIndex) => {
                input.placeholder = `Option ${optionIndex + 1}`;
            });
        }
    });

    questionsContainer.addEventListener("change", (e) => {
        // Handle question type change
        if (e.target.classList.contains("question-type-selector")) {
            const questionDiv = e.target.closest(".question-block");
            toggleOptions(questionDiv);
        }
    });

    // Initial setup for any questions that are already on the page (e.g., on edit)
    document.querySelectorAll(".question-block").forEach(toggleOptions);

    // After htmx adds a new question, run toggleOptions on it
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        if (evt.detail.target.id === 'questions-container') {
            const newQuestion = evt.detail.target.lastElementChild
            if (newQuestion && newQuestion.classList.contains('question-block')) {
                toggleOptions(newQuestion);
                // focus the new question's text input
                const textInput = newQuestion.querySelector('input[type="text"]');
                if (textInput) {
                    textInput.focus();
                }
            }
        }
    });

    // Before HTMX sends a request for a new question partial, set the correct index
    document.body.addEventListener('htmx:configRequest', function(evt) {
        const addQuestionBtn = document.getElementById('add-question-btn');
        if (addQuestionBtn && evt.detail.elt === addQuestionBtn) {
            const questionIndex = questionsContainer.querySelectorAll(".question-block").length;
            // modify the path to include the new index
            evt.detail.path = evt.detail.path.split('?')[0] + `?index=${questionIndex}`;
        }
        
        const addGroupHeadingBtn = document.getElementById('add-group-heading-btn');
        if (addGroupHeadingBtn && evt.detail.elt === addGroupHeadingBtn) {
            if (groupHeadingsContainer) {
                const headingIndex = groupHeadingsContainer.querySelectorAll(".group-heading-block").length;
                evt.detail.path = evt.detail.path.split('?')[0] + `?index=${headingIndex}`;
            }
        }
    });
});
