import van from "https://cdn.jsdelivr.net/gh/vanjs-org/van/public/van-1.5.1.min.js";

const container = document.getElementById("surveys-container");
if (container) {
    const pageData = JSON.parse(container.dataset.surveys);
    const surveys = van.state(pageData.Surveys || []);

    const { div, table, thead, tbody, tr, th, td, a, button, input, span } = van.tags;

    const COOKIE = "MAONI_AUTH";
    let authToken = "";
    document.cookie.split(';').forEach(cookie => {
        const parts = cookie.split('=');
        if (parts[0].trim() === COOKIE) {
            authToken = parts[1];
        }
    });

    async function toggleSurveyStatus(surveyID, checkbox) {
        const isEnabled = checkbox.checked;
        checkbox.disabled = true;

        const response = await fetch(`/api/admin/surveys/${surveyID}/toggle`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Auth': authToken,
            },
            body: JSON.stringify({ isEnabled: isEnabled }),
        });

        if (!response.ok) {
            alert('Failed to update survey status.');
            checkbox.checked = !isEnabled; // Revert checkbox on failure
        }
        checkbox.disabled = false;
    }

    const SurveysTable = () => table({ class: "min-w-full divide-y divide-gray-300" },
        thead(
            tr(
                th({ class: "py-3.5 pl-4 pr-3 text-left text-sm font-semibold text-gray-900 sm:pl-0" }, "Name"),
                th({ class: "px-3 py-3.5 text-left text-sm font-semibold text-gray-900" }, "Description"),
                th({ class: "px-3 py-3.5 text-left text-sm font-semibold text-gray-900" }, "Enabled"),
                th({ class: "px-3 py-3.5 text-left text-sm font-semibold text-gray-900" }, "Responses"),
                th({ class: "relative py-3.5 pl-3 pr-4 sm:pr-0" }, span({ class: "sr-only" }, "Edit")),
            )
        ),
        () => tbody({ class: "divide-y divide-gray-200" },
            surveys.val.length === 0
            ? tr(td({ colspan: 5, class: "text-center text-gray-500 py-8" }, "No surveys created yet."))
            : surveys.val.map(survey => tr(
                td({ class: "whitespace-nowrap py-4 pl-4 pr-3 text-sm font-medium text-gray-900 sm:pl-0" }, survey.Name),
                td({ class: "whitespace-nowrap px-3 py-4 text-sm text-gray-500" }, survey.Description),
                td({ class: "whitespace-nowrap px-3 py-4 text-sm text-gray-500" }, 
                    input({
                        type: "checkbox",
                        checked: survey.IsEnabled,
                        class: "h-4 w-4 rounded border-gray-300 text-green-x-light focus:ring-green-x-light",
                        onchange: (e) => toggleSurveyStatus(survey.ID, e.target),
                    })
                ),
                td({ class: "whitespace-nowrap px-3 py-4 text-sm text-gray-500" }, "0"), // Placeholder
                td({ class: "relative whitespace-nowrap py-4 pl-3 pr-4 text-right text-sm font-medium sm:pr-0" },
                    a({ href: `/admin/surveys/edit/${survey.ID}`, class: "text-green-x-dark hover:text-green-x-darker" }, "Edit"),
                )
            ))
        )
    );

    van.add(container, SurveysTable());
}