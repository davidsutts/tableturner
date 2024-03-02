var fileURL: string;

// Add event listener to generate timetables button.
let genBtn: HTMLButtonElement;
if (document.getElementById("gen-tt")) {
	genBtn = document.getElementById("gen-tt") as HTMLButtonElement;
	genBtn.addEventListener("click", submit);
};

// Get download link element.
let dloadLink: HTMLAnchorElement;
if (document.getElementById("download")) {
	dloadLink = document.getElementById("download") as HTMLAnchorElement;
};

// Request timetable using api.
function submit(): void {
	console.log("HERE")
	let xhr = new XMLHttpRequest();
	xhr.open("POST", "api/generate");

	// Get form data.
	let form = document.getElementById("form") as HTMLFormElement;
	let fd = new FormData(form);

	// Send request.
	xhr.send(fd);

	xhr.addEventListener("loadend", () => {
		if (xhr.status == 200) {
			fileURL = URL.createObjectURL(new Blob([xhr.response], { type: xhr.responseType }));
			dloadLink.classList.remove("hidden")
			dloadLink.href = fileURL;
		}
	});
};