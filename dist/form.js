var fileURL;
// Add event listener to generate timetables button.
var genBtn;
if (document.getElementById("gen-tt")) {
    genBtn = document.getElementById("gen-tt");
    genBtn.addEventListener("click", submit);
}
;
// Get download link element.
var dloadLink;
if (document.getElementById("download")) {
    dloadLink = document.getElementById("download");
}
;
// Request timetable using api.
function submit() {
    console.log("HERE");
    var xhr = new XMLHttpRequest();
    xhr.open("POST", "api/generate");
    // Get form data.
    var form = document.getElementById("form");
    var fd = new FormData(form);
    // Send request.
    xhr.send(fd);
    xhr.addEventListener("loadend", function () {
        if (xhr.status == 200) {
            fileURL = URL.createObjectURL(new Blob([xhr.response], { type: xhr.responseType }));
            dloadLink.classList.remove("hidden");
            dloadLink.href = fileURL;
        }
    });
}
;
//# sourceMappingURL=form.js.map