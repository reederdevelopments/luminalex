

const signinBtn = document.getElementById("signin-btn");

if (signinBtn) {
	signinBtn.addEventListener("click", (e) => {
		const popup = window.open("/auth/google", "popup", "popup=yes, width=600, height=700");

		window.addEventListener("message", (e) => {
			if (e.data === "AUTH_SUCCESS") {
				if (popup) {
					popup.close();
				}
				location.reload();
			}
		});
	});
}
