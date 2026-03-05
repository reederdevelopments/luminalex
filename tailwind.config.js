module.exports = {
	content: ["./**/*.templ", "./**/*.js", "!./**/node_modules/**"],
	theme: {
		fontFamily: {
            sans: ['Nexa', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        },
		data: {
			enabled: 'enabled~="true"',
		},
		extend: {
			maxWidth: {
				'8xl': '92rem', 
			},
			colors: {
				hl: "#60C3AD",
				"green-x-light": "#4AB8A1",
				"green-x-dark": "#286256",
				"red-x-light": "#A52022",
				"red-x-dark": "#631114",
			},
		},
	},
	plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};