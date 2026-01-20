import van from "https://cdn.jsdelivr.net/gh/vanjs-org/van/public/van-1.5.1.min.js";

const container = document.getElementById("survey-form-container");
if (container) {
	const surveyData = JSON.parse(container.dataset.survey);
	
	const { form, div, label, input, textarea, button, select, option, a, svg, path } = van.tags;

    const TrashIcon = () => svg({
		xmlns: "http://www.w3.org/2000/svg", fill: "none", viewBox: "0 0 24 24",
		"stroke-width": "1.5", stroke: "currentColor", class: "w-6 h-6"
	},
		path({
			"stroke-linecap": "round", "stroke-linejoin": "round",
			d: "m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0"
		})
	);
	
	const name = van.state(surveyData.Name || "");
	const description = van.state(surveyData.Description || "");
	const isEnabled = van.state(surveyData.IsEnabled ?? true);
	const questions = van.state(surveyData.Questions || []);
	
	const hiddenQuestionsInput = input({ type: "hidden", name: "questions" });
	
	van.derive(() => {
		hiddenQuestionsInput.value = JSON.stringify(questions.val);
	});
	
	const Question = ({ id, text, type, options }, index) => {
		const updateQuestion = (field, value) => {
    const questionToUpdate = questions.val[index];
    questionToUpdate[field] = value;
    if (field === 'type' && (value === 'text' || value === 'textarea')) {
        questionToUpdate.options = [];
    }
    questions.val = questions.val;
};
		
		const updateOption = (optIndex, value) => {
    questions.val[index].options[optIndex] = value;
    // Notify van.js that the state has been mutated
    questions.val = questions.val;
        }

		const addOption = () => {
    const q = questions.val[index];
    if (!q.options) q.options = [];
    q.options.push("");
    // Notify van.js that the state has been mutated
    questions.val = questions.val;
}

		const removeOption = (optIndex) => {
    questions.val[index].options.splice(optIndex, 1);
    // Notify van.js that the state has been mutated
    questions.val = questions.val;
}

		const removeQuestion = () => {
    questions.val.splice(index, 1);
    // Notify van.js that the state has been mutated
    questions.val = questions.val;
};
		
		return div({ class: "bg-white p-4 rounded-lg border border-gray-200 mt-4" },
			div({ class: "flex justify-between items-start gap-4" },
				textarea({
					class: "w-full border-gray-300 rounded-md shadow-sm focus:border-green-x-light focus:ring-green-x-light sm:text-sm text-lg font-medium",
					rows: 2,
					placeholder: `Question ${index + 1}`,
					value: text,
					oninput: e => updateQuestion('text', e.target.value)
				}),
				div({ class: "flex flex-col gap-2 items-end" },
					select({
						class: "rounded-md border-gray-300 shadow-sm focus:border-green-x-light focus:ring-green-x-light sm:text-sm",
						onchange: e => updateQuestion('type', e.target.value)
					},
						option({ value: "text", selected: type === 'text' }, "Short Answer"),
						option({ value: "textarea", selected: type === 'textarea' }, "Paragraph"),
						option({ value: "radio", selected: type === 'radio' }, "Multiple Choice"),
					),
					button({ 
                    type: "button", 
                    onclick: removeQuestion, 
                    class: "p-1 rounded text-gray-500 hover:bg-red-100 hover:text-red-700 transition-colors" 
                }, 
                TrashIcon()
				))
			),
			() => (type === 'radio') ? div({ class: "mt-4 space-y-2 pl-2" },
				(options || []).map((opt, i) => div({ class: "flex items-center gap-2" },
					input({ type: "radio", disabled: true, class: "text-green-x-light" }), 
					input({
						type: "text",
						class: "w-full border-gray-300 rounded-md shadow-sm sm:text-sm",
						placeholder: `Option ${i + 1}`,
						value: opt,
						oninput: e => updateOption(i, e.target.value),
					}),
					button({ type: "button", onclick: () => removeOption(i), class: "text-red-400 hover:text-red-600" }, "✕")
				)),
				button({ type: "button", onclick: addOption, class: "text-sm text-green-x-dark hover:underline mt-2" }, "Add option")
			) : div()
		);
	};
	
	const addQuestion = () => {
		questions.val = [...questions.val, {
			id: `q_${Date.now()}`,
			text: "",
			type: "text",
			options: []
		}];
	};
	
	const SurveyForm = () => form({ method: "POST" },
		hiddenQuestionsInput,
		div({ class: "space-y-6 bg-white p-6 rounded-lg shadow" },
			div(
				label({ for: "name", class: "block text-sm font-medium text-gray-700" }, "Survey Name"),
				input({ type: "text", name: "name", id: "name", required: true, class: "mt-1 block w-full border-gray-300 rounded-md shadow-sm sm:text-sm", value: name, oninput: e => name.val = e.target.value })
			),
			div(
				label({ for: "description", class: "block text-sm font-medium text-gray-700" }, "Description"),
				textarea({ name: "description", id: "description", rows: 3, class: "mt-1 block w-full border-gray-300 rounded-md shadow-sm sm:text-sm", oninput: e => description.val = e.target.value }, description.val)
			),
			div({ class: "flex items-center" },
				input({ type: "checkbox", name: "is_enabled", id: "is_enabled", class: "h-4 w-4 rounded border-gray-300 text-green-x-light", checked: isEnabled, onchange: e => isEnabled.val = e.target.checked }),
				label({ for: "is_enabled", class: "ml-2 block text-sm text-gray-900" }, "Enable this survey for users")
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