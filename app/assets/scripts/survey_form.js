
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
            const idElements = questionBlock.querySelectorAll('[id^="is_required_"], [id^="question_type_"], [id^="question_text_"], [id^="hidden_question_text_"], [id^="question_group_"]');
            idElements.forEach(el => {
                el.id = el.id.replace(/_\d+$/, `_${questionIndex}`);
            });

            const editorDiv = questionBlock.querySelector('.wysiwyg-editor');
            if (editorDiv) {
                editorDiv.dataset.hiddenInputId = `hidden_question_text_${questionIndex}`;
            }

            const labelElements = questionBlock.querySelectorAll('[for^="is_required_"], [for^="question_type_"], [for^="question_text_"], [for^="question_group_"]');
            labelElements.forEach(el => {
                el.setAttribute('for', el.getAttribute('for').replace(/_\d+$/, `_${questionIndex}`));
            });


            // Update the data-question-index for the options and rows list
            const optionsList = questionBlock.querySelector('.options-list');
            if (optionsList) {
                optionsList.dataset.questionIndex = questionIndex;
            }
            const rowsList = questionBlock.querySelector('.rows-list');
            if (rowsList) {
                rowsList.dataset.questionIndex = questionIndex;
                 // Re-index rows within the question
                rowsList.querySelectorAll('.row-input').forEach((row, rowIndex) => {
                    const hiddenInput = row.querySelector('input[type="hidden"]');
                    if (hiddenInput) {
                        hiddenInput.name = `questions[${questionIndex}].Rows`;
                        hiddenInput.id = `hidden_row_${questionIndex}_${rowIndex}`;
                    }
                    const editor = row.querySelector('.wysiwyg-editor');
                    if (editor) {
                        editor.dataset.hiddenInputId = `hidden_row_${questionIndex}_${rowIndex}`;
                        editor.setAttribute('placeholder', `Row ${rowIndex + 1}`);
                    }
                });
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
            const editor = block.querySelector('.wysiwyg-editor');
            const hiddenInput = block.querySelector('input[type="hidden"]');

            if (label) {
                label.setAttribute('for', `group_heading_editor_${index}`);
                label.textContent = `Group ${index + 1} Heading`;
            }
            
            if (hiddenInput) {
                hiddenInput.id = `hidden_group_heading_${index}`;
            }

            if (editor) {
                editor.id = `group_heading_editor_${index}`;
                editor.dataset.hiddenInputId = `hidden_group_heading_${index}`;
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

    // Function to add a new row input to a question
    function addRow(rowsDiv) {
        const questionIndex = rowsDiv.dataset.questionIndex;
        const rowIndex = rowsDiv.querySelectorAll(".row-input").length;

        const newRow = document.createElement("div");
        newRow.className = "row-input";
        newRow.innerHTML = `
            <div class="flex items-start gap-2">
                <div class="w-full">
                    <div class="editor-toolbar flex items-center space-x-2 p-1 bg-gray-100 rounded-t-md border border-b-0 border-gray-300">
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="bold" title="Bold"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path fill-rule="evenodd" d="M8.21 13c2.106 0 3.412-1.087 3.412-2.823 0-1.306-.984-2.283-2.324-2.386v-.055a2.176 2.176 0 0 0 1.852-2.14c0-1.51-1.162-2.46-3.014-2.46H3.843V13zM5.908 4.674h1.696c.963 0 1.517.451 1.517 1.244 0 .834-.629 1.32-1.73 1.32H5.908V4.673zm0 6.788V8.598h1.73c1.217 0 1.88.492 1.88 1.415 0 .943-.643 1.449-1.832 1.449H5.907z" clip-rule="evenodd"></path></svg></button>
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="italic" title="Italic"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path fill-rule="evenodd" d="M7.991 11.674 9.53 4.455c.123-.595.246-.71 1.347-.807l.11-.52H7.211l-.11.52c1.06.096 1.128.212 1.005.807L6.57 11.674c-.123.595-.246.71-1.346.806l-.11.52h3.774l.11-.52c-1.06-.095-1.129-.211-1.006-.806z" clip-rule="evenodd"></path></svg></button>
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="underline" title="Underline"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path d="M5 4.5a.5.5 0 0 0-.5.5v6a4.5 4.5 0 0 0 4.5 4.5h2a4.5 4.5 0 0 0 4.5-4.5v-6a.5.5 0 0 0-.5-.5h-10ZM13 11a3.5 3.5 0 0 1-3.5 3.5h-2A3.5 3.5 0 0 1 4 11V5h9v6Z"></path><path d="M3 15.5a.5.5 0 0 1 .5-.5h13a.5.5 0 0 1 0 1h-13a.5.5 0 0 1-.5-.5Z"></path></svg></button>
						<div class="h-4 border-l border-gray-300 mx-1"></div>
						<select class="editor-font-size-selector bg-transparent border-none rounded-md text-sm p-1 hover:bg-gray-200 focus:ring-0" title="Font Size">
							<option value="" disabled selected>Size</option>
							<option value="1">Tiny</option>
							<option value="2">Small</option>
							<option value="3">Normal</option>
							<option value="4">Large</option>
							<option value="5">XL</option>
							<option value="6">XXL</option>
						</select>
						<div class="h-4 border-l border-gray-300 mx-1"></div>
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="createLink" title="Link"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path d="M6.354 5.5H4a3 3 0 0 0 0 6h3a3 3 0 0 0 2.83-4H9q-.13 0-.25.031A2 2 0 0 1 7 10.5H4a2 2 0 1 1 0-4h1.535c.218-.376.495-.714.82-1z"></path><path d="M9 5.5a3 3 0 0 0-2.83 4h1.098A2 2 0 0 1 9 6.5h3a2 2 0 1 1 0 4h-1.535a4 4 0 0 1-.82 1H12a3 3 0 1 0 0-6z"></path></svg></button>
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="insertUnorderedList" title="Unordered List"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path fill-rule="evenodd" d="M3.75 6.75a.75.75 0 000 1.5h12.5a.75.75 0 000-1.5H3.75zM3.75 11.25a.75.75 0 000 1.5h12.5a.75.75 0 000-1.5H3.75zM3.75 15.75a.75.75 0 000 1.5h12.5a.75.75 0 000-1.5H3.75z" clip-rule="evenodd"></path></svg></button>
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="insertOrderedList" title="Ordered List"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path fill-rule="evenodd" d="M3.75 6.75a.75.75 0 000 1.5h12.5a.75.75 0 000-1.5H3.75zM3.75 11.25a.75.75 0 000 1.5h12.5a.75.75 0 000-1.5H3.75zM3.75 15.75a.75.75 0 000 1.5h12.5a.75.75 0 000-1.5H3.75z" clip-rule="evenodd"></path></svg></button>
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="formatBlock" data-value="blockquote" title="Quote"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path d="M14 1a1 1 0 0 1 1 1v12a1 1 0 0 1-1 1H2a1 1 0 0 1-1-1V2a1 1 0 0 1 1-1zM2 0a2 2 0 0 0-2 2v12a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V2a2 2 0 0 0-2-2z"></path><path d="M6.854 4.646a.5.5 0 0 1 0 .708L4.207 8l2.647 2.646a.5.5 0 0 1-.708.708l-3-3a.5.5 0 0 1 0-.708l3-3a.5.5 0 0 1 .708 0m2.292 0a.5.5 0 0 0 0 .708L11.793 8l-2.647 2.646a.5.5 0 0 0 .708.708l3-3a.5.5 0 0 0 0-.708l-3-3a.5.5 0 0 0-.708 0"></path></svg></button>
						<button type="button" class="editor-btn p-1 rounded hover:bg-gray-200" data-command="wrapWithCode" title="Code"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor" class="w-5 h-5 pointer-events-none"><path fill-rule="evenodd" d="M6.28 5.22a.75.75 0 0 1 0 1.06L2.56 10l3.72 3.72a.75.75 0 0 1-1.06 1.06L.97 10.53a.75.75 0 0 1 0-1.06l4.25-4.25a.75.75 0 0 1 1.06 0Zm7.44 0a.75.75 0 0 1 1.06 0l4.25 4.25a.75.75 0 0 1 0 1.06l-4.25 4.25a.75.75 0 0 1-1.06-1.06L17.44 10l-3.72-3.72a.75.75 0 0 1 0-1.06ZM11.378 2.47a.75.75 0 0 1 .848.353l3.25 6.5a.75.75 0 0 1-1.352.673l-3.25-6.5a.75.75 0 0 1 .504-1.026ZM8.622 17.53a.75.75 0 0 1-.848-.353l-3.25-6.5a.75.75 0 0 1 1.352-.673l3.25 6.5a.75.75 0 0 1-.504 1.026Z" clip-rule="evenodd"></path></svg></button>
                    </div>
                    <input type="hidden" name="questions[${questionIndex}].Rows" id="hidden_row_${questionIndex}_${rowIndex}" value=""/>
                    <div
                        class="wysiwyg-editor block w-full rounded-b-md border-t-0 border border-gray-300 shadow-sm focus:outline-none focus:border-green-x-dark focus:ring-1 focus:ring-green-x-dark sm:text-sm p-2 prose max-w-none"
                        contenteditable="true"
                        style="min-height: 2.5rem;"
                        data-hidden-input-id="hidden_row_${questionIndex}_${rowIndex}"
                        placeholder="Row ${rowIndex + 1}">
                    </div>
                </div>
                <button type="button" class="remove-row-btn text-red-x-dark hover:text-red-x-light mt-2">
                    <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-5 h-5 pointer-events-none">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" />
                    </svg>
                </button>
            </div>
        `;
        rowsDiv.appendChild(newRow);
        newRow.querySelector('.wysiwyg-editor').focus();
    }

    // Function to show/hide options and rows based on question type
    function toggleOptionsAndRows(questionDiv) {
        const typeSelector = questionDiv.querySelector(".question-type-selector");
        const optionsContainer = questionDiv.querySelector(".options-container");
        const rowsContainer = questionDiv.querySelector(".rows-container");

        if (!typeSelector) {
            return;
        }

        const selectedType = typeSelector.value;
        const hasOptions = ["multiple-choice", "checkbox", "dropdown", "multi-grid-radio"].includes(selectedType);
        const hasRows = ["multi-grid-radio"].includes(selectedType);

        if (optionsContainer) {
            if (hasOptions) {
                optionsContainer.classList.remove("hidden");
            } else {
                optionsContainer.classList.add("hidden");
            }
        }

        if (rowsContainer) {
            if (hasRows) {
                rowsContainer.classList.remove("hidden");
            } else {
                rowsContainer.classList.add("hidden");
            }
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

        // Handle move question up
        const moveUpBtn = e.target.closest(".move-question-up-btn");
        if (moveUpBtn) {
            const questionBlock = moveUpBtn.closest(".question-block");
            if (questionBlock && questionBlock.previousElementSibling) {
                questionBlock.previousElementSibling.before(questionBlock);
                reindexQuestions();
            }
            return;
        }

        // Handle move question down
        const moveDownBtn = e.target.closest(".move-question-down-btn");
        if (moveDownBtn) {
            const questionBlock = moveDownBtn.closest(".question-block");
            if (questionBlock && questionBlock.nextElementSibling) {
                questionBlock.nextElementSibling.after(questionBlock);
                reindexQuestions();
            }
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

        // Handle add row
        const addRowBtn = e.target.closest(".add-row-btn");
        if (addRowBtn) {
            const rowsDiv = addRowBtn.closest(".rows-container").querySelector(".rows-list");
            addRow(rowsDiv);
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
        
        // Handle remove row
        const removeRowBtn = e.target.closest(".remove-row-btn");
        if (removeRowBtn) {
            const questionBlock = removeRowBtn.closest('.question-block');
            const questionIndex = questionBlock.dataset.questionIndex;
            const rowsDiv = removeRowBtn.closest('.rows-list');
            removeRowBtn.closest(".row-input").remove();
            
            // Re-index remaining rows for this question
            rowsDiv.querySelectorAll('.row-input').forEach((row, rowIndex) => {
                const hiddenInput = row.querySelector('input[type="hidden"]');
                if (hiddenInput) {
                    hiddenInput.id = `hidden_row_${questionIndex}_${rowIndex}`;
                }
                const editor = row.querySelector('.wysiwyg-editor');
                if (editor) {
                    editor.dataset.hiddenInputId = `hidden_row_${questionIndex}_${rowIndex}`;
                    editor.setAttribute('placeholder', `Row ${rowIndex + 1}`);
                }
            });
        }
    });

    questionsContainer.addEventListener("change", (e) => {
        // Handle question type change
        if (e.target.classList.contains("question-type-selector")) {
            const questionDiv = e.target.closest(".question-block");
            toggleOptionsAndRows(questionDiv);
        }
    });

    // Initial setup for any questions that are already on the page (e.g., on edit)
    document.querySelectorAll(".question-block").forEach(toggleOptionsAndRows);

    // After htmx adds a new question, run toggleOptionsAndRows on it
    document.body.addEventListener('htmx:afterSwap', function(evt) {
        if (evt.detail.target.id === 'questions-container') {
            const newQuestion = evt.detail.target.lastElementChild
            if (newQuestion && newQuestion.classList.contains('question-block')) {
                toggleOptionsAndRows(newQuestion);
                // focus the new question's editor
                const editor = newQuestion.querySelector('.wysiwyg-editor');
                if (editor) {
                    editor.focus();
                }
            }
        }
        if (evt.detail.target.id === 'group-headings-container') {
             const newHeading = evt.detail.target.lastElementChild
             if(newHeading && newHeading.classList.contains('group-heading-block')) {
                 const editor = newHeading.querySelector('.wysiwyg-editor');
                 if (editor) {
                     editor.focus();
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

    // --- Rich Text Editor ---
    function wrapSelectionWithTag(tagName) {
        const selection = window.getSelection();
        if (selection.rangeCount > 0) {
            const range = selection.getRangeAt(0);
            if (range.collapsed) return; // Don't wrap empty selection

            const selectedContents = range.extractContents();
            const wrapper = document.createElement(tagName);
            wrapper.appendChild(selectedContents);
            range.insertNode(wrapper);
        }
    }

    document.body.addEventListener('click', e => {
        const button = e.target.closest('.editor-btn');
        if (!button) return;

        e.preventDefault();

        const toolbar = button.closest('.editor-toolbar');
        if (!toolbar) return;
        
        let editor;
        const parentFlex = toolbar.nextElementSibling;
        if (parentFlex && parentFlex.classList.contains('flex')) {
            // This is for group headings
            editor = parentFlex.querySelector('.wysiwyg-editor');
        } else {
            // This is for question text or rows
            editor = toolbar.nextElementSibling.nextElementSibling;
        }

        if (!editor || !editor.isContentEditable) {
            console.error('Could not find associated contenteditable editor for button.');
            return;
        }

        editor.focus();

        const command = button.dataset.command;
        const value = button.dataset.value;

        if (command === 'createLink') {
            const url = prompt("Enter the URL:");
            if (url) {
                document.execCommand('createLink', false, url);
                const selection = window.getSelection();
                if (selection.rangeCount > 0) {
                    let link = selection.anchorNode.parentElement;
                    while (link && link.tagName !== 'A') {
                        link = link.parentElement;
                    }
                    if (link) {
                        link.setAttribute('target', '_blank');
                        link.setAttribute('rel', 'noopener noreferrer');
                        link.classList.add('text-green-x-dark', 'hover:underline');
                    }
                }
            }
        } else if (command === 'wrapWithCode') {
            wrapSelectionWithTag('code');
        } else {
            document.execCommand(command, false, value);
        }

        // Trigger input event to update hidden field
        editor.dispatchEvent(new Event('input', {
            bubbles: true,
            cancelable: true
        }));
    });

    document.body.addEventListener('change', e => {
        const selector = e.target.closest('.editor-font-size-selector');
        if (!selector) return;

        const value = selector.value;
        if (!value) return;

        const toolbar = selector.closest('.editor-toolbar');
        if (!toolbar) return;
        
        let editor;
        const parentFlex = toolbar.nextElementSibling;
        if (parentFlex && parentFlex.classList.contains('flex')) {
            // This is for group headings
            editor = parentFlex.querySelector('.wysiwyg-editor');
        } else {
            // This is for question text or rows
            editor = toolbar.nextElementSibling.nextElementSibling;
        }

        if (!editor || !editor.isContentEditable) {
            console.error('Could not find associated contenteditable editor for font size selector.');
            return;
        }

        editor.focus();

        document.execCommand('fontSize', false, value);
    
        // Trigger input event to update hidden field
        editor.dispatchEvent(new Event('input', { bubbles: true, cancelable: true }));

        // Reset selector to placeholder
        selector.value = "";
    });

    // Sync contenteditable div content to hidden input
    document.body.addEventListener('input', e => {
        const editor = e.target.closest('.wysiwyg-editor');
        if (editor) {
            const hiddenInputId = editor.dataset.hiddenInputId;
            const hiddenInput = document.getElementById(hiddenInputId);
            if (hiddenInput) {
                hiddenInput.value = editor.innerHTML;
            }
        }
    });
});
