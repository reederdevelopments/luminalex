
import van from "https://cdn.jsdelivr.net/gh/vanjs-org/van/public/van-1.5.1.min.js";

const container = document.getElementById("survey-form-container");
if (container) {
    const surveyData = JSON.parse(container.dataset.survey || '{}');
    
    const { form, div, label, input, textarea, button, select, option, a, svg, path, span } = van.tags;

    const TrashIcon = () => svg({
        xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24",
        "stroke-width": "1.5", stroke: "currentColor", class: "w-6 h-6"
    },
        path({
            "stroke-linecap": "round", "stroke-linejoin": "round",
            d: "m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
        })
    );

    const XIcon = () => svg({
        xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24",
        "stroke-width": "1.5", stroke: "currentColor", class: "w-5 h-5"
    }, path({ "stroke-linecap": "round", "stroke-linejoin": "round", d: "M6 18L18 6M6 6l12 12" }));
    
    const name = van.state(surveyData.Name || "");
    const description = van.state(surveyData.Description || "");
    const isEnabled = van.state(surveyData.IsEnabled ?? true);
    const allowMultipleSubmissions = van.state(surveyData.AllowMultipleSubmissions || false);
    const questions = van.state(surveyData.Questions || []);

    const Question = ({ id, text, type, options, is_required, group_number }, index) => {
        const localText = van.state(text);
        const localType = van.state(type);
        const localOptions = van.state(options || []);
        const localIsRequired = van.state(is_required || false);
        const localGroupNumber = van.state(group_number || 1);
        
        const removeQuestion = () => {
            const newQuestions = [...questions.val];
            newQuestions.splice(index, 1);
            questions.val = newQuestions;
        };

        const addOption = () => {
            localOptions.val = [...localOptions.val, ""];
        };
        
        const removeOption = (optIndex) => {
            const newOptions = [...localOptions.val];
            newOptions.splice(optIndex, 1);
            localOptions.val = newOptions;
        };

        const updateOption = (optIndex, value) => {
            const newOptions = [...localOptions.val];
            newOptions[optIndex] = value;
            localOptions.val = newOptions;
        };

        return div({ class: "bg-white p-4 rounded-lg border border-gray-200 mt-4" },
            input({ type: "hidden", name: `questions[${index}].ID`, value: id || "" }),
            div({ class: "flex justify-between items-start gap-4" },
                textarea({
                    name: `questions[${index}].Text`,
                    class: "w-full border-gray-300 rounded-md shadow-sm focus:border-green-x-light focus:ring-green-x-light sm:text-sm text-lg font-medium",
                    rows: 2,
                    placeholder: `Question ${index + 1}`,
                    value: localText,
                    oninput: e => localText.val = e.target.value
                }),
                div({ class: "flex flex-col gap-2 items-end" },
                    select({
                        name: `questions[${index}].Type`,
                        class: "rounded-md border-gray-300 shadow-sm focus:border-green-x-light focus:ring-green-x-light sm:text-sm",
                        oninput: e => localType.val = e.target.value,
                    },
                        option({ value: "text", selected: localType.val === 'text' }, "Short Answer"),
                        option({ value: "textarea", selected: localType.val === 'textarea' }, "Paragraph"),
                        option({ value: "multiple-choice", selected: localType.val === 'multiple-choice' }, "Multiple Choice"),
                        option({ value: "checkbox", selected: localType.val === 'checkbox' }, "Checkboxes"),
                        option({ value: "dropdown", selected: localType.val === 'dropdown' }, "Dropdown"),
                    ),
                    button({ 
                        type: "button", 
                        onclick: removeQuestion, 
                        class: "p-1 rounded text-gray-500 hover:bg-red-100 hover:text-red-700 transition-colors" 
                    }, TrashIcon())
                )
            ),
            () => ["multiple-choice", "checkbox", "dropdown"].includes(localType.val) ? div({ class: "mt-4 space-y-2 pl-2" },
                () => div(localOptions.val.map((opt, i) => div({ class: "flex items-center gap-2" },
                    input({
                        type: "text",
                        name: `questions[${index}].Options`,
                        class: "w-full border-gray-300 rounded-md shadow-sm sm:text-sm",
                        placeholder: `Option ${i + 1}`,
                        value: opt,
                        oninput: e => updateOption(i, e.target.value),
                    }),
                    button({ type: "button", onclick: () => removeOption(i), class: "text-red-400 hover:text-red-600" }, XIcon())
                ))),
                button({ type: "button", onclick: addOption, class: "text-sm text-green-x-dark hover:underline mt-2" }, "Add option")
            ) : div(),
            div({ class: "mt-4 grid grid-cols-1 sm:grid-cols-2 gap-4" },
                div(
                    label({class: "block text-sm font-medium text-gray-700"}, "Group Number"),
                    input({
                        type: "number",
                        name: `questions[${index}].GroupNumber`,
                        min: "1",
                        class: "mt-1 block w-full rounded-md border-gray-300 shadow-sm sm:text-sm",
                        value: localGroupNumber,
                        oninput: e => localGroupNumber.val = e.target.value,
                    })
                )
            ),
            div({ class: "mt-6 pt-4 border-t border-gray-200 flex justify-between items-center" },
                div({ class: "flex items-center gap-2" },
                    input({
                        type: "checkbox",
                        name: `questions[${index}].IsRequired`,
                        value: "true",
                        class: "h-4 w-4 rounded border-gray-300 text-green-x-light",
                        checked: localIsRequired,
                        onchange: e => localIsRequired.val = e.target.checked
                    }),
                    label({ class: "text-sm text-gray-700" }, "Required")
                )
            )
        );
    };
    
    const addQuestion = () => {
        questions.val = [...questions.val, {
            id: `q_${Date.now()}`,
            text: "",
            type: "text",
            options: [],
            is_required: false,
            group_number: 1,
        }];
    };
    
    const SurveyForm = () => form({ method: "POST" },
        div({ class: "space-y-6 bg-white p-6 rounded-lg shadow" },
            div(
                label({ for: "name", class: "block text-sm font-medium text-gray-700" }, "Survey Name"),
                input({ type: "text", name: "Name", id: "name", required: true, class: "mt-1 block w-full border-gray-300 rounded-md shadow-sm sm:text-sm", value: name, oninput: e => name.val = e.target.value })
            ),
            div(
                label({ for: "description", class: "block text-sm font-medium text-gray-700" }, "Description"),
                textarea({ name: "Description", id: "description", rows: 3, class: "mt-1 block w-full border-gray-300 rounded-md shadow-sm sm:text-sm", oninput: e => description.val = e.target.value }, description.val)
            ),
            div({ class: "flex items-center" },
                input({ type: "checkbox", name: "IsEnabled", id: "is_enabled", value: "true", class: "h-4 w-4 rounded border-gray-300 text-green-x-light", checked: isEnabled, onchange: e => isEnabled.val = e.target.checked }),
                label({ for: "is_enabled", class: "ml-2 block text-sm text-gray-900" }, "Enable this survey for users")
            ),
            div({ class: "flex items-center mt-4" },
                input({ type: "checkbox", name: "AllowMultipleSubmissions", id: "allow_multiple", value: "true", class: "h-4 w-4 rounded border-gray-300 text-green-x-light", checked: allowMultipleSubmissions, onchange: e => allowMultipleSubmissions.val = e.target.checked }),
                label({ for: "allow_multiple", class: "ml-2 block text-sm text-gray-900" }, "Allow multiple submissions")
            )
        ),
        
        () => div({ id: "questions-list" },
            questions.val.map((q, i) => Question(q, i))
        ),

        div({ class: "mt-6" },
            button({ type: "button", onclick: addQuestion, class: "rounded-md bg-gray-200 px-4 py-2 text-sm font-semibold text-gray-800 shadow-sm hover:bg-gray-300" }, "Add Question")
        ),
        
        div({ class: "mt-8 flex justify-end gap-4 border-t pt-6" },
            a({ href: "/admin/surveys", class: "rounded-md bg-white px-4 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50" }, "Cancel"),
            button({ type: "submit", class: "rounded-md bg-green-x-light px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-green-x-dark" }, "Save Survey"),
        )
    );
    
    van.add(container, SurveyForm());
}
