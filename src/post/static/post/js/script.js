function run() {
    const toggles = document.querySelectorAll(".tog");
    for ( let i = 0; i < toggles.length; i++  ) {
        toggles[i].addEventListener("click", (e) => {
            hideOrShow(e)
        });
    }
}

function hideOrShow(event) {
    console.log(event.target)
}

window.addEventListener("load", run);
