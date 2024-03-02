var fileURL: string;

// Add event listener to generate timetables button.
let genBtn: HTMLButtonElement;
if (document.getElementById("gen-tt")) {
	genBtn = document.getElementById("gen-tt") as HTMLButtonElement;
	genBtn.addEventListener("click", submit);
};

// Add event listener to generate timetables button.
let dloadLink: HTMLAnchorElement;
if (document.getElementById("download")) {
	dloadLink = document.getElementById("download") as HTMLAnchorElement;
	// dloadLink.addEventListener("click", download);
};

// Request timetable using api.
function submit(): void {
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
			dloadLink.href = fileURL;
		}
	});
};